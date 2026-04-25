package tekton

import (
	"fmt"

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
			step := Step{
				Name:    stepCfg.Name,
				Image:   stepCfg.Image,
				Command: stepCfg.Command,
				Args:    stepCfg.Args,
			}
			// Inject params as env vars
			step.Env = t.buildEnvFromParams(params)
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
