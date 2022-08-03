ALTER TABLE "policies"
    ADD "appeal_config" jsonb default '{"allow_cross_individual_user":false}';