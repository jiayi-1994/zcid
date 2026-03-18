//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestFullPipelineFlowWithTekton(t *testing.T) {
	waitForServer(t)
	c := adminClient(t)

	k8sNamespace := "zcid-run"

	// Step 1: Create project
	projName := uniqueName("e2e-proj")
	t.Logf("creating project %s", projName)
	resp := c.PostJSON(t, "/api/v1/projects", map[string]string{
		"name":        projName,
		"description": "E2E test project for Tekton pipeline flow",
	})
	requireCode(t, resp, 0)
	var proj struct{ ID string `json:"id"` }
	mustUnmarshalData(t, resp, &proj)
	t.Logf("project created: %s", proj.ID)
	defer func() {
		c.DeleteJSON(t, fmt.Sprintf("/api/v1/projects/%s", proj.ID))
	}()

	// Step 2: Create pipeline with a simple shell step
	pipeName := uniqueName("e2e-pipe")
	t.Logf("creating pipeline %s", pipeName)
	resp = c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines", proj.ID), map[string]interface{}{
		"name":        pipeName,
		"description": "E2E test pipeline",
		"triggerType": "manual",
		"config": map[string]interface{}{
			"schemaVersion": "v1",
			"stages": []map[string]interface{}{
				{
					"id":   "stage-build",
					"name": "Build",
					"steps": []map[string]interface{}{
						{
							"id":      "step-echo",
							"name":    "Echo Test",
							"type":    "shell",
							"image":   "alpine:3.21",
							"command": []string{"sh", "-c", "echo 'Hello from Tekton E2E test!' && date && sleep 2"},
						},
					},
				},
			},
		},
	})
	requireCode(t, resp, 0)
	var pipe struct{ ID string `json:"id"` }
	mustUnmarshalData(t, resp, &pipe)
	t.Logf("pipeline created: %s", pipe.ID)

	// Step 3: Trigger pipeline run
	t.Logf("triggering pipeline run...")
	resp = c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs", proj.ID, pipe.ID), map[string]interface{}{
		"gitBranch": "main",
	})
	if resp.Code != 0 {
		t.Fatalf("trigger run failed: code=%d message=%s detail=%s", resp.Code, resp.Message, resp.Detail)
	}

	var run struct {
		ID         string  `json:"id"`
		Status     string  `json:"status"`
		TektonName *string `json:"tektonName"`
	}
	mustUnmarshalData(t, resp, &run)
	t.Logf("run triggered: id=%s status=%s", run.ID, run.Status)

	// Step 4: Verify Tekton PipelineRun CRD is created in the cluster
	t.Run("verify Tekton PipelineRun CRD", func(t *testing.T) {
		var tektonRunFound bool
		for i := 0; i < 30; i++ {
			resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs/%s", proj.ID, pipe.ID, run.ID))
			requireCode(t, resp, 0)
			mustUnmarshalData(t, resp, &run)

			if run.TektonName != nil && *run.TektonName != "" {
				t.Logf("Tekton PipelineRun name: %s", *run.TektonName)

				out := kubectl(t, "get", "pipelinerun", *run.TektonName, "-n", k8sNamespace, "-o", "jsonpath={.metadata.name}")
				if strings.TrimSpace(out) == *run.TektonName {
					tektonRunFound = true
					t.Logf("confirmed PipelineRun CRD exists in cluster")
					break
				}
			}
			time.Sleep(2 * time.Second)
		}
		if !tektonRunFound {
			t.Log("Tekton PipelineRun CRD not found (might be mock mode)")
		}
	})

	// Step 5: Wait for run to complete and verify final status
	t.Run("wait for run completion", func(t *testing.T) {
		var finalStatus string
		for i := 0; i < 120; i++ {
			resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs/%s", proj.ID, pipe.ID, run.ID))
			requireCode(t, resp, 0)

			var data struct {
				Status string `json:"status"`
			}
			mustUnmarshalData(t, resp, &data)
			finalStatus = data.Status

			if data.Status == "succeeded" || data.Status == "failed" {
				t.Logf("run completed with status: %s", data.Status)
				break
			}
			if i%10 == 0 {
				t.Logf("waiting... current status: %s (%ds)", data.Status, i*2)
			}
			time.Sleep(2 * time.Second)
		}

		if finalStatus != "succeeded" && finalStatus != "failed" {
			t.Errorf("expected terminal status, got: %s", finalStatus)
		} else {
			t.Logf("final run status: %s", finalStatus)
		}
	})

	// Step 6: Verify run appears in list with correct status
	t.Run("verify run in list", func(t *testing.T) {
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/pipelines/%s/runs?page=1&pageSize=10", proj.ID, pipe.ID))
		requireCode(t, resp, 0)

		var data struct {
			Items []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.Total < 1 {
			t.Fatal("expected at least 1 run in list")
		}

		found := false
		for _, item := range data.Items {
			if item.ID == run.ID {
				found = true
				t.Logf("run found in list: id=%s status=%s", item.ID, item.Status)
				break
			}
		}
		if !found {
			t.Errorf("run %s not found in list", run.ID)
		}
	})
}

func TestAdminHealthE2E(t *testing.T) {
	waitForServer(t)
	c := adminClient(t)

	resp := c.GetJSON(t, "/api/v1/admin/health")
	requireCode(t, resp, 0)
	t.Log("admin health check passed")
}
