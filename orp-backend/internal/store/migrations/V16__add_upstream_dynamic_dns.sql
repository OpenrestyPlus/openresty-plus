ALTER TABLE http_upstream
  ADD COLUMN zone_size_kilobytes INT NOT NULL DEFAULT 64 AFTER keepalive_connections;

ALTER TABLE http_upstream_target
  ADD COLUMN resolve_enabled BOOLEAN NOT NULL DEFAULT FALSE AFTER fail_timeout_seconds;

ALTER TABLE stream_upstream
  ADD COLUMN resolve_enabled BOOLEAN NOT NULL DEFAULT FALSE AFTER target_port,
  ADD COLUMN zone_size_kilobytes INT NOT NULL DEFAULT 64 AFTER resolve_enabled;
