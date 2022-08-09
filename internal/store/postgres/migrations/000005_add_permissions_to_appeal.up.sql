BEGIN;

ALTER TABLE
  "appeals"
ADD
  COLUMN "permissions" text [];

-- backfill permissions for existing appeals
WITH "provider_resources" AS (
  SELECT
    *,
    jsonb_array_elements("config" -> 'resources') AS "resource"
  FROM
    "providers"
),
"provider_role_configs" AS (
  SELECT
    *,
    jsonb_array_elements(
      CASE
        resource -> 'roles' -- add null element to it's included in "role_config"
        WHEN 'null' :: jsonb THEN '[null]' :: jsonb
        WHEN '[]' :: jsonb THEN '[null]' :: jsonb
        ELSE resource -> 'roles'
      END
    ) AS "role_config",
    resource ->> 'type' AS "resource_type"
  FROM
    "provider_resources"
),
"provider_roles" AS (
  SELECT
    *,
    "role_config" ->> 'id' AS "role",
    (
      SELECT
        ARRAY (
          SELECT
            jsonb_array_elements_text("role_config" -> 'permissions')
        )
    ) AS "permissions"
  FROM
    "provider_role_configs"
)
UPDATE
  "appeals"
SET
  "permissions" = (
    CASE
      WHEN "resources"."provider_type" = 'gcloud_iam' THEN string_to_array(a."role", ',')
      ELSE "provider_roles"."permissions"
    END
  )
FROM
  "appeals" a
  LEFT JOIN "resources" ON a."resource_id" = "resources"."id"
  LEFT JOIN "provider_roles" ON "resources"."provider_type" = "provider_roles"."type"
  AND "resources"."provider_urn" = "provider_roles"."urn"
WHERE
  "appeals"."id" = a."id";

ALTER TABLE
  "appeals"
ALTER COLUMN
  "permissions"
SET
  NOT NULL;

COMMIT;