package tekton

import (
	"encoding/json"
)

// SerializeToJSON serializes PipelineRun to JSON (K8s accepts JSON).
// For true YAML output, consider using sigs.k8s.io/yaml in the future.
func SerializeToJSON(pr *PipelineRun) ([]byte, error) {
	return json.MarshalIndent(pr, "", "  ")
}
