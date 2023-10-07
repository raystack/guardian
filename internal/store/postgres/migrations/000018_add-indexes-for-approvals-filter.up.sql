CREATE INDEX if NOT EXISTS idx_appeals_account_id_at_null ON appeals (account_id, deleted_at)
WHERE
    deleted_at IS NULL;

CREATE INDEX if NOT EXISTS idx_appeals_account_type_at_null ON appeals (account_type, deleted_at)
WHERE
    deleted_at IS NULL;

CREATE INDEX if NOT EXISTS idx_appeals_role_at_null ON appeals (role, deleted_at)
WHERE
    deleted_at IS NULL;

CREATE INDEX if NOT EXISTS idx_resources_type_at_null ON resources (type, deleted_at)
WHERE
    deleted_at IS NULL;

CREATE INDEX if NOT EXISTS idx_resources_urn_at_null ON resources (urn, deleted_at)
WHERE
    deleted_at IS NULL;