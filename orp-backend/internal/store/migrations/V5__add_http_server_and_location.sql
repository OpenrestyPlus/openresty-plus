CREATE TABLE http_server (
  id BINARY(16) NOT NULL, center_id BINARY(16) NOT NULL, domain VARCHAR(255) NOT NULL, listen_port INT NOT NULL,
  ssl_enabled BOOLEAN NOT NULL DEFAULT FALSE, upstream_id BINARY(16), access_log VARCHAR(512) NOT NULL, error_log VARCHAR(512) NOT NULL, created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id), CONSTRAINT fk_http_server_center FOREIGN KEY (center_id) REFERENCES center(id), CONSTRAINT fk_http_server_upstream FOREIGN KEY (upstream_id) REFERENCES http_upstream(id),
  CONSTRAINT uk_http_server_center_domain_port UNIQUE (center_id, domain, listen_port), INDEX idx_http_server_center (center_id)
) ENGINE=InnoDB;
CREATE TABLE http_location (
  id BINARY(16) NOT NULL, server_id BINARY(16) NOT NULL, path VARCHAR(512) NOT NULL, methods JSON NOT NULL, content_types JSON NOT NULL,
  header_length_min INT NOT NULL DEFAULT 0, header_length_max INT NOT NULL, body_length_min BIGINT NOT NULL DEFAULT 0, body_length_max BIGINT NOT NULL,
  upstream_id BINARY(16) NOT NULL, proxy_connect_timeout_ms INT NOT NULL, proxy_read_timeout_ms INT NOT NULL, proxy_send_timeout_ms INT NOT NULL, created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id), CONSTRAINT fk_http_location_server FOREIGN KEY (server_id) REFERENCES http_server(id) ON DELETE CASCADE, CONSTRAINT fk_http_location_upstream FOREIGN KEY (upstream_id) REFERENCES http_upstream(id),
  CONSTRAINT uk_http_location_server_path UNIQUE (server_id, path), INDEX idx_http_location_server (server_id), INDEX idx_http_location_upstream (upstream_id)
) ENGINE=InnoDB;
