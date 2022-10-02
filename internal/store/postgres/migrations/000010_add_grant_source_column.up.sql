ALTER TABLE
  "grants"
ADD
  COLUMN IF NOT EXISTS "source" text,
ADD
  COLUMN IF NOT EXISTS "status_in_provider" text;

UPDATE
  "grants"
SET
  "source" = 'appeal';