# Story 9.3: Deploy History & Rollback

**Status:** done

## Summary
Added RollbackDeploy and GetDeployHistory to service; rollback and deploy-history handlers; frontend deployment pages and routes.

## Deliverables
- `internal/deployment/service.go` - RollbackDeploy, GetDeployHistory
- `internal/deployment/handler.go` - POST /:deployId/rollback, GET /environments/:envId/deploy-history
- `web/src/services/deployment.ts` - API service
- `web/src/pages/projects/deployments/DeploymentListPage.tsx` - deployment list
- `web/src/pages/projects/deployments/DeploymentDetailPage.tsx` - detail + status
- `web/src/App.tsx` - routes for deployments
- `web/src/pages/projects/ProjectLayout.tsx` - "部署" menu item

## Notes
- Rollback deploys previous healthy image to same environment
- Deploy history paginated per environment
