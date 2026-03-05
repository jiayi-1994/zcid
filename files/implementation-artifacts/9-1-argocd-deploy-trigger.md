# Story 9.1: ArgoCD Integration & Deploy Trigger

**Status:** done

## Summary
Implemented ArgoCD client interface with mock, deployments table, deployment model/repo/service/handler, and deploy trigger API.

## Deliverables
- `pkg/argocd/client.go` - ArgoClient interface, ArgoApp, AppStatus, ResourceStatus, MockArgoClient (TODO for real gRPC)
- `pkg/argocd/client_test.go` - TestMockArgoClient
- `migrations/000015_create_deployments.up.sql` - deployments table
- `internal/deployment/model.go` - Deployment GORM model, DeployStatus constants
- `internal/deployment/dto.go` - TriggerDeployRequest, DeploymentResponse, DeploymentSummary, DeploymentListResponse
- `internal/deployment/repo.go` - CRUD, ListByProject, ListByEnvironment
- `internal/deployment/service.go` - TriggerDeploy, GetDeployStatus, ListDeployments, GetDeployment
- `internal/deployment/handler.go` - POST/GET deployments routes
- `internal/deployment/service_test.go` - TestTriggerDeploy, TestGetDeployStatus, TestListDeployments

## Notes
- ArgoCD operations use mock client; replace with real gRPC when ArgoCD is configured
- Deployment status synced from ArgoCD on GetStatus/GetDeployment
