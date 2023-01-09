BEGIN;

ALTER TABLE
  "resources" DROP CONSTRAINT IF EXISTS "fk_resources_parent";

ALTER TABLE
  "resources" DROP COLUMN IF EXISTS "parent_id";

COMMIT;