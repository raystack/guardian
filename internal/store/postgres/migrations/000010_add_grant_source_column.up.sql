ALTER TABLE
  "grants"
ADD
  IF NOT EXISTS "source" text;

UPDATE
  "grants"
SET
  "source" = 'appeal';