package tekton

import (
	"strings"
	"testing"
)

func TestGenerateContainerBuildSteps(t *testing.T) {
	config := ContainerBuildConfig{
		GitURL:         "https://github.com/example/app.git",
		Branch:         "main",
		Commit:         "abc123",
		BuildImage:     "golang:1.21",
		Dockerfile:     "Dockerfile",
		Context:        "/workspace/source",
		RegistryURL:    "harbor.example.com/library",
		ImageName:      "myapp",
		ImageTag:       "v1.0",
		RegistrySecret: "registry-auth",
	}

	g := NewBuildChainGenerator()
	steps := g.GenerateContainerBuildSteps(config)

	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
}

func TestContainerBuildStepNames(t *testing.T) {
	config := ContainerBuildConfig{
		GitURL:      "https://github.com/example/app.git",
		Branch:      "main",
		RegistryURL: "harbor.example.com/library",
		ImageName:   "app",
		ImageTag:    "latest",
	}

	g := NewBuildChainGenerator()
	steps := g.GenerateContainerBuildSteps(config)

	expectedNames := []string{"git-clone", "build", "kaniko-build-push"}
	for i, name := range expectedNames {
		if steps[i].Name != name {
			t.Errorf("step %d: expected name %q, got %q", i, name, steps[i].Name)
		}
	}
}

func TestContainerBuildKanikoArgs(t *testing.T) {
	config := ContainerBuildConfig{
		GitURL:      "https://github.com/example/app.git",
		Branch:      "main",
		Dockerfile:  "build/Dockerfile",
		Context:     "/workspace/source",
		RegistryURL: "harbor.example.com/proj",
		ImageName:   "myapp",
		ImageTag:    "v2.0",
	}

	g := NewBuildChainGenerator()
	steps := g.GenerateContainerBuildSteps(config)

	kaniko := steps[2]
	if kaniko.Name != "kaniko-build-push" {
		t.Fatalf("expected kaniko-build-push step, got %q", kaniko.Name)
	}

	argsStr := strings.Join(kaniko.Args, " ")
	if !strings.Contains(argsStr, "--dockerfile=build/Dockerfile") {
		t.Errorf("Kaniko args missing --dockerfile: %s", argsStr)
	}
	if !strings.Contains(argsStr, "--context=dir:///workspace/source") {
		t.Errorf("Kaniko args missing --context: %s", argsStr)
	}
	if !strings.Contains(argsStr, "--destination=harbor.example.com/proj/myapp:v2.0") {
		t.Errorf("Kaniko args missing --destination: %s", argsStr)
	}
}
