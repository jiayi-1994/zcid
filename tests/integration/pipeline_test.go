//go:build integration

package integration

import (
	"fmt"
	"testing"
)

func TestPipelineCRUD(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	projectName := uniqueName("pipe-proj")
	resp := c.PostJSON(t, "/api/v1/projects", map[string]string{
		"name":        projectName,
		"description": "Project for pipeline tests",
	})
	requireCode(t, resp, 0)
	var proj struct {
		ID string `json:"id"`
	}
	mustUnmarshalData(t, resp, &proj)
	projectID := proj.ID
	defer c.DeleteJSON(t, fmt.Sprintf("/api/v1/projects/%s", projectID))

	pipelineName := uniqueName("test-pipe")
	var pipelineID string

	t.Run("create pipeline", func(t *testing.T) {
		resp := c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines", projectID), map[string]interface{}{
			"name":        pipelineName,
			"description": "Integration test pipeline",
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
								"command": []string{"echo", "hello world"},
							},
						},
					},
				},
			},
		})
		requireCode(t, resp, 0)

		var data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.ID == "" {
			t.Fatal("pipeline ID should not be empty")
		}
		pipelineID = data.ID
	})

	t.Run("get pipeline", func(t *testing.T) {
		if pipelineID == "" {
			t.Skip("no pipeline created")
		}
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s", projectID, pipelineID))
		requireCode(t, resp, 0)

		var data struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Config struct {
				Stages []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"stages"`
			} `json:"config"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.Name != pipelineName {
			t.Errorf("expected name %q, got %q", pipelineName, data.Name)
		}
		if len(data.Config.Stages) != 1 {
			t.Errorf("expected 1 stage, got %d", len(data.Config.Stages))
		}
	})

	t.Run("list pipelines", func(t *testing.T) {
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines", projectID))
		requireCode(t, resp, 0)

		var data struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.Total < 1 {
			t.Error("expected at least 1 pipeline")
		}
	})

	t.Run("update pipeline", func(t *testing.T) {
		if pipelineID == "" {
			t.Skip("no pipeline created")
		}
		updatedDesc := "Updated description"
		resp := c.PutJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s", projectID, pipelineID), map[string]interface{}{
			"description": &updatedDesc,
		})
		requireCode(t, resp, 0)
	})

	t.Run("copy pipeline", func(t *testing.T) {
		if pipelineID == "" {
			t.Skip("no pipeline created")
		}
		resp := c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/copy", projectID, pipelineID), nil)
		requireCode(t, resp, 0)

		var data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.ID == "" {
			t.Error("copied pipeline ID should not be empty")
		}
		if data.ID == pipelineID {
			t.Error("copied pipeline should have a different ID")
		}
	})

	t.Run("delete pipeline", func(t *testing.T) {
		if pipelineID == "" {
			t.Skip("no pipeline created")
		}
		resp := c.DeleteJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s", projectID, pipelineID))
		requireCode(t, resp, 0)
	})
}

func TestPipelineTemplates(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	t.Run("list templates", func(t *testing.T) {
		resp := c.GetJSON(t, "/api/v1/pipeline-templates")
		requireCode(t, resp, 0)

		var data struct {
			Items []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"items"`
			Total int `json:"total"`
		}
		mustUnmarshalData(t, resp, &data)
		if len(data.Items) == 0 {
			t.Error("expected at least 1 built-in template")
		}
		t.Logf("found %d pipeline templates", len(data.Items))
	})
}
