package argocd

import (
	"context"
	"testing"
)

func TestMockArgoClient_CreateOrUpdateApp(t *testing.T) {
	client := &MockArgoClient{}
	app := &ArgoApp{
		Name:           "test-app",
		Project:        "default",
		RepoURL:        "https://github.com/example/repo",
		Path:           "manifests",
		TargetRevision: "HEAD",
		Namespace:      "default",
		Image:          "nginx:latest",
	}
	err := client.CreateOrUpdateApp(context.Background(), app)
	if err != nil {
		t.Fatalf("CreateOrUpdateApp: %v", err)
	}
}

func TestMockArgoClient_SyncApp(t *testing.T) {
	client := &MockArgoClient{}
	err := client.SyncApp(context.Background(), "test-app")
	if err != nil {
		t.Fatalf("SyncApp: %v", err)
	}
}

func TestMockArgoClient_GetAppStatus(t *testing.T) {
	client := &MockArgoClient{}
	status, err := client.GetAppStatus(context.Background(), "test-app")
	if err != nil {
		t.Fatalf("GetAppStatus: %v", err)
	}
	if status.Health != "Healthy" || status.Sync != "Synced" {
		t.Errorf("unexpected status: Health=%q Sync=%q", status.Health, status.Sync)
	}
	if len(status.Resources) == 0 {
		t.Error("expected at least one resource")
	}
}

func TestMockArgoClient_DeleteApp(t *testing.T) {
	client := &MockArgoClient{}
	err := client.DeleteApp(context.Background(), "test-app")
	if err != nil {
		t.Fatalf("DeleteApp: %v", err)
	}
}
