CREATE TABLE stream_upstream (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  name VARCHAR(128) NOT NULL,
  target_host VARCHAR(255) NOT NULL,
  target_port INT NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT uk_stream_upstream_center_name UNIQUE (center_id, name)
) ENGINE=InnoDB;

CREATE TABLE stream_server (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  service_name VARCHAR(128) NOT NULL,
  listen_port INT NOT NULL,
  protocol VARCHAR(8) NOT NULL,
  upstream_id BINARY(16) NOT NULL,
  access_log VARCHAR(512) NOT NULL,
  error_log VARCHAR(512) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT uk_stream_server_center_port_protocol UNIQUE (center_id, listen_port, protocol)
) ENGINE=InnoDB;
