package argocd

import "context"

// ArgoApp represents an ArgoCD Application spec for create/update.
type ArgoApp struct {
	Name           string
	Project        string
	RepoURL        string
	Path           string
	TargetRevision string
	Namespace      string
	Image          string
}

// AppStatus represents the sync and health status of an ArgoCD application.
type AppStatus struct {
	Health     string
	Sync       string
	Resources  []ResourceStatus
}

// ResourceStatus represents a single resource status within an app.
type ResourceStatus struct {
	Kind    string
	Name    string
	Status  string
	Message string
}

// ArgoClient defines the interface for ArgoCD operations.
// Implementations may use the real ArgoCD gRPC API or a mock.
type ArgoClient interface {
	CreateOrUpdateApp(ctx context.Context, app *ArgoApp) error
	SyncApp(ctx context.Context, appName string) error
	GetAppStatus(ctx context.Context, appName string) (*AppStatus, error)
	DeleteApp(ctx context.Context, appName string) error
}

// MockArgoClient implements ArgoClient with in-memory stubs.
// TODO: Replace with real ArgoCD gRPC client (e.g. github.com/argoproj/argo-cd/v2/pkg/apiclient)
// when ArgoCD is configured. Requires: ARGOCD_SERVER, ARGOCD_AUTH_TOKEN or ARGOCD_OPTS.
type MockArgoClient struct{}

var _ ArgoClient = (*MockArgoClient)(nil)

func (m *MockArgoClient) CreateOrUpdateApp(ctx context.Context, app *ArgoApp) error {
	_ = ctx
	_ = app
	// TODO: Implement real ArgoCD Application create/update via gRPC
	return nil
}

func (m *MockArgoClient) SyncApp(ctx context.Context, appName string) error {
	_ = ctx
	_ = appName
	// TODO: Implement real ArgoCD Application Sync via gRPC
	return nil
}

func (m *MockArgoClient) GetAppStatus(ctx context.Context, appName string) (*AppStatus, error) {
	_ = ctx
	// TODO: Implement real ArgoCD Application status via gRPC
	return &AppStatus{
		Health: "Healthy",
		Sync:   "Synced",
		Resources: []ResourceStatus{
			{Kind: "Deployment", Name: appName + "-deploy", Status: "Synced", Message: ""},
		},
	}, nil
}

func (m *MockArgoClient) DeleteApp(ctx context.Context, appName string) error {
	_ = ctx
	_ = appName
	// TODO: Implement real ArgoCD Application delete via gRPC
	return nil
}
