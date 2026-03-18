//go:build integration

package integration

import (
	"fmt"
	"testing"
)

func TestProjectCRUD(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)
	projectName := uniqueName("test-proj")

	var projectID string

	t.Run("create project", func(t *testing.T) {
		resp := c.PostJSON(t, "/api/v1/projects", map[string]string{
			"name":        projectName,
			"description": "Integration test project",
		})
		requireCode(t, resp, 0)

		var data struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.ID == "" {
			t.Fatal("project ID should not be empty")
		}
		if data.Name != projectName {
			t.Errorf("expected name %q, got %q", projectName, data.Name)
		}
		if data.Status != "active" {
			t.Errorf("expected status 'active', got %q", data.Status)
		}
		projectID = data.ID
	})

	t.Run("get project", func(t *testing.T) {
		if projectID == "" {
			t.Skip("no project created")
		}
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s", projectID))
		requireCode(t, resp, 0)

		var data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.ID != projectID {
			t.Errorf("expected id %q, got %q", projectID, data.ID)
		}
	})

	t.Run("list projects", func(t *testing.T) {
		resp := c.GetJSON(t, "/api/v1/projects?page=1&pageSize=50")
		requireCode(t, resp, 0)

		var data struct {
			Items []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.Total < 1 {
			t.Error("expected at least 1 project in the list")
		}
		found := false
		for _, item := range data.Items {
			if item.ID == projectID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created project %s not found in list", projectID)
		}
	})

	t.Run("update project", func(t *testing.T) {
		if projectID == "" {
			t.Skip("no project created")
		}
		updatedName := projectName + "-updated"
		resp := c.PutJSON(t, fmt.Sprintf("/api/v1/projects/%s", projectID), map[string]interface{}{
			"name": updatedName,
		})
		requireCode(t, resp, 0)

		var data struct {
			Name string `json:"name"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.Name != updatedName {
			t.Errorf("expected updated name %q, got %q", updatedName, data.Name)
		}
	})

	t.Run("delete project", func(t *testing.T) {
		if projectID == "" {
			t.Skip("no project created")
		}
		resp := c.DeleteJSON(t, fmt.Sprintf("/api/v1/projects/%s", projectID))
		requireCode(t, resp, 0)
	})

	t.Run("get deleted project returns 404", func(t *testing.T) {
		if projectID == "" {
			t.Skip("no project created")
		}
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s", projectID))
		if resp.Code == 0 {
			t.Error("expected non-zero code when getting deleted project")
		}
	})
}

func TestProjectValidation(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	t.Run("create project without name", func(t *testing.T) {
		resp := c.PostJSON(t, "/api/v1/projects", map[string]string{
			"description": "Missing name",
		})
		if resp.Code == 0 {
			t.Error("expected validation error for missing name")
		}
	})
}
