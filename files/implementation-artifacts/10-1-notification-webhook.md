# Story 10.1: Notification Rules & Webhook

**Status:** done

## Summary
Implemented notification rules CRUD with webhook delivery, idempotency key support, and REST API under `/api/v1/projects/:id/notification-rules`.

## Deliverables
- `migrations/000016_create_notification_rules.up.sql` - notification_rules table
- `internal/notification/model.go` - NotificationRule, EventType (build_success, build_failed, deploy_success, deploy_failed)
- `internal/notification/dto.go` - Create/Update/List DTOs
- `internal/notification/repo.go` - CRUD, ListByProjectAndEvent
- `internal/notification/service.go` - CRUD, SendWebhook with idempotency key check
- `internal/notification/handler.go` - REST handlers
- `internal/notification/service_test.go` - tests

## Notes
- Webhook uses net/http POST with X-Idempotency-Key header
- Redis cache used for idempotency when available
