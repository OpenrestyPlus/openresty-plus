CREATE TABLE configuration_dictionary (
  id BINARY(16) NOT NULL,
  dictionary_type VARCHAR(64) NOT NULL,
  dictionary_code VARCHAR(128) NOT NULL,
  dictionary_name VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  PRIMARY KEY (id),
  UNIQUE KEY uk_configuration_dictionary_type_code (dictionary_type, dictionary_code),
  INDEX idx_configuration_dictionary_type_enabled_sort (dictionary_type, enabled, sort_order)
) ENGINE=InnoDB;

INSERT INTO configuration_dictionary (id, dictionary_type, dictionary_code, dictionary_name, sort_order, enabled) VALUES
  (UUID_TO_BIN(UUID()), 'HTTP_CONTENT_TYPE', 'application/json', 'JSON', 10, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_CONTENT_TYPE', 'application/x-www-form-urlencoded', '表单编码', 20, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_CONTENT_TYPE', 'multipart/form-data', '文件上传表单', 30, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_CONTENT_TYPE', 'application/xml', 'XML', 40, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_CONTENT_TYPE', 'text/plain', '纯文本', 50, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_METHOD', 'GET', '查询', 10, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_METHOD', 'POST', '创建', 20, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_METHOD', 'PUT', '全量更新', 30, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_METHOD', 'PATCH', '部分更新', 40, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_METHOD', 'DELETE', '删除', 50, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_METHOD', 'HEAD', '仅响应头', 60, TRUE),
  (UUID_TO_BIN(UUID()), 'HTTP_METHOD', 'OPTIONS', '跨域预检', 70, TRUE);
