BEGIN;

-- add additional columns
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS namespaces (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    name text,
    state text,
    metadata jsonb,
    created_at timestamptz DEFAULT NOW(),
    updated_at timestamptz DEFAULT NOW(),
    deleted_at timestamptz
    );

ALTER TABLE activities ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_activities_namespace_id ON activities(namespace_id);

ALTER TABLE appeals ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_appeals_namespace_id ON appeals(namespace_id);

ALTER TABLE approvals ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_approvals_namespace_id ON approvals(namespace_id);

ALTER TABLE approvers ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_approvers_namespace_id ON approvers(namespace_id);

-- not doing it for audit_logs as the table is not owned by us
-- ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
-- CREATE INDEX IF NOT EXISTS idx_audit_logs_namespace_id ON audit_logs(namespace_id);

ALTER TABLE grants ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_grants_namespace_id ON grants(namespace_id);

ALTER TABLE policies ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_policies_namespace_id ON policies(namespace_id);

ALTER TABLE providers ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_providers_namespace_id ON providers(namespace_id);

ALTER TABLE resources ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_resources_namespace_id ON resources(namespace_id);

-- drop all unique index/foreign constraints in use
ALTER TABLE resources DROP CONSTRAINT fk_resources_provider;
ALTER TABLE approvals DROP CONSTRAINT fk_appeals_approvals;

DROP INDEX IF EXISTS provider_activity_index;
DROP INDEX IF EXISTS provider_index;
DROP INDEX IF EXISTS resource_index;

-- include namespace in unique index/foreign constraints

CREATE UNIQUE INDEX activities_provider_activity_provider_idx ON activities(namespace_id, provider_activity_id, provider_id);
CREATE UNIQUE INDEX providers_type_urn ON providers(namespace_id,type,urn);
CREATE UNIQUE INDEX resources_provider_type_provider_urn_type_urn ON resources(namespace_id,provider_type,provider_urn,type,urn);

ALTER TABLE resources
    ADD CONSTRAINT fk_resources_provider_type_urn FOREIGN KEY (namespace_id,provider_type,provider_urn)
    REFERENCES providers(namespace_id,type,urn);

COMMIT;