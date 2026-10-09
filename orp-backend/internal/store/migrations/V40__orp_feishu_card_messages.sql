ALTER TABLE orp_alert_channel
  ADD COLUMN message_format VARCHAR(16) NOT NULL DEFAULT 'text' AFTER channel_type;
