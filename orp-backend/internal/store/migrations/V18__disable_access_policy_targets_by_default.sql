UPDATE http_server SET ip_policy_enabled = FALSE, api_policy_enabled = FALSE;
UPDATE http_location SET ip_policy_enabled = FALSE, api_policy_enabled = FALSE;
UPDATE stream_server SET ip_policy_enabled = FALSE;

ALTER TABLE http_server
  ALTER COLUMN ip_policy_enabled SET DEFAULT FALSE,
  ALTER COLUMN api_policy_enabled SET DEFAULT FALSE;

ALTER TABLE http_location
  ALTER COLUMN ip_policy_enabled SET DEFAULT FALSE,
  ALTER COLUMN api_policy_enabled SET DEFAULT FALSE;

ALTER TABLE stream_server
  ALTER COLUMN ip_policy_enabled SET DEFAULT FALSE;
