ALTER TABLE
  "appeals"
ADD
  COLUMN "revoked_by" text,
ADD
  COLUMN "revoked_at" timestamptz,
ADD
  COLUMN "revoke_reason" text;

UPDATE
  "appeals"
SET
  "revoked_by" = "grants"."revoked_by",
  "revoked_at" = "grants"."revoked_at",
  "revoke_reason" = "grants"."revoke_reason"
FROM
  "grants"
WHERE
  "grants"."appeal_id" = "appeals"."id";

UPDATE
  "appeals"
SET
  "status" = 'terminated'
WHERE
  "revoked_by" <> '';