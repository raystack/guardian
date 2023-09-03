BEGIN;

ALTER TABLE activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE appeals ENABLE ROW LEVEL SECURITY;
ALTER TABLE approvals ENABLE ROW LEVEL SECURITY;
ALTER TABLE approvers ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE resources ENABLE ROW LEVEL SECURITY;


DROP POLICY IF EXISTS activities_isolation_policy ON activities;
CREATE POLICY activities_isolation_policy on activities USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS appeals_isolation_policy ON appeals;
CREATE POLICY appeals_isolation_policy on appeals USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS approvals_isolation_policy ON approvals;
CREATE POLICY approvals_isolation_policy on approvals USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS approvers_isolation_policy ON approvals;
CREATE POLICY approvers_isolation_policy on approvers USING (namespace_id = current_setting('app.current_tenant')::UUID);

-- DROP POLICY IF EXISTS audit_logs_isolation_policy ON audit_logs;
-- CREATE POLICY audit_logs_isolation_policy on audit_logs USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS grants_isolation_policy ON grants;
CREATE POLICY grants_isolation_policy on grants USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS policies_isolation_policy ON policies;
CREATE POLICY policies_isolation_policy on policies USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS providers_isolation_policy ON providers;
CREATE POLICY providers_isolation_policy on providers USING (namespace_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS resources_isolation_policy ON resources;
CREATE POLICY resources_isolation_policy on resources USING (namespace_id = current_setting('app.current_tenant')::UUID);

COMMIT;