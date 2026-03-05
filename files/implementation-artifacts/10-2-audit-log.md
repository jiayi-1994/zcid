# Story 10.2: Audit Log

**Status:** done

## Summary
Implemented audit log table, async write service, Gin middleware for write operations, and admin-only list API.

## Deliverables
- `migrations/000017_create_audit_logs.up.sql` - audit_logs table
- `internal/audit/model.go` - AuditLog model
- `internal/audit/repo.go` - Create, List (paginated, filterable)
- `internal/audit/service.go` - LogAction (async), List
- `internal/audit/middleware.go` - records POST/PUT/DELETE on success
- `internal/audit/handler.go` - GET /api/v1/admin/audit-logs (admin only)
- `internal/audit/service_test.go` - tests

## Notes
- Audit middleware applied to project scope routes
- LogAction uses goroutine for async write
