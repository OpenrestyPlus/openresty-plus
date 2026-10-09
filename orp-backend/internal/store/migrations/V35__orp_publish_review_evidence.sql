ALTER TABLE orp_publish_precheck
  ADD COLUMN baseline_release_uuid CHAR(36) NULL,
  ADD COLUMN baseline_digest CHAR(64) NULL,
  ADD COLUMN semantic_diff JSON NULL,
  ADD COLUMN text_diff JSON NULL,
  ADD COLUMN publish_plan JSON NULL;
