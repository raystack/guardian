CREATE TABLE IF NOT EXISTS "resources" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "provider_type" text,
  "provider_urn" text,
  "type" text,
  "urn" text,
  "name" text,
  "details" JSONB,
  "labels" JSONB,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,
  "is_deleted" boolean,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_resources_provider" FOREIGN KEY ("provider_type", "provider_urn") REFERENCES "providers"("type", "urn")
);

CREATE INDEX IF NOT EXISTS "idx_resources_deleted_at" ON "resources" ("deleted_at");
CREATE UNIQUE INDEX IF NOT EXISTS "resource_index" ON "resources" ("provider_type", "provider_urn", "type", "urn");