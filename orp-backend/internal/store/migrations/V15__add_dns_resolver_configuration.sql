CREATE TABLE dns_resolver_configuration (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  scope VARCHAR(32) NOT NULL,
  target_resource_id BINARY(16) NULL,
  resolver_addresses JSON NOT NULL,
  valid_seconds INT NOT NULL DEFAULT 30,
  timeout_milliseconds INT NOT NULL DEFAULT 3000,
  ipv6_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  PRIMARY KEY (id),
  CONSTRAINT fk_dns_resolver_center FOREIGN KEY (center_id) REFERENCES center(id),
  UNIQUE KEY uk_dns_resolver_scope_target (center_id, scope, target_resource_id)
) ENGINE=InnoDB;

ALTER TABLE http_location ADD COLUMN dynamic_dns_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE http_location ADD COLUMN dynamic_dns_host VARCHAR(255) NULL;
ALTER TABLE http_location ADD COLUMN dynamic_dns_port INT NULL;
ALTER TABLE stream_server ADD COLUMN dynamic_dns_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE stream_server ADD COLUMN dynamic_dns_host VARCHAR(255) NULL;
ALTER TABLE stream_server ADD COLUMN dynamic_dns_port INT NULL;
