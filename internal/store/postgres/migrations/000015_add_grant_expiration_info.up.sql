ALTER TABLE
  grants
ADD
  COLUMN IF NOT EXISTS "expiration_date_reason" text,
ADD
  COLUMN IF NOT EXISTS "requested_expiration_date" timestamp with time zone;
  