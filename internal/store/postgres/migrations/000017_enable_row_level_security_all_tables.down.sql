BEGIN;

DROP POLICY IF EXISTS activities_isolation_policy ON activities;
DROP POLICY IF EXISTS appeals_isolation_policy ON appeals;
DROP POLICY IF EXISTS approvals_isolation_policy ON approvals;
DROP POLICY IF EXISTS approvers_isolation_policy ON approvers;
-- DROP POLICY IF EXISTS audit_logs_isolation_policy ON audit_logs;
DROP POLICY IF EXISTS grants_isolation_policy ON grants;
DROP POLICY IF EXISTS policies_isolation_policy ON policies;
DROP POLICY IF EXISTS providers_isolation_policy ON providers;
DROP POLICY IF EXISTS resources_isolation_policy ON resources;

ALTER TABLE activities DISABLE ROW LEVEL SECURITY;
ALTER TABLE appeals DISABLE ROW LEVEL SECURITY;
ALTER TABLE approvals DISABLE ROW LEVEL SECURITY;
ALTER TABLE approvers DISABLE ROW LEVEL SECURITY;
-- ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE grants DISABLE ROW LEVEL SECURITY;
ALTER TABLE policies DISABLE ROW LEVEL SECURITY;
ALTER TABLE providers DISABLE ROW LEVEL SECURITY;
ALTER TABLE resources DISABLE ROW LEVEL SECURITY;

COMMIT;