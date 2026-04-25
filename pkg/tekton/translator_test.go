package tekton

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/internal/pipeline"
)

func TestTranslateBasicPipeline(t *testing.T) {
	tr := NewTranslator()
	config := pipeline.PipelineConfig{
		SchemaVersion: "1.0",
		Stages: []pipeline.StageConfig{
			{
				ID:   "s1",
				Name: "build",
				Steps: []pipeline.StepConfig{
					{ID: "step1", Name: "build-step", Type: "build", Image: "golang:1.21", Command: []string{"go", "build"}},
				},
			},
		},
	}

	pr, err := tr.Translate("pipeline-123", "run-456", "proj-789", "default", config, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, pr)

	assert.Equal(t, "tekton.dev/v1", pr.TypeMeta.APIVersion)
	assert.Equal(t, "PipelineRun", pr.TypeMeta.Kind)
	assert.Equal(t, "pipeline-123", pr.Metadata.Labels["zcid.io/pipeline-id"])
	assert.Equal(t, "run-456", pr.Metadata.Labels["zcid.io/run-id"])
	assert.Equal(t, "proj-789", pr.Metadata.Labels["zcid.io/project-id"])

	require.Len(t, pr.Spec.PipelineSpec.Tasks, 1)
	task := pr.Spec.PipelineSpec.Tasks[0]
	assert.Equal(t, "build", task.Name)
	assert.Nil(t, task.RunAfter)
	require.Len(t, task.TaskSpec.Steps, 1)
	step := task.TaskSpec.Steps[0]
	assert.Equal(t, "build-step", step.Name)
	assert.Equal(t, "golang:1.21", step.Image)
	assert.Equal(t, []string{"go", "build"}, step.Command)
}

func TestTranslateWithParams(t *testing.T) {
	tr := NewTranslator()
	config := pipeline.PipelineConfig{
		Stages: []pipeline.StageConfig{
			{ID: "s1", Name: "deploy", Steps: []pipeline.StepConfig{
				{ID: "s1", Name: "deploy", Type: "deploy", Image: "busybox"},
			}},
		},
	}
	params := map[string]string{"ENV": "prod", "VERSION": "1.0.0"}

	pr, err := tr.Translate("p1", "r1", "proj1", "ns1", config, params, nil)
	require.NoError(t, err)

	paramMap := make(map[string]string)
	for _, p := range pr.Spec.Params {
		paramMap[p.Name] = p.Value.StringVal
	}
	assert.Equal(t, "prod", paramMap["ENV"])
	assert.Equal(t, "1.0.0", paramMap["VERSION"])

	// Params also injected as env
	require.Len(t, pr.Spec.PipelineSpec.Tasks[0].TaskSpec.Steps[0].Env, 2)
}

func TestTranslateWithGitInfo(t *testing.T) {
	tr := NewTranslator()
	config := pipeline.PipelineConfig{
		Stages: []pipeline.StageConfig{
			{ID: "s1", Name: "build", Steps: []pipeline.StepConfig{
				{ID: "s1", Name: "build", Type: "build", Image: "alpine"},
			}},
		},
	}
	gitInfo := &GitInfo{
		CommitSHA: "abc123def",
		Branch:    "main",
		Author:    "dev@example.com",
		Message:   "feat: add thing",
	}

	pr, err := tr.Translate("p1", "r1", "proj1", "ns1", config, nil, gitInfo)
	require.NoError(t, err)

	paramMap := make(map[string]string)
	for _, p := range pr.Spec.Params {
		paramMap[p.Name] = p.Value.StringVal
	}
	assert.Equal(t, "abc123def", paramMap["GIT_COMMIT"])
	assert.Equal(t, "main", paramMap["GIT_BRANCH"])
	assert.Equal(t, "dev@example.com", paramMap["GIT_AUTHOR"])
	assert.Equal(t, "feat: add thing", paramMap["GIT_MESSAGE"])
}

