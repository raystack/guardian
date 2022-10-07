BEGIN;

ALTER TABLE
  "grants"
ADD
  COLUMN IF NOT EXISTS "created_by" text;

UPDATE
  "grants"
SET
  "created_by" = "owner";

ALTER TABLE
  "grants" DROP COLUMN IF EXISTS "source",
  DROP COLUMN IF EXISTS "status_in_provider",
  DROP COLUMN IF EXISTS "owner";

END;