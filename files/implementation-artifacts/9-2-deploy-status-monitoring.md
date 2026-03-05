# Story 9.2: Deploy Status Monitoring

**Status:** done

## Summary
Added RefreshStatus and ResyncDeploy to deployment service; POST /:deployId/resync handler.

## Deliverables
- `internal/deployment/service.go` - RefreshStatus, ResyncDeploy
- `internal/deployment/handler.go` - POST /:deployId/resync

## Notes
- RefreshStatus fetches fresh status from ArgoCD mock
- ResyncDeploy triggers ArgoCD Application sync
