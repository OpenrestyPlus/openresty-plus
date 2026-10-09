CREATE TABLE orp_agent_heartbeat (
  node_id BIGINT NOT NULL PRIMARY KEY,
  certificate_fingerprint CHAR(64) NOT NULL,
  state JSON NOT NULL,
  received_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  KEY idx_orp_agent_heartbeat_received (received_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orp_agent_task (
  id CHAR(36) NOT NULL PRIMARY KEY,
  node_id BIGINT NOT NULL,
  operation VARCHAR(24) NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'queued',
  issued_at DATETIME(6) NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  claimed_at DATETIME(6) NULL,
  started_at DATETIME(6) NULL,
  completed_at DATETIME(6) NULL,
  result_status VARCHAR(24) NULL,
  output_text TEXT NULL,
  error_text TEXT NULL,
  created_by VARCHAR(128) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  KEY idx_orp_agent_task_next (node_id, status, issued_at),
  KEY idx_orp_agent_task_expiry (status, expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
