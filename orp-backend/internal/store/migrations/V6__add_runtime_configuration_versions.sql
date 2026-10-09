CREATE TABLE runtime_configuration_version (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  version_no BIGINT NOT NULL,
  checksum CHAR(64) NOT NULL,
  state VARCHAR(16) NOT NULL,
  content JSON NOT NULL,
  created_by VARCHAR(128) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT uk_runtime_configuration_version UNIQUE (center_id, version_no),
  INDEX idx_runtime_configuration_center_created (center_id, created_at)
) ENGINE=InnoDB;

ALTER TABLE audit_event ADD COLUMN detail JSON NULL;
