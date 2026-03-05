package tekton

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/internal/pipeline"
)

func TestSerializeToJSON(t *testing.T) {
	tr := NewTranslator()
	config := pipeline.PipelineConfig{
		Stages: []pipeline.StageConfig{
			{ID: "s1", Name: "build", Steps: []pipeline.StepConfig{{ID: "s1", Name: "build", Type: "build", Image: "alpine"}}},
		},
	}

	pr, err := tr.Translate("pipeline-123", "run-456", "proj-789", "default", config, nil, nil)
	require.NoError(t, err)

	data, err := SerializeToJSON(pr)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "tekton.dev/v1beta1")
	assert.Contains(t, string(data), "PipelineRun")
}
