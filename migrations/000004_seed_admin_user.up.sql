-- Story 2.2: Legacy bootstrap placeholder.
-- The account is disabled by default. First-admin setup is handled by the
-- one-time bootstrap token flow so fresh installs do not expose a shared
-- credential.
INSERT INTO users (id, username, password_hash, role, status, created_at, updated_at)
VALUES (
    'admin-bootstrap-001',
    'admin',
    '$2a$10$ndEii.uLBDvrbVvV/G6wO.F1N7gX5PXIbMec0KGBYqSIgYcpoX3Qa',
    'admin',
    'disabled',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO NOTHING;
