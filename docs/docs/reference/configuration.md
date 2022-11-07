# Configuration Reference

## Client Configuration

```yml
host: "localhost:8080"
```

| Field    | Type     | Description                            |
| -------- | -------- | -------------------------------------- |
| `host`   | `string` | Guardian server host (`<host>:<port>`) |

## Server Configuration

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
```


### Config

| Field                                        | Type                             | Description                                                             |
| -------------------------------------------- | -------------------------------- | ----------------------------------------------------------------------- |
| `port`                                       | `int`                            | Server Listen Port  (eg: `8080`)                                        |
| `encryption_secret_key`                      | `string`                         | Encryption secret key encrypt and decrypt credentials                   |
| `notifier`                                   | [`Object(NotifierConfig)`](#notifierconfig)  | Notification Configuration                                              |
| `log_level`                                  | `string`                         | Log level (default: `info`)                                             |
| `db`                                         | [`Object(DatabaseConfig)`](#databaseconfig)  | Database configuration                                                  |
| `authenticated_user_header_key`              | `string`                         | Header key name for authenticated user (eg: `X-Auth-Email`)             |
| `audit_log_trace_id_header_key`              | `string`                         | Header key name for trace id (eg: `X-Trace-Id`)                         |
| `jobs`                                       | [`Object(Jobs)`](#jobs)          | Server Jobs Configuration                                               |

### NotifierConfig

| Field          | Type                                                     | Description                                                             |
| -------------- | -------------------------------------------------------- | ----------------------------------------------------------------------- |
| `provider`     | `string`                                                 | Provider for notification (Only `slack` supported for now)              |
| `access_token` | `string`                                                 | Access Token for notification provider (eg: slack access token)         |
| `messages`     | [`Object(NotificationMessages)`](#notificationmessages)  | Message templates configuration                                         |

### NotificationMessages

| Field                    | Type      | Description                                                             |
| -------------------------| --------- | ----------------------------------------------------------------------- |
| `expiration_reminder`    | `string`  | Message template for expiration reminder                                |
| `appeal_approved`        | `string`  | Message template for appeal approved                                    |
| `appeal_rejected`        | `string`  | Message template for appeal rejected                                    |
| `access_revoked`         | `string`  | Message template for access revoked                                     |
| `approver_notification`  | `string`  | Message template for approver notification                              |
| `others_appeal_approved` | `string`  | Message template for other appeal approved                              |



### DatabaseConfig

| Field        | Type                             | Description                                                             |
| ------------ | -------------------------------- | ----------------------------------------------------------------------- |
| `host`       | `string`                         | Database host                                                           |
| `user`       | `string`                         | Database user                                                           |
| `password`   | `string`                         | Database password                                                       |
| `name`       | `string`                         | Database name                                                           |
| `port`       | `string`                         | Database port                                                           |
| `sslmode`    | `string`                         | Database sslmode                                                        |
| `log_level`  | `string`                         | Database log_level                                                      |

### Jobs

| Field                                | Type                              | Description                                                             |
| -------------------------------------| --------------------------------- | ----------------------------------------------------------------------- |
| `fetch_resources`                    | [`Object(JobConfig)`](#jobconfig) | When Enabled, the Guardian server fetches resources from the providers and updated the database.                                              |
| `revoke_expired_grants`              | [`Object(JobConfig)`](#jobconfig) | When Enabled, the Guardian server will revoke the user permissions for the resource                                        |
| `expiring_grant_notification`        | [`Object(JobConfig)`](#jobconfig) | When Enabled, the Guardian server will notify the user on the notifier (currently slack only) before the user appeal is about to expire. The user gets notified before 7 days, 3 days and 1 day of appeal expiry                                  |

### JobConfig

| Field      | Type                             | Description                                                             |
| -----------| -------------------------------- | ----------------------------------------------------------------------- |
| `enabled`  | `boolean`                        | Job Enabled                                                             |
| `interval` | `string`                         | Job interval ([cron format](https://crontab.guru), eg: `0 */2 * * *`)   |

## Using environment variables

All the configs can be passed as environment variables using underscore _ as the delimiter between nested keys. See the following examples

| Configuration key      | Environment variable |
| ---------------------- | -------------------- |
| `notifier.provider`    | `NOTIFIER_PROVIDER`  |

Set the env variable using export

```bash
export NOTIFIER_PROVIDER=slack
```