BEGIN;

ALTER TABLE
  "resources"
ADD
  COLUMN IF NOT EXISTS "parent_id" uuid;

ALTER TABLE
  "resources"
ADD
  CONSTRAINT "fk_resources_parent" FOREIGN KEY ("parent_id") REFERENCES "resources"("id");

COMMIT;