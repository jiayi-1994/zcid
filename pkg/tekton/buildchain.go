package tekton

const (
	// KanikoImage is the standard Kaniko executor for building Docker images without Docker daemon
	KanikoImage = "gcr.io/kaniko-project/executor:v1.22.0"
	// GitImage is the standard git clone image
	GitImage = "alpine/git:latest"
)

// Retry annotations for push step (max 3 retries)
const (
	RetryAnnotationKey   = "tekton.dev/retries"
	RetryAnnotationValue = "3"
)

// ContainerBuildConfig holds configuration for containerized build chains
type ContainerBuildConfig struct {
	GitURL         string
	Branch         string
	Commit         string
	BuildImage     string
	Dockerfile     string
	Context        string
	RegistryURL    string
	ImageName      string
	ImageTag       string
	RegistrySecret string
}

// BuildChainGenerator generates Tekton steps for build chains
type BuildChainGenerator struct{}

// NewBuildChainGenerator creates a new BuildChainGenerator
func NewBuildChainGenerator() *BuildChainGenerator {
	return &BuildChainGenerator{}
}

// GenerateContainerBuildSteps generates steps for containerized builds: git-clone -> build -> kaniko-build-push
func (g *BuildChainGenerator) GenerateContainerBuildSteps(config ContainerBuildConfig) []Step {
	steps := make([]Step, 0, 3)

	// 1. git-clone - Clone repo using git image
	gitClone := Step{
		Name:  "git-clone",
		Image: GitImage,
		Command: []string{
			"git", "clone",
		},
		Args: []string{
			"--branch", config.Branch,
			"--single-branch",
			config.GitURL,
			"/workspace/source",
		},
		Env: []EnvVar{
			{Name: "GIT_BRANCH", Value: config.Branch},
			{Name: "GIT_COMMIT", Value: config.Commit},
		},
	}
	if config.Commit != "" {
		gitClone.Args = append([]string{"--depth", "1"}, gitClone.Args...)
	}
	steps = append(steps, gitClone)

	// 2. build - Compile using specified build image (optional; can be a no-op if Kaniko builds from source)
	buildImg := config.BuildImage
	if buildImg == "" {
		buildImg = "alpine:latest"
	}
	build := Step{
		Name:  "build",
		Image: buildImg,
		Command: []string{
			"/bin/sh", "-c",
		},
		Args: []string{
			"cd /workspace/source && echo 'Build step - override with actual build command'",
		},
		Env: []EnvVar{
			{Name: "GIT_COMMIT", Value: config.Commit},
			{Name: "GIT_BRANCH", Value: config.Branch},
		},
	}
	steps = append(steps, build)

	// 3. kaniko-build-push - Build+push Docker image using Kaniko
	dockerfile := config.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}
	context := config.Context
	if context == "" {
		context = "/workspace/source"
	}
	destination := config.RegistryURL + "/" + config.ImageName + ":" + config.ImageTag

	kanikoArgs := []string{
		"--dockerfile=" + dockerfile,
		"--context=dir://" + context,
		"--destination=" + destination,
	}

	kaniko := Step{
		Name:    "kaniko-build-push",
		Image:   KanikoImage,
		Command: []string{"/kaniko/executor"},
		Args:    kanikoArgs,
		Env:     []EnvVar{},
	}
	if config.RegistrySecret != "" {
		kaniko.Env = append(kaniko.Env, EnvVar{
			Name: "DOCKER_CONFIG",
			ValueFrom: &EnvVarSource{
				SecretKeyRef: &SecretKeyRef{
					Name: config.RegistrySecret,
					Key:  ".dockerconfigjson",
				},
			},
		})
	}
	// Add retry logic for push as annotations would be on the Task/Step - we encode intent via a special env
	// Tekton step-level retries are configured in the Task spec; we add a hint env for documentation
	kaniko.Env = append(kaniko.Env, EnvVar{Name: "KANIKO_RETRIES", Value: "3"})

	steps = append(steps, kaniko)

	return steps
}
