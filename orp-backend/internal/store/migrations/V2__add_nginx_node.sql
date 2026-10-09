CREATE TABLE nginx_node (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  name VARCHAR(128) NOT NULL,
  protocol VARCHAR(32) NOT NULL,
  host VARCHAR(255) NOT NULL,
  service_port INT NOT NULL,
  control_api_url VARCHAR(512),
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_nginx_node_center FOREIGN KEY (center_id) REFERENCES center(id),
  CONSTRAINT uk_nginx_node_center_name UNIQUE (center_id, name),
  INDEX idx_nginx_node_center (center_id)
) ENGINE=InnoDB;
