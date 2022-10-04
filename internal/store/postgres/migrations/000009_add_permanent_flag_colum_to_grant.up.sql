ALTER TABLE
  "grants"
ADD
  COLUMN "is_permanent" boolean NOT NULL DEFAULT false;

UPDATE
  "grants"
SET
  "is_permanent" = true
WHERE
  "expiration_date" IS NULL
  OR "expiration_date" <= '0001-01-02';