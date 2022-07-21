BEGIN;

ALTER TABLE
  "appeals"
ADD
  COLUMN "permissions" text [];

-- backfill permissions for existing appeals
WITH "provider_roles_json" AS (
  SELECT
    "type",
    "urn",
    jsonb_array_elements(config -> 'resources') ->> 'type' AS "resource_type",
    jsonb_array_elements(config -> 'resources') -> 'roles' AS "roles"
  FROM
    "providers"
),
"provider_roles" AS (
  SELECT
    "type",
    "urn",
    "resource_type",
    jsonb_array_elements(
      CASE
        jsonb_typeof(roles)
        WHEN 'array' THEN roles
        ELSE '[]'
      END
    ) ->> 'id' AS "role",
    (
      SELECT
        ARRAY (
          SELECT
            jsonb_array_elements_text(
              jsonb_array_elements(
                CASE
                  jsonb_typeof(roles)
                  WHEN 'array' THEN roles
                  ELSE '[]'
                END
              ) -> 'permissions'
            )
        )
    ) AS "permissions"
  FROM
    "provider_roles_json"
)
UPDATE
  "appeals"
SET
  "permissions" = "provider_roles"."permissions"
FROM
  "appeals" a
  LEFT JOIN "resources" ON a."resource_id" = "resources"."id"
  LEFT JOIN "provider_roles" ON "resources"."provider_type" = "provider_roles"."type"
  AND "resources"."provider_urn" = "provider_roles"."urn"
WHERE
  "appeals"."id" = a."id";

COMMIT;