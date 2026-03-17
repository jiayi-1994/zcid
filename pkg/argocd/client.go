package argocd

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type ArgoApp struct {
	Name           string
	Project        string
	RepoURL        string
	Path           string
	TargetRevision string
	Namespace      string
	Image          string
}

type AppStatus struct {
	Health    string
	Sync      string
	Resources []ResourceStatus
}

type ResourceStatus struct {
	Kind    string
	Name    string
	Status  string
	Message string
}

type ArgoClient interface {
	CreateOrUpdateApp(ctx context.Context, app *ArgoApp) error
	SyncApp(ctx context.Context, appName string) error
	GetAppStatus(ctx context.Context, appName string) (*AppStatus, error)
	DeleteApp(ctx context.Context, appName string) error
}

type mockApp struct {
	health    string
	sync      string
	createdAt time.Time
}

type MockArgoClient struct {
	mu   sync.Mutex
	apps map[string]*mockApp
}

var _ ArgoClient = (*MockArgoClient)(nil)

func (m *MockArgoClient) init() {
	if m.apps == nil {
		m.apps = make(map[string]*mockApp)
	}
}

func (m *MockArgoClient) CreateOrUpdateApp(ctx context.Context, app *ArgoApp) error {
	m.mu.Lock()
	m.init()
	m.apps[app.Name] = &mockApp{health: "Progressing", sync: "OutOfSync", createdAt: time.Now()}
	m.mu.Unlock()
	slog.Info("MOCK ArgoCD: Created/Updated app", "name", app.Name, "image", app.Image, "namespace", app.Namespace)
	return nil
}

func (m *MockArgoClient) SyncApp(ctx context.Context, appName string) error {
	m.mu.Lock()
	m.init()
	if app, ok := m.apps[appName]; ok {
		app.sync = "Syncing"
		app.createdAt = time.Now()
	}
	m.mu.Unlock()
	slog.Info("MOCK ArgoCD: Sync triggered", "name", appName)

	go m.simulateSync(appName)
	return nil
}

func (m *MockArgoClient) simulateSync(appName string) {
	time.Sleep(3 * time.Second)
	m.mu.Lock()
	if app, ok := m.apps[appName]; ok {
		app.sync = "Synced"
		app.health = "Progressing"
	}
	m.mu.Unlock()

	duration := 5 + rand.Intn(8)
	time.Sleep(time.Duration(duration) * time.Second)

	m.mu.Lock()
	if app, ok := m.apps[appName]; ok {
		if rand.Intn(10) < 1 {
			app.health = "Degraded"
		} else {
			app.health = "Healthy"
		}
	}
	m.mu.Unlock()
	slog.Info("MOCK ArgoCD: Sync completed", "name", appName)
}

func (m *MockArgoClient) GetAppStatus(ctx context.Context, appName string) (*AppStatus, error) {
	m.mu.Lock()
	m.init()
	app, ok := m.apps[appName]
	m.mu.Unlock()

	if !ok {
		return &AppStatus{
			Health: "Healthy",
			Sync:   "Synced",
			Resources: []ResourceStatus{
				{Kind: "Deployment", Name: appName + "-deploy", Status: "Synced"},
			},
		}, nil
	}

	return &AppStatus{
		Health: app.health,
		Sync:   app.sync,
		Resources: []ResourceStatus{
			{Kind: "Deployment", Name: appName + "-deploy", Status: app.sync},
			{Kind: "Service", Name: appName + "-svc", Status: app.sync},
		},
	}, nil
}

func (m *MockArgoClient) DeleteApp(ctx context.Context, appName string) error {
	m.mu.Lock()
	m.init()
	delete(m.apps, appName)
	m.mu.Unlock()
	slog.Info("MOCK ArgoCD: Deleted app", "name", appName)
	return nil
}
