CREATE TABLE orp_alert_channel (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  channel_type VARCHAR(16) NOT NULL,
  target_encrypted TEXT NOT NULL,
  secret_encrypted TEXT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uk_orp_alert_channel_name (name),
  KEY idx_orp_alert_channel_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orp_tls_alert_rule (
  certificate_id BIGINT NOT NULL PRIMARY KEY,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  days_before INT NOT NULL DEFAULT 30,
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  KEY idx_orp_tls_alert_rule_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orp_tls_alert_rule_channel (
  certificate_id BIGINT NOT NULL,
  channel_id BIGINT NOT NULL,
  PRIMARY KEY (certificate_id, channel_id),
  CONSTRAINT fk_orp_tls_alert_rule_channel_rule FOREIGN KEY (certificate_id)
    REFERENCES orp_tls_alert_rule(certificate_id) ON DELETE CASCADE,
  CONSTRAINT fk_orp_tls_alert_rule_channel_channel FOREIGN KEY (channel_id)
    REFERENCES orp_alert_channel(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orp_alert_delivery (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  event_key VARCHAR(255) NOT NULL,
  event_type VARCHAR(32) NOT NULL,
  channel_id BIGINT NOT NULL,
  channel_name VARCHAR(128) NOT NULL,
  certificate_id BIGINT NULL,
  certificate_name VARCHAR(255) NOT NULL DEFAULT '',
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'pending',
  attempt_count INT NOT NULL DEFAULT 0,
  response_status INT NULL,
  response_text VARCHAR(2048) NULL,
  error_text VARCHAR(2048) NULL,
  next_attempt_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  delivered_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uk_orp_alert_delivery_event_channel (event_key, channel_id),
  KEY idx_orp_alert_delivery_retry (status, next_attempt_at),
  KEY idx_orp_alert_delivery_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
