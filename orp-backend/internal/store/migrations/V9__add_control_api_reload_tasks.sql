CREATE TABLE control_api_reload_task (
  id BINARY(16) NOT NULL,
  center_id BINARY(16) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL,
  completed_at TIMESTAMP(6) NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_control_api_reload_task_center FOREIGN KEY (center_id) REFERENCES center(id),
  INDEX idx_control_api_reload_task_center_created (center_id, created_at)
) ENGINE=InnoDB;

CREATE TABLE control_api_reload_node_result (
  id BINARY(16) NOT NULL,
  task_id BINARY(16) NOT NULL,
  node_id BINARY(16) NOT NULL,
  node_name VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  http_status INT NULL,
  message VARCHAR(1024) NULL,
  completed_at TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_control_api_reload_node_result_task FOREIGN KEY (task_id) REFERENCES control_api_reload_task(id) ON DELETE CASCADE,
  INDEX idx_control_api_reload_node_result_task (task_id)
) ENGINE=InnoDB;
