CREATE TABLE orp_node_health_sample (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  node_id BIGINT NOT NULL,
  sampled_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  ok BOOLEAN NOT NULL,
  rtt_ms DOUBLE NULL,
  active_connections INT NULL,
  requests_total BIGINT NULL,
  KEY idx_orp_node_health_sample (node_id, sampled_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
