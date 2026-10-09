CREATE TABLE orp_publish_candidate (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  release_uuid CHAR(36) NOT NULL,
  center_id BIGINT NOT NULL,
  actor VARCHAR(128) NOT NULL,
  target_ids JSON NOT NULL,
  snapshot JSON NOT NULL,
  config_text LONGTEXT NOT NULL,
  digest CHAR(64) NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'frozen',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY uk_orp_candidate_release (release_uuid),
  KEY idx_orp_candidate_center (center_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orp_publish_precheck (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  candidate_id BIGINT NOT NULL,
  node_id BIGINT NOT NULL,
  passed BOOLEAN NOT NULL,
  command_text VARCHAR(512) NOT NULL,
  output_text TEXT NOT NULL,
  checked_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  KEY idx_orp_precheck_candidate (candidate_id, node_id, checked_at),
  CONSTRAINT fk_orp_precheck_candidate FOREIGN KEY (candidate_id) REFERENCES orp_publish_candidate(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orp_publish_batch (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  actor VARCHAR(128) NOT NULL,
  mode VARCHAR(16) NOT NULL,
  phase VARCHAR(24) NOT NULL,
  canary_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  canary_wait_seconds INT NOT NULL DEFAULT 0,
  current_batch_index INT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  finished_at DATETIME(6) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orp_publish_item (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  batch_id BIGINT NOT NULL,
  candidate_id BIGINT NOT NULL,
  node_id BIGINT NOT NULL,
  center_id BIGINT NOT NULL,
  batch_index INT NOT NULL DEFAULT 0,
  weight INT NOT NULL DEFAULT 10,
  status VARCHAR(24) NOT NULL DEFAULT 'queued',
  error_text TEXT NULL,
  previous_release_uuid CHAR(36) NULL,
  started_at DATETIME(6) NULL,
  finished_at DATETIME(6) NULL,
  CONSTRAINT fk_orp_publish_item_batch FOREIGN KEY (batch_id) REFERENCES orp_publish_batch(id),
  CONSTRAINT fk_orp_publish_item_candidate FOREIGN KEY (candidate_id) REFERENCES orp_publish_candidate(id),
  KEY idx_orp_publish_item_batch (batch_id, batch_index),
  KEY idx_orp_publish_item_node (node_id, finished_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
