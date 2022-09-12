UPDATE
  "appeals"
SET
  "status" = "approved"
WHERE
  "status" = "terminated";

ALTER TABLE
  "appeals" DROP COLUMN IF EXISTS "revoked_by",
  "revoked_at",
  "revoke_reason";