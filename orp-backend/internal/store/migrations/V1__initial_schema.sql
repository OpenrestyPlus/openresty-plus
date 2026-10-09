CREATE TABLE center (
  id BINARY(16) NOT NULL,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT uk_center_code UNIQUE (code)
) ENGINE=InnoDB;

CREATE TABLE audit_event (
  id BINARY(16) NOT NULL,
  actor VARCHAR(128) NOT NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(128) NOT NULL,
  result VARCHAR(32) NOT NULL,
  request_id VARCHAR(128),
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  INDEX idx_audit_created_at (created_at),
  INDEX idx_audit_resource (resource_type, resource_id)
) ENGINE=InnoDB;
