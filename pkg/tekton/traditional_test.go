package tekton

import (
	"strings"
	"testing"
)

func TestGenerateTraditionalBuildSteps(t *testing.T) {
	config := TraditionalBuildConfig{
		GitURL:        "https://github.com/example/app.git",
		Branch:        "main",
		Commit:        "abc123",
		BuildImage:    "golang:1.21",
		BuildCommand:  "go build -o bin/app .",
		ArtifactPath:  "/workspace/source/bin",
		MinioEndpoint: "minio:9000",
		MinioBucket:   "artifacts",
		MinioSecret:   "minio-credentials",
	}

	g := NewBuildChainGenerator()
	steps, err := g.GenerateTraditionalBuildSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}

	expectedNames := []string{"git-clone", "build", "upload-artifact"}
	for i, name := range expectedNames {
		if steps[i].Name != name {
			t.Errorf("step %d: expected name %q, got %q", i, name, steps[i].Name)
		}
	}
}

func TestTraditionalBuildMinioUpload(t *testing.T) {
	config := TraditionalBuildConfig{
		GitURL:        "https://github.com/example/lib.git",
		Branch:        "develop",
		BuildCommand:  "mvn package",
		ArtifactPath:  "/workspace/source/target",
		MinioEndpoint: "minio.svc:9000",
		MinioBucket:   "build-outputs",
		MinioSecret:   "minio-secret",
	}

	g := NewBuildChainGenerator()
	steps, err := g.GenerateTraditionalBuildSteps(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	upload := steps[2]
	if upload.Name != "upload-artifact" {
		t.Fatalf("expected upload-artifact step, got %q", upload.Name)
	}

	argsStr := strings.Join(upload.Args, " ")
	if !strings.Contains(argsStr, "minio.svc:9000") {
		t.Errorf("upload args missing MinIO endpoint: %s", argsStr)
	}
	if !strings.Contains(argsStr, "build-outputs") {
		t.Errorf("upload args missing bucket: %s", argsStr)
	}
	if !strings.Contains(argsStr, "/workspace/source/target") {
		t.Errorf("upload args missing artifact path: %s", argsStr)
	}

	// Check MinIO secret refs
	foundAccessKey := false
	foundSecretKey := false
	for _, env := range upload.Env {
		if env.Name == "MINIO_ACCESS_KEY" && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
			if env.ValueFrom.SecretKeyRef.Name == "minio-secret" && env.ValueFrom.SecretKeyRef.Key == "accessKey" {
				foundAccessKey = true
			}
		}
		if env.Name == "MINIO_SECRET_KEY" && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
			if env.ValueFrom.SecretKeyRef.Name == "minio-secret" && env.ValueFrom.SecretKeyRef.Key == "secretKey" {
				foundSecretKey = true
			}
		}
	}
	if !foundAccessKey || !foundSecretKey {
		t.Errorf("upload step missing MinIO secret refs: accessKey=%v secretKey=%v", foundAccessKey, foundSecretKey)
	}
}
