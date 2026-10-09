ALTER TABLE http_upstream
  ADD COLUMN health_check_type VARCHAR(16) NOT NULL DEFAULT 'HTTP' AFTER health_check_enabled,
  ADD COLUMN health_check_host VARCHAR(255) NULL AFTER health_check_path,
  ADD COLUMN health_check_request_headers JSON NOT NULL DEFAULT (JSON_ARRAY()) AFTER health_check_expected_status,
  ADD COLUMN health_check_rise INT NOT NULL DEFAULT 2 AFTER health_check_request_headers,
  ADD COLUMN health_check_fall INT NOT NULL DEFAULT 3 AFTER health_check_rise;

ALTER TABLE stream_upstream
  ADD COLUMN health_check_enabled BOOLEAN NOT NULL DEFAULT FALSE AFTER zone_size_kilobytes,
  ADD COLUMN health_check_type VARCHAR(16) NOT NULL DEFAULT 'TCP' AFTER health_check_enabled,
  ADD COLUMN health_check_path VARCHAR(255) NOT NULL DEFAULT '/health' AFTER health_check_type,
  ADD COLUMN health_check_host VARCHAR(255) NULL AFTER health_check_path,
  ADD COLUMN health_check_interval_seconds INT NOT NULL DEFAULT 10 AFTER health_check_host,
  ADD COLUMN health_check_timeout_milliseconds INT NOT NULL DEFAULT 1000 AFTER health_check_interval_seconds,
  ADD COLUMN health_check_expected_status INT NOT NULL DEFAULT 200 AFTER health_check_timeout_milliseconds,
  ADD COLUMN health_check_request_headers JSON NOT NULL DEFAULT (JSON_ARRAY()) AFTER health_check_expected_status,
  ADD COLUMN health_check_rise INT NOT NULL DEFAULT 2 AFTER health_check_request_headers,
  ADD COLUMN health_check_fall INT NOT NULL DEFAULT 3 AFTER health_check_rise;
