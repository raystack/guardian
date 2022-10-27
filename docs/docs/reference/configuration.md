import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Configuration

<Tabs groupId="api">
  <TabItem value="client" label="Client" default>

## Client Configuration Reference

```yml
host: localhost:8080
```

| Key      | Description                            |
| -------- | -------------------------------------- |
| `host`   | Guardian Server Host (`<host>:<port>`) |

  </TabItem>
  <TabItem value="server" label="Server">

## Server Configuration Reference

```yml
port: 8080
encryption_secret_key: "<secret-key>"
notifier:
    provider: "slack"
    access_token: "<slack-access-token>"
    messages:
        expiration_reminder: "Your access to {{.resource_name}} with role {{.role}} will expire at {{.expiration_date}}. Extend the access if it's still needed"
        appeal_approved: "Your appeal to {{.resource_name}} with role {{.role}} has been approved"
        appeal_rejected: "Your appeal to {{.resource_name}} with role {{.role}} has been rejected"
        access_revoked: "Your access to {{.resource_name}}} with role {{.role}} has been revoked"
        approver_notification: "You have an appeal created by {{.requestor}} requesting access to {{.resource_name}} with role {{.role}}. Appeal ID: {{.appeal_id}}"
        others_appeal_approved: "Your appeal to {{.resource_name}} with role {{.role}} created by {{.requestor}} has been approved"
log_level: "info"
db:
    host: "localhost"
    user: "postgres"
    password: ""
    name: "postgres"
    port: "5432"
    sslmode: "disable"
    log_level: "info"
authenticated_user_header_key: "X-Auth-Email"
audit_log_trace_id_header_key: "X-Trace-Id"
jobs:
    fetch_resources:
        enabled: true
        interval: "0 */2 * * *"
    revoke_expired_grants:
        enabled: true
        interval: "*/20 * * * *"
    expiring_grant_notification:
        enabled: true
        interval: "0 9 * * *"
    revoke_expired_access:
        enabled: true
        interval: ""
    expiring_access_notification:
        enabled: true
        interval: ""
```

| Field                                        | Type      | Description                                                             |
| -------------------------------------------- | --------- | ----------------------------------------------------------------------- |
| `port`                                       | `int`     | Server Listen Port  (eg: `8080`)                                        |
| `encryption_secret_key`                      | `string`  | Encryption secret key encrypt and decrypt credentials                   |
| `notifier.provider`                          | `string`  | Provider for notification (Only `slack` supported for now)              |
| `notifier.access_token`                      | `string`  | Access Token for notification provider (eg: slack access token)         |
| `notifier.messages.expiration_reminder`      | `string`  | Message template for expiration reminder                                |
| `notifier.messages.appeal_approved`          | `string`  | Message template for appeal approved                                    |
| `notifier.messages.appeal_rejected`          | `string`  | Message template for appeal rejected                                    |
| `notifier.messages.access_revoked`           | `string`  | Message template for access revoked                                     |
| `notifier.messages.approver_notification`    | `string`  | Message template for approver notification                              |
| `notifier.messages.others_appeal_approved`   | `string`  | Message template for other appeal approved                              |
| `log_level`                                  | `string`  | Log level (default: `info`)                                             |
| `db.host`                                    | `string`  | Database host                                                           |
| `db.user`                                    | `string`  | Database user                                                           |
| `db.password`                                | `string`  | Database password                                                       |
| `db.name`                                    | `string`  | Database name                                                           |
| `db.port`                                    | `string`  | Database port                                                           |
| `db.sslmode`                                 | `string`  | Database sslmode                                                        |
| `db.log_level`                               | `string`  | Database log_level                                                      |
| `authenticated_user_header_key`              | `string`  | Header key name for authenticated user (eg: `X-Auth-Email`)             |
| `audit_log_trace_id_header_key`              | `string`  | Header key name for trace id (eg: `X-Trace-Id`)                         |
| `jobs.fetch_resources.enabled`               | `boolean` | Enable fetch resources job                                              |
| `jobs.fetch_resources.interval`              | `string`  | Fetch resources job interval ([cron format](https://crontab.guru), eg: `0 */2 * * *`)           |
| `jobs.revoke_expired_grants.enabled`         | `boolean` | Enable revoke expired grants job                                        |
| `jobs.revoke_expired_grants.interval`        | `string`  | Revoke expired grants Job interval ([cron format](https://crontab.guru), eg: `*/20 * * * *`)    |
| `jobs.expiring_grant_notification.enabled`   | `boolean` | Enable expiring grant notification job                                  |
| `jobs.expiring_grant_notification.interval`  | `string`  | Expiring grant notification job interval ([cron format](https://crontab.guru), eg: `0 9 * * *`) |
| `jobs.revoke_expired_access.enabled`         | `boolean` | Enable Revoke expired access                                            |
| `jobs.revoke_expired_access.interval`        | `string`  | Revoke expired access job interval ([cron format](https://crontab.guru), eg: `*/20 * * * *`)     |
| `jobs.expiring_access_notification.enabled`  | `boolean` | Enable expiring access notification job                                 |
| `jobs.expiring_access_notification.interval` | `string`  | Expiring access notification job interval ([cron format](https://crontab.guru), eg: `0 9 * * *`) |

  </TabItem>
</Tabs>




## Using environment variables

All the configs can be passed as environment variables using underscore _ as the delimiter between nested keys. See the following examples

| Configuration key      | Environment variable |
| ---------------------- | -------------------- |
| `notifier.provider`    | `NOTIFIER_PROVIDER`  |

Set the env variable using export

```bash
export NOTIFIER_PROVIDER=slack
```