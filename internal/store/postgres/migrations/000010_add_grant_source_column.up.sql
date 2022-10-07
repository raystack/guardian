BEGIN;

ALTER TABLE
  "grants"
ADD
  COLUMN IF NOT EXISTS "source" text,
ADD
  COLUMN IF NOT EXISTS "status_in_provider" text,
ADD
  COLUMN IF NOT EXISTS "owner" text;

UPDATE
  "grants"
SET
  "source" = 'appeal',
  "status_in_provider" = "status",
  "owner" = "created_by";

ALTER TABLE
  "grants" DROP COLUMN IF EXISTS "created_by";

COMMIT;