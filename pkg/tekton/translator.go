package tekton

import (
	"fmt"
	"strings"

	"github.com/xjy/zcid/internal/pipeline"
)

// GitInfo holds git metadata for a run
type GitInfo struct {
	CommitSHA string
	Branch    string
	Author    string
	Message   string
}

// Translator converts PipelineConfig to Tekton PipelineRun
type Translator struct{}

// NewTranslator creates a new Translator
func NewTranslator() *Translator {
	return &Translator{}
}

// Translate converts PipelineConfig to a Tekton PipelineRun
func (t *Translator) Translate(pipelineID, runID, projectID, namespace string, config pipeline.PipelineConfig, params map[string]string, gitInfo *GitInfo) (*PipelineRun, error) {
	if len(config.Stages) == 0 {
		return nil, fmt.Errorf("pipeline has no stages")
	}

	// Build params slice for Tekton (git info + user params)
	tektonParams := t.buildParams(params, gitInfo)

	// Build tasks from stages
	tasks := make([]PipelineTask, 0, len(config.Stages))
	var prevTaskNames []string

	for i, stage := range config.Stages {
		steps := make([]Step, 0, len(stage.Steps))
		for _, stepCfg := range stage.Steps {
			step, err := t.translateStep(stepCfg, params, gitInfo)
			if err != nil {
				return nil, err
			}
			// Inject params as env vars
			step.Env = append(step.Env, t.buildEnvFromParams(params)...)
			// Add step-level env
			for k, v := range stepCfg.Env {
				step.Env = append(step.Env, EnvVar{Name: k, Value: v})
			}
			steps = append(steps, step)
		}

		taskName := stage.Name
		if taskName == "" {
			taskName = fmt.Sprintf("stage-%d", i+1)
		}

		pt := PipelineTask{
			Name:     taskName,
			TaskSpec: &TaskSpec{Steps: steps},
		}
		if len(prevTaskNames) > 0 {
			pt.RunAfter = make([]string, len(prevTaskNames))
			copy(pt.RunAfter, prevTaskNames)
		}
		tasks = append(tasks, pt)
		prevTaskNames = append(prevTaskNames, taskName)
	}

	safePrefix := func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return s[:n]
	}
	name := fmt.Sprintf("run-%s-%s", safePrefix(pipelineID, 8), safePrefix(runID, 8))
	if len(name) > 63 {
		name = name[:63]
	}

	pr := &PipelineRun{
		TypeMeta: TypeMeta{
			APIVersion: "tekton.dev/v1",
			Kind:       "PipelineRun",
		},
		Metadata: ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"zcid.io/managed-by":  "zcid",
				"zcid.io/pipeline-id": pipelineID,
				"zcid.io/run-id":      runID,
				"zcid.io/project-id":  projectID,
			},
		},
		Spec: PipelineRunSpec{
			PipelineSpec: PipelineSpec{Tasks: tasks},
			Params:       tektonParams,
		},
	}

	return pr, nil
}

func (t *Translator) translateStep(stepCfg pipeline.StepConfig, params map[string]string, gitInfo *GitInfo) (Step, error) {
	switch stepCfg.Type {
	case "git-clone":
		return t.translateGitCloneStep(stepCfg, params, gitInfo)
	case "kaniko", "kaniko-build":
		return t.translateKanikoStep(stepCfg)
	case "buildkit", "buildkit-build":
		return t.translateBuildKitStep(stepCfg)
	case "shell":
		return t.translateShellStep(stepCfg), nil
	default:
		return Step{Name: stepCfg.Name, Image: stepCfg.Image, Command: stepCfg.Command, Args: stepCfg.Args}, nil
	}
}

func (t *Translator) translateGitCloneStep(stepCfg pipeline.StepConfig, params map[string]string, gitInfo *GitInfo) (Step, error) {
	repoURL := configString(stepCfg.Config, "repoUrl", "url", "repository")
	if repoURL == "" {
		return Step{}, fmt.Errorf("git-clone step %q missing repoUrl", stepCfg.Name)
	}

	branch := configString(stepCfg.Config, "branch")
	if branch == "" && gitInfo != nil {
		branch = gitInfo.Branch
	}
	if branch == "" {
		branch = params["GIT_BRANCH"]
	}
	if branch == "" {
		branch = "main"
	}

	args := []string{"--branch", branch, "--single-branch", repoURL, "/workspace/source"}
	if depth := configString(stepCfg.Config, "depth"); depth != "" {
		args = append([]string{"--depth", depth}, args...)
	}

	step := Step{Name: stepCfg.Name, Image: GitImage, Command: []string{"git", "clone"}, Args: args}
	if step.Name == "" {
		step.Name = "git-clone"
	}
	step.Env = append(step.Env, EnvVar{Name: "GIT_BRANCH", Value: branch})
	if gitInfo != nil && gitInfo.CommitSHA != "" {
		step.Env = append(step.Env, EnvVar{Name: "GIT_COMMIT", Value: gitInfo.CommitSHA})
	}
	return step, nil
}