func TestTranslateEmptyStages(t *testing.T) {
	tr := NewTranslator()
	config := pipeline.PipelineConfig{
		Stages: []pipeline.StageConfig{},
	}

	pr, err := tr.Translate("p1", "r1", "proj1", "ns1", config, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, pr)
}

func TestTranslateMultiStageOrdering(t *testing.T) {
	tr := NewTranslator()
	config := pipeline.PipelineConfig{
		Stages: []pipeline.StageConfig{
			{ID: "s1", Name: "build", Steps: []pipeline.StepConfig{{ID: "a", Name: "a", Type: "build", Image: "alpine"}}},
			{ID: "s2", Name: "test", Steps: []pipeline.StepConfig{{ID: "b", Name: "b", Type: "test", Image: "alpine"}}},
			{ID: "s3", Name: "deploy", Steps: []pipeline.StepConfig{{ID: "c", Name: "c", Type: "deploy", Image: "alpine"}}},
		},
	}

	pr, err := tr.Translate("p1", "r1", "proj1", "ns1", config, nil, nil)
	require.NoError(t, err)

	// First task: no RunAfter
	assert.Empty(t, pr.Spec.PipelineSpec.Tasks[0].RunAfter)

	// Second task: RunAfter build
	assert.Equal(t, []string{"build"}, pr.Spec.PipelineSpec.Tasks[1].RunAfter)

	// Third task: RunAfter build, test
	assert.Equal(t, []string{"build", "test"}, pr.Spec.PipelineSpec.Tasks[2].RunAfter)
}

func TestTranslateExpandsActionStepTypes(t *testing.T) {
	tr := NewTranslator()
	config := pipeline.PipelineConfig{
		Stages: []pipeline.StageConfig{
			{ID: "checkout", Name: "checkout", Steps: []pipeline.StepConfig{
				{ID: "clone", Name: "clone", Type: "git-clone", Config: map[string]any{"repoUrl": "https://example.com/acme/app.git", "branch": "main", "depth": "1"}},
			}},
			{ID: "build", Name: "build", Steps: []pipeline.StepConfig{
				{ID: "script", Name: "script", Type: "shell", Image: "alpine", Command: []string{"echo preparing", "go test ./..."}},
			}},
			{ID: "docker", Name: "docker", Steps: []pipeline.StepConfig{
				{ID: "image", Name: "image", Type: "kaniko", Config: map[string]any{"imageName": "registry.example.com/app:latest", "dockerfile": "Dockerfile", "context": "."}},
			}},
		},
	}

	pr, err := tr.Translate("p1", "r1", "proj1", "ns1", config, nil, &GitInfo{CommitSHA: "abc123"})
	require.NoError(t, err)

	clone := pr.Spec.PipelineSpec.Tasks[0].TaskSpec.Steps[0]
	assert.Equal(t, GitImage, clone.Image)
	assert.Equal(t, []string{"git", "clone"}, clone.Command)
	assert.Equal(t, []string{"--depth", "1", "--branch", "main", "--single-branch", "https://example.com/acme/app.git", "/workspace/source"}, clone.Args)

	script := pr.Spec.PipelineSpec.Tasks[1].TaskSpec.Steps[0]
	assert.Equal(t, []string{"/bin/sh", "-c"}, script.Command)
	assert.Equal(t, []string{"echo preparing\ngo test ./..."}, script.Args)

	kaniko := pr.Spec.PipelineSpec.Tasks[2].TaskSpec.Steps[0]
	assert.Equal(t, KanikoImage, kaniko.Image)
	assert.Equal(t, []string{"/kaniko/executor"}, kaniko.Command)
	assert.Contains(t, kaniko.Args, "--dockerfile=Dockerfile")
	assert.Contains(t, kaniko.Args, "--context=dir:///workspace/source")
	assert.Contains(t, kaniko.Args, "--destination=registry.example.com/app:latest")
}
