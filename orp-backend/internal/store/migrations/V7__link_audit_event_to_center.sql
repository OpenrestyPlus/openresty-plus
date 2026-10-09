ALTER TABLE audit_event ADD COLUMN center_id BINARY(16) NULL;
CREATE INDEX idx_audit_center_created_at ON audit_event (center_id, created_at);
