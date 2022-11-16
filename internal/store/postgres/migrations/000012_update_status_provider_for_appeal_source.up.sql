UPDATE
    "grants"
SET
    "status_in_provider" = "status"
WHERE
    "source" = 'appeal'
    AND "status_in_provider" != "status"