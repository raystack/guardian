CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "providers" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "type" text,
  "urn" text,
  "config" JSONB,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,
  PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "idx_providers_deleted_at" ON "providers" ("deleted_at");
CREATE UNIQUE INDEX IF NOT EXISTS "provider_index" ON "providers" ("type", "urn");