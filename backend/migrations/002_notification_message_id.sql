ALTER TABLE notifications ADD COLUMN IF NOT EXISTS message_id UUID;
CREATE INDEX IF NOT EXISTS idx_notifications_message_id ON notifications(message_id);
