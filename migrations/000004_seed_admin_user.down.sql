-- Rollback: Remove bootstrap admin user
DELETE FROM users WHERE id = 'admin-bootstrap-001';
