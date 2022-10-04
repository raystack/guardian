CREATE TABLE IF NOT EXISTS "grants" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "status" text,
  "account_id" text,
  "account_type" text,
  "resource_id" uuid,
  "role" text,
  "permissions" text [],
  "expiration_date" timestamptz,
  "appeal_id" uuid,
  "revoked_by" text,
  "revoked_at" timestamptz,
  "revoke_reason" text,
  "created_by" text,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_grants_resource" FOREIGN KEY ("resource_id") REFERENCES "resources"("id"),
  CONSTRAINT "fk_grants_appeal" FOREIGN KEY ("appeal_id") REFERENCES "appeals"("id")
);

CREATE INDEX IF NOT EXISTS "idx_grants_deleted_at" ON "grants" ("deleted_at");

INSERT INTO
  "grants" (
    "status",
    "account_id",
    "account_type",
    "resource_id",
    "role",
    "permissions",
    "expiration_date",
    "appeal_id",
    "revoked_by",
    "revoked_at",
    "revoke_reason",
    "created_by",
    "created_at",
    "updated_at"
  )
SELECT
  CASE
    WHEN status = 'active' THEN 'active'
    WHEN status = 'terminated' THEN 'inactive'
  END AS "status",
  "account_id",
  "account_type",
  "resource_id",
  "role",
  "permissions",
  ("options" ->> 'expiration_date') :: TIMESTAMP WITH TIME ZONE AS "expiration_date",
  "id" AS "appeal_id",
  "revoked_by",
  "revoked_at",
  "revoke_reason",
  "created_by",
  "updated_at" AS "created_at",
  "updated_at"
FROM
  "appeals"
WHERE
  status = 'active'
  OR status = 'terminated';