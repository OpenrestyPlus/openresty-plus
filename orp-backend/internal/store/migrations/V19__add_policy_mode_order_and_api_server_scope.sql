ALTER TABLE http_server
  ADD COLUMN ip_policy_mode_order VARCHAR(24) NOT NULL DEFAULT 'BLACKLIST_FIRST',
  ADD COLUMN api_policy_mode_order VARCHAR(24) NOT NULL DEFAULT 'BLACKLIST_FIRST';

ALTER TABLE stream_server
  ADD COLUMN ip_policy_mode_order VARCHAR(24) NOT NULL DEFAULT 'BLACKLIST_FIRST';

ALTER TABLE api_policy
  MODIFY COLUMN http_location_id BINARY(16) NULL,
  ADD COLUMN scope VARCHAR(24) NOT NULL DEFAULT 'HTTP_LOCATION',
  ADD COLUMN target_resource_id BINARY(16) NULL;

UPDATE api_policy
SET target_resource_id = http_location_id
WHERE target_resource_id IS NULL;

ALTER TABLE api_policy
  MODIFY COLUMN target_resource_id BINARY(16) NOT NULL,
  ADD INDEX idx_api_policy_center_scope_target_priority (center_id, scope, target_resource_id, priority);
