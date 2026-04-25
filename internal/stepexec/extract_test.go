package stepexec

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestExtractTaskRunClassifiesEnvSources(t *testing.T) {
	taskRun := map[string]any{
		"metadata": map[string]any{
			"name":   "run-build",
			"labels": map[string]any{"zcid.io/run-id": "run-1", "zcid.io/managed-by": "zcid"},
		},
		"status": map[string]any{
			"taskSpec": map[string]any{"steps": []any{map[string]any{
				"name":    "build",
				"image":   "alpine:3.20",
				"command": []any{"sh", "-c"},
				"args":    []any{"make build"},
				"env": []any{
					map[string]any{"name": "PUBLIC_FLAG", "value": "true"},
					map[string]any{"name": "TOKEN", "valueFrom": map[string]any{"secretKeyRef": map[string]any{"name": "github-token", "key": "token"}}},
					map[string]any{"name": "POD_IP", "valueFrom": map[string]any{"fieldRef": map[string]any{"fieldPath": "status.podIP"}}},
				},
			}}},
			"steps": []any{map[string]any{
				"name":    "build",
				"imageID": "docker-pullable://alpine@sha256:abc123",
				"terminated": map[string]any{
					"exitCode":   float64(0),
					"reason":     "Completed",
					"startedAt":  "2026-04-24T00:00:00Z",
					"finishedAt": "2026-04-24T00:00:10Z",
				},
			}},
		},
	}
	obj := &unstructured.Unstructured{Object: taskRun}
	rows, err := ExtractTaskRun(obj)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	row := rows[0]
	require.Equal(t, "run-1", row.PipelineRunID)
	require.Equal(t, StatusSucceeded, row.Status)
	require.NotNil(t, row.DurationMS)
	require.Equal(t, int64(10_000), *row.DurationMS)
	require.NotNil(t, row.ImageDigest)
	require.Equal(t, "sha256:abc123", *row.ImageDigest)

	var env map[string]any
	require.NoError(t, json.Unmarshal(row.EnvPublic.Bytes(), &env))
	require.Equal(t, map[string]any{"captured": false, "source": "literal"}, env["PUBLIC_FLAG"])
	require.NotContains(t, string(row.EnvPublic.Bytes()), "true")
	require.NotContains(t, env, "TOKEN")

	var refs []map[string]any
	require.NoError(t, json.Unmarshal(row.SecretRefs.Bytes(), &refs))
	require.Len(t, refs, 2)
	require.Equal(t, "secretKeyRef", refs[0]["source"])
	require.Equal(t, "fieldRef", refs[1]["source"])
}

func TestApplySizeLimitsMarksTruncatedJSON(t *testing.T) {
	row := &StepExecution{CommandArgs: JSONRaw([]byte(`{"script":"` + strings.Repeat("a", CommandArgsLimitKB+10) + `"}`))}
	ApplySizeLimits(row)
	var got map[string]any
	require.NoError(t, json.Unmarshal(row.CommandArgs.Bytes(), &got))
	require.Equal(t, true, got["_truncated"])
	require.NotContains(t, got, "raw_prefix")
	require.NotContains(t, string(row.CommandArgs.Bytes()), "aaa")
}
