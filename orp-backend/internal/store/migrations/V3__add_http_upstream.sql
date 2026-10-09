CREATE TABLE http_upstream (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  name VARCHAR(128) NOT NULL,
  keepalive_connections INT NOT NULL DEFAULT 32,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_http_upstream_center FOREIGN KEY (center_id) REFERENCES center(id),
  CONSTRAINT uk_http_upstream_center_name UNIQUE (center_id, name)
) ENGINE=InnoDB;

CREATE TABLE http_upstream_member (
  id BINARY(16) NOT NULL,
  upstream_id BINARY(16) NOT NULL,
  host VARCHAR(255) NOT NULL,
  port INT NOT NULL,
  weight INT NOT NULL DEFAULT 1,
  max_fails INT NOT NULL DEFAULT 3,
  fail_timeout_seconds INT NOT NULL DEFAULT 10,
  PRIMARY KEY (id),
  CONSTRAINT fk_http_upstream_member_upstream FOREIGN KEY (upstream_id) REFERENCES http_upstream(id) ON DELETE CASCADE
) ENGINE=InnoDB;
