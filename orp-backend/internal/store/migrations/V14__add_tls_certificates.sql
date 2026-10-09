CREATE TABLE tls_certificate (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  name VARCHAR(128) NOT NULL,
  common_name VARCHAR(255) NOT NULL,
  certificate_pem MEDIUMTEXT NOT NULL,
  private_key_pem MEDIUMTEXT NOT NULL,
  chain_pem MEDIUMTEXT,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_tls_certificate_center FOREIGN KEY (center_id) REFERENCES center(id),
  CONSTRAINT uk_tls_certificate_center_name UNIQUE (center_id, name),
  INDEX idx_tls_certificate_center (center_id)
) ENGINE=InnoDB;

ALTER TABLE http_server ADD COLUMN certificate_id BINARY(16) NULL AFTER ssl_enabled;
ALTER TABLE http_server ADD CONSTRAINT fk_http_server_certificate FOREIGN KEY (certificate_id) REFERENCES tls_certificate(id);
