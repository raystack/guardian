UPDATE
  "appeals"
SET
  "status" = 'approved'
WHERE
  "status" = 'terminated';

ALTER TABLE
  "appeals" DROP COLUMN IF EXISTS "revoked_by",
  DROP COLUMN IF EXISTS "revoked_at",
  DROP COLUMN IF EXISTS "revoke_reason";