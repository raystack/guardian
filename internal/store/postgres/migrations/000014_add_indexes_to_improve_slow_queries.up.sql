-- This index reducing query cost from 13997.45..13997 to 17.66..17.67
-- SELECT ...
-- FROM "appeals"
--     LEFT JOIN "resources" "Resource" ON "appeals"."resource_id" = "Resource"."id"
--     AND "Resource"."deleted_at" IS NULL
--     LEFT JOIN "grants" "Grant" ON "appeals"."id" = "Grant"."appeal_id"
--     AND "Grant"."deleted_at" IS NULL
-- WHERE "appeals"."created_by" = '<creator>'
--     AND "appeals"."status" IN (...)
--     AND "appeals"."deleted_at" IS NULL
-- ORDER BY ARRAY_POSITION(
--         ARRAY [...],
--         "appeals"."status"
--     ),
--     "updated_at" desc;

CREATE INDEX IF NOT EXISTS idx_appeals_created_by_status_deleted_at_updated_at
ON appeals (created_by, status, deleted_at, updated_at DESC)
WHERE deleted_at IS NULL;

-- This index reducing query cost from 0.15..8.17 to 0.00..1.02
-- Since we have a lot of queries like this due to the gorm soft delete:
-- SELECT * FROM <table> WHERE id = '<uuid>' AND "deleted_at" IS NULL ORDER BY "id" LIMIT 1
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

