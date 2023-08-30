BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS namespaces (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    name text UNIQUE NOT NULL,
    state text,
    metadata jsonb,
    created_at timestamp DEFAULT NOW(),
    updated_at timestamp DEFAULT NOW(),
    deleted_at timestamp
    );

ALTER TABLE activities ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_activities_namespace_id ON activities(namespace_id);

ALTER TABLE appeals ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_appeals_namespace_id ON appeals(namespace_id);

ALTER TABLE approvals ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_approvals_namespace_id ON approvals(namespace_id);

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS namespace_id uuid NOT NULL DEFAULT uuid_nil();
CREATE INDEX IF NOT EXISTS idx_audit_logs_namespace_id ON audit_logs(namespace_id);

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
ALTER TABLE appeals DROP CONSTRAINT fk_appeals_resource;
ALTER TABLE appeals DROP CONSTRAINT fk_appeals_policy;
ALTER TABLE approvals DROP CONSTRAINT fk_approvals_appeal;
ALTER TABLE approvals DROP CONSTRAINT fk_appeals_approvals;
ALTER TABLE approvers DROP CONSTRAINT fk_approvals_approvers;
ALTER TABLE grants DROP CONSTRAINT fk_grants_resource;
ALTER TABLE grants DROP CONSTRAINT fk_grants_appeal;
ALTER TABLE resources DROP CONSTRAINT fk_resources_parent;
ALTER TABLE activities DROP CONSTRAINT fk_activities_provider;
ALTER TABLE activities DROP CONSTRAINT fk_activities_resource;

DROP INDEX IF EXISTS provider_activity_index
DROP INDEX IF EXISTS provider_index;
DROP INDEX IF EXISTS resource_index;

-- include namespace in unique index/foreign constraints
ALTER TABLE resources
    ADD CONSTRAINT fk_resources_provider_type_urn FOREIGN KEY (namespace_id,provider_type,provider_urn)
    REFERENCES providers(namespace_id,type,urn);
ALTER TABLE appeals
    ADD CONSTRAINT fk_appeals_resource FOREIGN KEY (namespace_id,resource_id) REFERENCES resources(namespace_id,id);
ALTER TABLE appeals
    ADD CONSTRAINT fk_appeals_policy_id_version FOREIGN KEY (namespace_id,policy_id,policy_version) REFERENCES policies(namespace_id,id,version);
ALTER TABLE approvals
    ADD CONSTRAINT fk_approvals_appeal FOREIGN KEY (namespace_id,appeal_id) REFERENCES appeals(namespace_id,id);
ALTER TABLE approvers
    ADD CONSTRAINT fk_approvals_approvers FOREIGN KEY (namespace_id,approval_id) REFERENCES approvals(namespace_id,id);
ALTER TABLE grants
    ADD CONSTRAINT fk_grants_resource_id FOREIGN KEY (namespace_id,resource_id) REFERENCES resources(namespace_id,id);
ALTER TABLE grants
    ADD CONSTRAINT fk_grants_appeal_id FOREIGN KEY (namespace_id,appeal_id) REFERENCES appeals(namespace_id,id);
ALTER TABLE resources
    ADD CONSTRAINT fk_resources_parent_id FOREIGN KEY (namespace_id,parent_id) REFERENCES resources(namespace_id,id);
ALTER TABLE activities
    ADD CONSTRAINT fk_activities_provider_id FOREIGN KEY (namespace_id,provider_id) REFERENCES providers(namespace_id,id);
ALTER TABLE activities
    ADD CONSTRAINT fk_activities_resource_id FOREIGN KEY (namespace_id,resource_id) REFERENCES resources(namespace_id,id);

CREATE UNIQUE INDEX activities_provider_activity_provider_idx ON activities(namespace_id, provider_activity_id, provider_id);
CREATE UNIQUE INDEX providers_type_urn ON providers(namespace_id,type,urn);
CREATE UNIQUE INDEX resources_provider_type_provider_urn_type_urn ON resources(namespace_id,provider_type,provider_urn,type,urn);


COMMIT;