func (t *Translator) translateKanikoStep(stepCfg pipeline.StepConfig) (Step, error) {
	destination := configString(stepCfg.Config, "imageName", "destination", "image")
	if destination == "" {
		return Step{}, fmt.Errorf("kaniko step %q missing imageName", stepCfg.Name)
	}

	dockerfile := configString(stepCfg.Config, "dockerfile")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}
	contextDir := workspacePath(configString(stepCfg.Config, "context"))

	args := []string{"--dockerfile=" + dockerfile, "--context=dir://" + contextDir, "--destination=" + destination}
	if target := configString(stepCfg.Config, "target"); target != "" {
		args = append(args, "--target="+target)
	}
	for _, buildArg := range configLines(stepCfg.Config, "buildArgs") {
		args = append(args, "--build-arg="+buildArg)
	}
	for _, tag := range configLines(stepCfg.Config, "extraTags") {
		args = append(args, "--destination="+tag)
	}

	image := stepCfg.Image
	if image == "" {
		image = KanikoImage
	}
	step := Step{Name: stepCfg.Name, Image: image, Command: []string{"/kaniko/executor"}, Args: args}
	if step.Name == "" {
		step.Name = "kaniko-build-push"
	}
	return step, nil
}

func (t *Translator) translateBuildKitStep(stepCfg pipeline.StepConfig) (Step, error) {
	destination := configString(stepCfg.Config, "imageName", "destination", "image")
	if destination == "" {
		return Step{}, fmt.Errorf("buildkit step %q missing imageName", stepCfg.Name)
	}
	dockerfile := configString(stepCfg.Config, "dockerfile")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}
	contextDir := workspacePath(configString(stepCfg.Config, "context"))

	cmdParts := []string{
		"buildctl-daemonless.sh build",
		"--frontend dockerfile.v0",
		"--local context=" + shellQuote(contextDir),
		"--local dockerfile=" + shellQuote(dockerfileDir(dockerfile)),
		"--opt filename=" + shellQuote(dockerfileName(dockerfile)),
		"--output type=image,name=" + shellQuote(destination) + ",push=true",
	}
	if platform := configString(stepCfg.Config, "platform"); platform != "" {
		cmdParts = append(cmdParts, "--opt platform="+shellQuote(platform))
	}
	for _, buildArg := range configLines(stepCfg.Config, "buildArgs") {
		if k, v, ok := strings.Cut(buildArg, "="); ok && strings.TrimSpace(k) != "" {
			cmdParts = append(cmdParts, "--opt build-arg:"+strings.TrimSpace(k)+"="+shellQuote(strings.TrimSpace(v)))
		}
	}

	image := stepCfg.Image
	if image == "" {
		image = "moby/buildkit:buildx-stable-1"
	}
	step := Step{Name: stepCfg.Name, Image: image, Command: []string{"/bin/sh", "-c"}, Args: []string{strings.Join(cmdParts, " ")}}
	if step.Name == "" {
		step.Name = "buildkit-build-push"
	}
	return step, nil
}

func (t *Translator) translateShellStep(stepCfg pipeline.StepConfig) Step {
	step := Step{Name: stepCfg.Name, Image: stepCfg.Image, Command: stepCfg.Command, Args: stepCfg.Args}
	if step.Image == "" {
		step.Image = "alpine:latest"
	}
	if len(step.Command) > 1 || (len(step.Command) == 1 && strings.ContainsAny(step.Command[0], " \t\n;&|<>")) {
		step.Command = []string{"/bin/sh", "-c"}
		step.Args = []string{strings.Join(stepCfg.Command, "\n")}
	}
	return step
}

func configString(config map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := config[key]; ok {
			switch v := value.(type) {
			case string:
				if s := strings.TrimSpace(v); s != "" {
					return s
				}
			case fmt.Stringer:
				if s := strings.TrimSpace(v.String()); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func configLines(config map[string]any, key string) []string {
	value := configString(config, key)
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '\n' || r == '\r' || r == ',' })
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		if s := strings.TrimSpace(part); s != "" {
			lines = append(lines, s)
		}
	}
	return lines
}

func workspacePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "." {
		return "/workspace/source"
	}
	path = strings.ReplaceAll(path, "\\", "/")
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/workspace/source/" + strings.TrimPrefix(path, "./")
}

func dockerfileDir(path string) string {
	path = strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	if path == "" || !strings.Contains(path, "/") {
		return "/workspace/source"
	}
	idx := strings.LastIndex(path, "/")
	return workspacePath(path[:idx])
}

func dockerfileName(path string) string {
	path = strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	if path == "" {
		return "Dockerfile"
	}
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func (t *Translator) buildParams(params map[string]string, gitInfo *GitInfo) []Param {
	seen := make(map[string]bool)
	result := make([]Param, 0)

	if gitInfo != nil {
		if gitInfo.CommitSHA != "" {
			result = append(result, Param{Name: "GIT_COMMIT", Value: ParamValue{Type: "string", StringVal: gitInfo.CommitSHA}})
			seen["GIT_COMMIT"] = true
		}
		if gitInfo.Branch != "" {
			result = append(result, Param{Name: "GIT_BRANCH", Value: ParamValue{Type: "string", StringVal: gitInfo.Branch}})
			seen["GIT_BRANCH"] = true
		}
		if gitInfo.Author != "" {
			result = append(result, Param{Name: "GIT_AUTHOR", Value: ParamValue{Type: "string", StringVal: gitInfo.Author}})
			seen["GIT_AUTHOR"] = true
		}
		if gitInfo.Message != "" {
			result = append(result, Param{Name: "GIT_MESSAGE", Value: ParamValue{Type: "string", StringVal: gitInfo.Message}})
			seen["GIT_MESSAGE"] = true
		}
	}

	for k, v := range params {
		if seen[k] {
			continue
		}
		result = append(result, Param{Name: k, Value: ParamValue{Type: "string", StringVal: v}})
	}
	return result
}

func (t *Translator) buildEnvFromParams(params map[string]string) []EnvVar {
	env := make([]EnvVar, 0, len(params))
	for k, v := range params {
		env = append(env, EnvVar{Name: k, Value: v})
	}
	return env
}
