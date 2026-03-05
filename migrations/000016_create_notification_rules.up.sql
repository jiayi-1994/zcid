CREATE TABLE notification_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    name VARCHAR(200) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    webhook_url TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(255) NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_notif_rules_project ON notification_rules(project_id);
CREATE INDEX idx_notif_rules_event ON notification_rules(event_type);
