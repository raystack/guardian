CREATE INDEX IF NOT EXISTS idx_appeals_created_by_status_deleted_at_updated_at
ON appeals (created_by, status, deleted_at, updated_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_appeals_id_deleted_at_null
ON appeals (id, deleted_at)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_approvers_approval_id_deleted_at_null
ON approvers (approval_id, deleted_at)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_grants_appeal_id_deleted_at_null
ON grants (appeal_id, deleted_at)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_grants_status_deleted_at_null
ON grants (status, deleted_at)
WHERE deleted_at IS NULL and status = 'active';

CREATE INDEX IF NOT EXISTS idx_resources_id_deleted_at_null
ON resources (id, deleted_at)
WHERE deleted_at IS NULL;

