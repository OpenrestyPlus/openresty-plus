CREATE TABLE http_upstream_target (
  id BINARY(16) NOT NULL,
  upstream_id BINARY(16) NOT NULL,
  target_host VARCHAR(255) NOT NULL,
  target_port INT NOT NULL,
  weight INT NOT NULL DEFAULT 1,
  max_fails INT NOT NULL DEFAULT 3,
  fail_timeout_seconds INT NOT NULL DEFAULT 10,
  backup BOOLEAN NOT NULL DEFAULT FALSE,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  PRIMARY KEY (id),
  CONSTRAINT fk_http_upstream_target_upstream FOREIGN KEY (upstream_id) REFERENCES http_upstream(id) ON DELETE CASCADE,
  INDEX idx_http_upstream_target_upstream (upstream_id)
) ENGINE=InnoDB;
