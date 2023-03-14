# Setup server

Guardian binary contains both the CLI client and the server itself. Each has it's own configuration in order to run. Server configuration contains information such as database credentials, log severity, etc. while CLI client configuration only has configuration about which server to connect.

## Server

#### Pre-requisites

- Postgres
- Slackbot access token for notification (optional)

#### Initialization

Create a config.yaml file (`touch config.yaml`) in the root folder of guardian project or [use `--config` flag](#using---config-flag) to customize to config file location, or you can also [use environment variables](#using-environment-variables) to provide the server config. Setup up a database in postgres and provide the details in the DB field as given in the example below. For the purpose of this tutorial, we'll assume that the username is `your_user`, database name is `guardian`, host and port are `localhost` and `5432`.

> If you're new to YAML and want to learn more, see [Learn YAML in Y minutes.](https://learnxinyminutes.com/docs/yaml/)

Following is a sample server configuration yaml:

```yaml
PORT: 3000
LOG:
  LEVEL: info # debug|info|warning|error|fatal - default: info
DB:
  HOST: localhost
  USER: your_user
  PASSWORD: your_password
  NAME: guardian
  PORT: 5432
NOTIFIER:
  PROVIDER: slack
  ACCESS_TOKEN: <slack-access-token>
  ...
AUTHENTICATED_USER_HEADER_KEY: X-Auth-Email
JOBS:
  FETCH_RESOURCES:
    ENABLED: true
    INTERVAL: '0 */2 * * *'  #"At minute 0 past every 2nd hour"
  REVOKE_EXPIRED_ACCESS:
    ENABLED: true
    INTERVAL: '*/20 * * * *'  #“At every 20th minute"
  EXPIRING_ACCESS_NOTIFICATION:
    ENABLED: true
    INTERVAL: '0 9 * * *' #"At minute 0 past hour 9"
```

<!-- TODO: add documentation for notifier messsages -->

#### Starting the server

Database migration is required during the first server initialization. In addition, re-running the migration command might be needed in a new release to apply the new schema changes (if any). It's safer to always re-run the migration script before deploying/starting a new release.

To initialize the database schema, Run Migrations with the following command:

```sh
$ guardian server migrate
```

To run the Guardian server use command:

```sh
$ guardian server start
```

##### Using `--config` flag

```sh
$ guardian server migrate --config=<path-to-file>
```

```sh
$ guardian server start --config=<path-to-file>
```

##### Using environment variables

All the configs can be passed as environment variables using underscore `_` as the delimiter between nested keys. See the following examples

```yaml
PORT: 8080
DB:
  HOST: localhost
  USER: test
```

Here is the corresponding environment variable for the above

| Configuration key | Environment variable |
| ----------------- | -------------------- |
| PORT              | PORT                 |
| DB.HOST           | DB_HOST              |
| DB.USER           | DB_USER              |

Set the env variable using export

```
$ export PORT=8080
```

---

## CLI Client

### Initialization

Guardian CLI supports CLI client to communicate with a Guardian server. To initialize the client configuration, run the following command:

```sh
$ guardian config init
```

A yaml file will be created in the `~/.config/goto/guardian.yaml` directory. Open this file to configure the host for Guardian server as in the example below:

```yaml
host: "localhost:8080"
```
