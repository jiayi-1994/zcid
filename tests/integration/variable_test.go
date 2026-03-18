//go:build integration

package integration

import (
	"fmt"
	"testing"
)

func TestVariableCRUD(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	projectName := uniqueName("var-proj")
	resp := c.PostJSON(t, "/api/v1/projects", map[string]string{
		"name": projectName,
	})
	requireCode(t, resp, 0)
	var proj struct{ ID string `json:"id"` }
	mustUnmarshalData(t, resp, &proj)
	defer c.DeleteJSON(t, fmt.Sprintf("/api/v1/projects/%s", proj.ID))

	var varID string

	t.Run("create variable", func(t *testing.T) {
		resp := c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/variables", proj.ID), map[string]interface{}{
			"key":      "TEST_VAR",
			"value":    "test-value-123",
			"isSecret": false,
			"scope":    "project",
		})
		requireCode(t, resp, 0)

		var data struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.ID == "" {
			t.Fatal("variable ID should not be empty")
		}
		varID = data.ID
	})

	t.Run("create secret variable", func(t *testing.T) {
		resp := c.PostJSON(t, fmt.Sprintf("/api/v1/projects/%s/variables", proj.ID), map[string]interface{}{
			"key":      "SECRET_KEY",
			"value":    "super-secret-value",
			"isSecret": true,
			"scope":    "project",
		})
		requireCode(t, resp, 0)
	})

	t.Run("list variables", func(t *testing.T) {
		resp := c.GetJSON(t, fmt.Sprintf("/api/v1/projects/%s/variables", proj.ID))
		requireCode(t, resp, 0)

		var data struct {
			Items []struct {
				ID       string `json:"id"`
				Key      string `json:"key"`
				IsSecret bool   `json:"isSecret"`
			} `json:"items"`
		}
		mustUnmarshalData(t, resp, &data)
		if len(data.Items) < 2 {
			t.Errorf("expected at least 2 variables, got %d", len(data.Items))
		}
	})

	t.Run("update variable", func(t *testing.T) {
		if varID == "" {
			t.Skip("no variable created")
		}
		resp := c.PutJSON(t, fmt.Sprintf("/api/v1/projects/%s/variables/%s", proj.ID, varID), map[string]interface{}{
			"value": "updated-value",
		})
		requireCode(t, resp, 0)
	})

	t.Run("delete variable", func(t *testing.T) {
		if varID == "" {
			t.Skip("no variable created")
		}
		resp := c.DeleteJSON(t, fmt.Sprintf("/api/v1/projects/%s/variables/%s", proj.ID, varID))
		requireCode(t, resp, 0)
	})
}
