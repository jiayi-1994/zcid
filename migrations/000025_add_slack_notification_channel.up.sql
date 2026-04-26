ALTER TABLE notification_rules
  ADD COLUMN IF NOT EXISTS channel_type VARCHAR(20) NOT NULL DEFAULT 'webhook',
  ADD COLUMN IF NOT EXISTS slack_token TEXT,
  ADD COLUMN IF NOT EXISTS slack_channel VARCHAR(120);

ALTER TABLE notification_rules
  ALTER COLUMN webhook_url DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_notif_rules_channel ON notification_rules(channel_type);
