CREATE TABLE IF NOT EXISTS "activities" (
  "id" uuid DEFAULT uuid_generate_v4(),
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
  CONSTRAINT "fk_provider_activities_provider" FOREIGN KEY ("provider_id") REFERENCES "providers"("id"),
  CONSTRAINT "fk_provider_activities_resource" FOREIGN KEY ("resource_id") REFERENCES "resources"("id")
);