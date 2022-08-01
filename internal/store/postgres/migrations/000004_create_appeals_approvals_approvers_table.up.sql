CREATE TABLE IF NOT EXISTS "appeals" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "resource_id" uuid,
  "policy_id" text,
  "policy_version" bigint,
  "status" text,
  "account_id" text,
  "account_type" text,
  "created_by" text,
  "creator" JSONB,
  "role" text,
  "options" JSONB,
  "labels" JSONB,
  "details" JSONB,
  "revoked_by" text,
  "revoked_at" timestamptz,
  "revoke_reason" text,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_appeals_resource" FOREIGN KEY ("resource_id") REFERENCES "resources"("id"),
  CONSTRAINT "fk_appeals_policy" FOREIGN KEY ("policy_id", "policy_version") REFERENCES "policies"("id", "version")
);

CREATE INDEX IF NOT EXISTS "idx_appeals_deleted_at" ON "appeals" ("deleted_at");

CREATE TABLE IF NOT EXISTS "approvals" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "name" text,
  "index" bigint,
  "appeal_id" uuid,
  "status" text,
  "actor" text,
  "reason" text,
  "policy_id" text,
  "policy_version" bigint,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_approvals_appeal" FOREIGN KEY ("appeal_id") REFERENCES "appeals"("id"),
  CONSTRAINT "fk_appeals_approvals" FOREIGN KEY ("appeal_id") REFERENCES "appeals"("id")
);

CREATE INDEX IF NOT EXISTS "idx_approvals_name" ON "approvals" ("name");
CREATE INDEX IF NOT EXISTS "idx_approvals_deleted_at" ON "approvals" ("deleted_at");

CREATE TABLE IF NOT EXISTS "approvers" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "approval_id" uuid,
  "appeal_id" text,
  "email" text,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_approvals_approvers" FOREIGN KEY ("approval_id") REFERENCES "approvals"("id")
);

CREATE INDEX IF NOT EXISTS "idx_approvers_deleted_at" ON "approvers" ("deleted_at");
CREATE INDEX IF NOT EXISTS "idx_approvers_email" ON "approvers" ("email");
CREATE INDEX IF NOT EXISTS "idx_approvers_appeal_id" ON "approvers" ("appeal_id");