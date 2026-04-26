DROP INDEX IF EXISTS idx_notif_rules_channel;

UPDATE notification_rules SET webhook_url = '' WHERE webhook_url IS NULL;

ALTER TABLE notification_rules
  ALTER COLUMN webhook_url SET NOT NULL;

ALTER TABLE notification_rules
  DROP COLUMN IF EXISTS slack_channel,
  DROP COLUMN IF EXISTS slack_token,
  DROP COLUMN IF EXISTS channel_type;
