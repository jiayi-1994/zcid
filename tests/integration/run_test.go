//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"
)

func TestPipelineRunMockMode(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	projectName := uniqueName("run-proj")
	resp := c.PostJSON(t, "/api/v1/projects", map[string]string{
		"name": projectName,
	})
	requireCode(t, resp, 0)
	var proj struct{ ID string `json:"id"` }
	mustUnmarshalData(t, resp, &proj)
	defer c.DeleteJSON(t, fmt.Sprintf("/api/v1/projects/%s", proj.ID))

	resp = c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines", proj.ID), map[string]interface{}{
		"name":        uniqueName("run-pipe"),
		"triggerType": "manual",
		"config": map[string]interface{}{
			"schemaVersion": "v1",
			"stages": []map[string]interface{}{
				{
					"id":   "stage-1",
					"name": "Build",
					"steps": []map[string]interface{}{
						{
							"id":      "step-1",
							"name":    "echo",
							"type":    "shell",
							"image":   "alpine:latest",
							"command": []string{"echo", "hello"},
						},
					},
				},
			},
		},
	})
	requireCode(t, resp, 0)
	var pipe struct{ ID string `json:"id"` }
	mustUnmarshalData(t, resp, &pipe)

	var runID string

	t.Run("trigger run", func(t *testing.T) {
		resp := c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs", proj.ID, pipe.ID), map[string]interface{}{
			"gitBranch": "main",
		})
		requireCode(t, resp, 0)

		var data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.ID == "" {
			t.Fatal("run ID should not be empty")
		}
		runID = data.ID
		t.Logf("triggered run %s with status %s", data.ID, data.Status)
	})

	t.Run("get run detail", func(t *testing.T) {
		if runID == "" {
			t.Skip("no run triggered")
		}
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs/%s", proj.ID, pipe.ID, runID))
		requireCode(t, resp, 0)

		var data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.ID != runID {
			t.Errorf("expected run ID %q, got %q", runID, data.ID)
		}
	})

	t.Run("list runs", func(t *testing.T) {
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs", proj.ID, pipe.ID))
		requireCode(t, resp, 0)

		var data struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.Total < 1 {
			t.Error("expected at least 1 run")
		}
	})

	t.Run("wait for mock run completion", func(t *testing.T) {
		if runID == "" {
			t.Skip("no run triggered")
		}

		var finalStatus string
		for i := 0; i < 30; i++ {
			resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs/%s", proj.ID, pipe.ID, runID))
			requireCode(t, resp, 0)

			var data struct {
				Status string `json:"status"`
			}
			mustUnmarshalData(t, resp, &data)
			finalStatus = data.Status

			if data.Status == "succeeded" || data.Status == "failed" {
				break
			}
			time.Sleep(time.Second)
		}

		t.Logf("final run status: %s", finalStatus)
		if finalStatus != "succeeded" && finalStatus != "failed" {
			t.Errorf("expected terminal status, got %q (mock runs should complete within ~10s)", finalStatus)
		}
	})
}
