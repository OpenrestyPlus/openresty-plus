ALTER TABLE http_upstream
  ADD COLUMN health_check_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN health_check_path VARCHAR(255) NOT NULL DEFAULT '/health',
  ADD COLUMN health_check_interval_seconds INT NOT NULL DEFAULT 10,
  ADD COLUMN health_check_timeout_milliseconds INT NOT NULL DEFAULT 1000,
  ADD COLUMN health_check_expected_status INT NOT NULL DEFAULT 200;
