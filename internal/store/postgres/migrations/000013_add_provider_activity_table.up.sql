CREATE TABLE IF NOT EXISTS "activities" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "provider_activity_id" text,
  "provider_id" uuid,
  "resource_id" uuid,
  "account_type" text,
  "account_id" text,
  "timestamp" timestamptz,
  "authorizations" text [],
  "type" text,
  "metadata" jsonb,
  "created_at" timestamptz,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_activities_provider" FOREIGN KEY ("provider_id") REFERENCES "providers"("id"),
  CONSTRAINT "fk_activities_resource" FOREIGN KEY ("resource_id") REFERENCES "resources"("id")
);

CREATE INDEX IF NOT EXISTS "idx_activities_provider_id" ON "activities" ("provider_id");
CREATE INDEX IF NOT EXISTS "idx_activities_account_id" ON "activities" ("account_id");
CREATE INDEX IF NOT EXISTS "idx_activities_timestamp" ON "activities" ("timestamp");
CREATE UNIQUE INDEX IF NOT EXISTS "provider_activity_index" ON "activities" ("provider_activity_id", "provider_id");
