ALTER TABLE http_configuration
  ADD COLUMN server_names_hash_bucket_size INT NOT NULL DEFAULT 512 AFTER large_client_header_buffers,
  ADD COLUMN gzip_enabled BOOLEAN NOT NULL DEFAULT TRUE AFTER server_names_hash_bucket_size,
  ADD COLUMN gzip_min_length VARCHAR(32) NOT NULL DEFAULT '1k' AFTER gzip_enabled,
  ADD COLUMN gzip_comp_level INT NOT NULL DEFAULT 2 AFTER gzip_min_length;
