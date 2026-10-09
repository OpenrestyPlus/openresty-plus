CREATE TABLE ip_policy (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  mode VARCHAR(16) NOT NULL,
  priority INT NOT NULL DEFAULT 100,
  scope VARCHAR(32) NOT NULL,
  target_resource_id BINARY(16) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ip_rules JSON NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_ip_policy_center FOREIGN KEY (center_id) REFERENCES center(id),
  CONSTRAINT ck_ip_policy_priority CHECK (priority >= 0)
) ENGINE=InnoDB;

CREATE INDEX idx_ip_policy_center_scope_target ON ip_policy (center_id, scope, target_resource_id);

CREATE TABLE api_policy (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  mode VARCHAR(16) NOT NULL,
  priority INT NOT NULL DEFAULT 100,
  http_location_id BINARY(16) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_api_policy_center FOREIGN KEY (center_id) REFERENCES center(id),
  CONSTRAINT ck_api_policy_priority CHECK (priority >= 0)
) ENGINE=InnoDB;

CREATE INDEX idx_api_policy_center_location ON api_policy (center_id, http_location_id);

CREATE TABLE api_policy_rule (
  id BINARY(16) NOT NULL,
  api_policy_id BINARY(16) NOT NULL,
  rule_order INT NOT NULL,
  http_method VARCHAR(16) NOT NULL,
  path_pattern VARCHAR(1024) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_api_policy_rule_policy FOREIGN KEY (api_policy_id) REFERENCES api_policy(id) ON DELETE CASCADE
) ENGINE=InnoDB;
