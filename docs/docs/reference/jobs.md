# Jobs

You can run jobs using `guardian` cli command to perform one time actions. You can also run them periodically using cronjob through [helm chart](../guides/deployment.md#use-the-helm-chart). The following jobs are available:

- `fetch_resources`
- `expiring_grant_notification`
- `revoke_expired_grants`
- `revoke_grants_by_user_criteria`
- `grant_dormancy_check`

Reference: [Jobs](https://github.com/goto/guardian/blob/main/jobs/jobs.go)

| Field                            | Description                                                                                                                                                                                                                 | 
|----------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| 
| `FETCH_RESOURCES`                | When Enabled, the Guardian server fetches resources from the providers and updated the database.                                                                                                                            | 
| `REVOKE_EXPIRED_GRANTS`          | When Enabled, the Guardian server will revoke the user permissions for the resource                                                                                                                                         |
| `EXPIRING_GRANT_NOTIFICATION`    | When Enabled, the Guardian server will notify the user on the notifier (currently `slack` only) before the user appeal is about to expire.<br/><br/>The user gets notified before 7 days, 3 days and 1 day of appeal expiry |
| `REVOKE_GRANTS_BY_USER_CRITERIA` | When Enabled, the Guardian server will revoke the user permissions for the resource based on the criteria provided in the `user_criteria` field.                                                                            |
| `GRANT_DORMANCY_CHECK`           | When Enabled, the Guardian server will check for the dormancy of the grant and will update the expiry date of the grant based on `retain_grant_for` field                                                                   |                                                                                           
