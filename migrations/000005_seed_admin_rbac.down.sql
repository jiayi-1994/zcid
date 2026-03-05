-- Rollback admin user RBAC policy
DELETE FROM casbin_rule
WHERE ptype = 'g'
  AND v0 = 'admin-bootstrap-001'
  AND v1 = 'admin';
