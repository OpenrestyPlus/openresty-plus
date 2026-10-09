ALTER TABLE http_configuration
  ADD COLUMN sendfile_enabled BOOLEAN NOT NULL DEFAULT TRUE AFTER hide_version,
  ADD COLUMN tcp_nopush_enabled BOOLEAN NOT NULL DEFAULT TRUE AFTER sendfile_enabled,
  ADD COLUMN tcp_nodelay_enabled BOOLEAN NOT NULL DEFAULT TRUE AFTER tcp_nopush_enabled,
  ADD COLUMN keepalive_timeout_seconds INT NOT NULL DEFAULT 65 AFTER tcp_nodelay_enabled,
  ADD COLUMN client_max_body_size VARCHAR(32) NOT NULL DEFAULT '10m' AFTER keepalive_timeout_seconds,
  ADD COLUMN client_header_buffer_size VARCHAR(32) NOT NULL DEFAULT '1k' AFTER client_max_body_size,
  ADD COLUMN large_client_header_buffers VARCHAR(64) NOT NULL DEFAULT '4 8k' AFTER client_header_buffer_size;
