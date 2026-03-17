package tekton

import (
	"fmt"
	"regexp"
)

const (
	MinioMCImage = "minio/mc:latest"
)

var unsafeShellChars = regexp.MustCompile("[;|`$&><]")

// TraditionalBuildConfig holds configuration for non-containerized (traditional) build chains
type TraditionalBuildConfig struct {
	GitURL        string
	Branch        string
	Commit        string
	BuildImage    string
	BuildCommand  string
	ArtifactPath  string
	MinioEndpoint string
	MinioBucket   string
	MinioSecret   string
}

func validateShellSafe(name, value string) error {
	if unsafeShellChars.MatchString(value) {
		return fmt.Errorf("%s contains unsafe shell characters", name)
	}
	return nil
}

// GenerateTraditionalBuildSteps generates steps for traditional builds: git-clone -> build -> upload-artifact.
// Returns error if any input contains unsafe shell characters.
func (g *BuildChainGenerator) GenerateTraditionalBuildSteps(config TraditionalBuildConfig) ([]Step, error) {
	if err := validateShellSafe("BuildCommand", config.BuildCommand); err != nil {
		return nil, err
	}
	if err := validateShellSafe("ArtifactPath", config.ArtifactPath); err != nil {
		return nil, err
	}
	if err := validateShellSafe("MinioBucket", config.MinioBucket); err != nil {
		return nil, err
	}

	steps := make([]Step, 0, 3)

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

	// 2. build - Compile/package
	buildImg := config.BuildImage
	if buildImg == "" {
		buildImg = "alpine:latest"
	}
	buildCmd := config.BuildCommand
	if buildCmd == "" {
		buildCmd = "echo 'Build step - no command specified'"
	}
	build := Step{
		Name:  "build",
		Image: buildImg,
		Command: []string{
			"/bin/sh", "-c",
		},
		Args: []string{
			"cd /workspace/source && " + buildCmd,
		},
		Env: []EnvVar{
			{Name: "GIT_COMMIT", Value: config.Commit},
			{Name: "GIT_BRANCH", Value: config.Branch},
		},
	}
	steps = append(steps, build)

	// 3. upload-artifact - Upload to MinIO using minio/mc image
	artifactPath := config.ArtifactPath
	if artifactPath == "" {
		artifactPath = "/workspace/source"
	}

	upload := Step{
		Name:  "upload-artifact",
		Image: MinioMCImage,
		Command: []string{
			"/bin/sh", "-c",
		},
		Args: []string{
			"mc alias set minio http://" + config.MinioEndpoint + " $MINIO_ACCESS_KEY $MINIO_SECRET_KEY && mc cp -r " + artifactPath + " minio/" + config.MinioBucket + "/",
		},
		Env: []EnvVar{},
	}
	if config.MinioSecret != "" {
		upload.Env = append(upload.Env,
			EnvVar{
				Name: "MINIO_ACCESS_KEY",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeyRef{
						Name: config.MinioSecret,
						Key:  "accessKey",
					},
				},
			},
			EnvVar{
				Name: "MINIO_SECRET_KEY",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeyRef{
						Name: config.MinioSecret,
						Key:  "secretKey",
					},
				},
			},
		)
	}
	steps = append(steps, upload)

	return steps, nil
}
