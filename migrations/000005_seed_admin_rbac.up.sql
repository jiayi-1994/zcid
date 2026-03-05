-- Seed admin user RBAC policy
-- This assigns the bootstrap admin user to the admin role in Casbin

INSERT INTO casbin_rule (ptype, v0, v1)
VALUES ('g', 'admin-bootstrap-001', 'admin')
ON CONFLICT DO NOTHING;
