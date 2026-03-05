-- Story 2.2: Seed initial admin user for bootstrapping
-- Username: admin
-- Password: admin123 (bcrypt hashed with cost 10)
INSERT INTO users (id, username, password_hash, role, status, created_at, updated_at)
VALUES (
    'admin-bootstrap-001',
    'admin',
    '$2a$10$ndEii.uLBDvrbVvV/G6wO.F1N7gX5PXIbMec0KGBYqSIgYcpoX3Qa',
    'admin',
    'active',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO NOTHING;
