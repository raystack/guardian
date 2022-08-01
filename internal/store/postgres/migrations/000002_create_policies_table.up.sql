CREATE TABLE IF NOT EXISTS "policies" (
  "id" text,
  "version" bigint,
  "description" text,
  "steps" JSONB,
  "labels" JSONB,
  "requirements" JSONB,
  "iam" JSONB,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,

  PRIMARY KEY ("id", "version")
);

CREATE INDEX IF NOT EXISTS "idx_policies_deleted_at" ON "policies" ("deleted_at");