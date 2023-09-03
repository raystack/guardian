BEGIN;

-- drop all index we created
ALTER TABLE resources DROP CONSTRAINT fk_resources_provider_type_urn;
ALTER TABLE appeals DROP CONSTRAINT fk_appeals_policy_id_version;

DROP INDEX IF EXISTS activities_provider_activity_provider_idx;
DROP INDEX IF EXISTS providers_type_urn;
DROP INDEX IF EXISTS resources_provider_type_provider_urn_type_urn;

-- create at least all unique index back

CREATE UNIQUE INDEX provider_activity_index ON activities(provider_activity_id, provider_id);
CREATE UNIQUE INDEX provider_index ON providers(type,urn);
CREATE UNIQUE INDEX resource_index ON resources(provider_type,provider_urn,type,urn);

-- drop all columns we created

DROP INDEX IF EXISTS idx_activities_namespace_id;
ALTER TABLE activities DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_appeals_namespace_id;
ALTER TABLE appeals DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_approvals_namespace_id;
ALTER TABLE approvals DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_approvers_namespace_id;
ALTER TABLE approvers DROP COLUMN IF EXISTS namespace_id;

-- DROP INDEX IF EXISTS idx_audit_logs_namespace_id;
-- ALTER TABLE audit_logs DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_grants_namespace_id;
ALTER TABLE grants DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_policies_namespace_id;
ALTER TABLE policies DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_providers_namespace_id;
ALTER TABLE providers DROP COLUMN IF EXISTS namespace_id;

DROP INDEX IF EXISTS idx_resources_namespace_id;
ALTER TABLE resources DROP COLUMN IF EXISTS namespace_id;

----

DROP TABLE IF EXISTS namespaces;

COMMIT;