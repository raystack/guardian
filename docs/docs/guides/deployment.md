# Server Installation

There are several approaches to setup Guardian Server

1. [Using the CLI](#using-the-cli)
1. [Using the Docker](#use-the-docker-image)
1. [Using the Helm Chart](#use-the-helm-chart)

## General pre-requisites

- PostgreSQL (version 13 or above)
- Slackbot access token for notification (optional)

## Using the CLI

### Pre-requisites for CLI

- [Create guardian config file](/docs/tour/configuration#initialization)

To run the Guardian server use command:

```sh
$ guardian server start -c <path-to-config>
```

## Use the Docker

To run the Guardian server using Docker, you need to have Docker installed on your system. You can find the installation instructions [here](https://docs.docker.com/get-docker/).

You can choose to set the configuration using environment variables or a config file. The environment variables will override the config file.

### Using environment variables

All the configs can be passed as environment variables using underscore `_` as the delimiter between nested keys. See the following examples

See [configuration reference](/docs/reference/configuration) for the list of all the configuration keys.

```sh title=".env"
PORT=8080
AUTHENTICATED_USER_HEADER_KEY=X-Auth-Email
DB_HOST=<db-host>
DB_NAME=<db-name>
DB_PASSWORD=<db-password>
DB_PORT=<db-port>
DB_USER=<db-user>
ENCRYPTION_SECRET_KEY=<secure-encription-key>
JOBS_EXPIRING_GRANT_NOTIFICATION_ENABLED=true
JOBS_REVOKE_EXPIRED_GRANTS_ENABLED=true
JOBS_REVOKE_EXPIRED_GRANTS_INTERVAL="*/20 * * * *"
NOTIFIER_ACCESS_TOKEN=<slack-access-token>
NOTIFIER_PROVIDER=slack
```

Run the following command to start the server

```sh
$ docker run -d \
    --restart=always \
    -p 8080:8080 \
    --env-file .env \
    --name guardian-server \
    raystack/guardian:<version> \
    server start
```

### Using config file

```yaml title="config.yaml"
port: 8080
encryption_secret_key: "<secret-key>"
db:
  host: "<db-host>"
  user: "<db-user>"
  password: "<db-password>"
  name: "<db-name>"
  port: "<db-port>"
authenticated_user_header_key: "X-Auth-Email"
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
notifier:
  provider: "slack"
  access_token: "<slack-access-token>"
```

Run the following command to start the server

```sh
$ docker run -d \
    --restart=always \
    -p 8080:8080 \
    -v $(pwd)/config.yaml:/config.yaml \
    --name guardian-server \
    raystack/guardian:<version> \
    server start -c /config.yaml
```

## Use the Helm chart

### Pre-requisites for Helm chart

Guardian can be installed in Kubernetes using the Helm chart from https://github.com/raystack/charts.

Ensure that the following requirements are met:

- Kubernetes 1.14+
- Helm version 3.x is [installed](https://helm.sh/docs/intro/install/)

### Add Raystack Helm repository

Add Raystack chart repository to Helm:

```
helm repo add raystack https://guardian.vercel.app/charts/
```

You can update the chart repository by running:

```
helm repo update
```

### Setup helm values

The following table lists the configurable parameters of the Guardian chart and their default values.

See full helm values guide [here](https://github.com/raystack/charts/tree/main/stable/guardian#values).

```yaml title="values.yaml"
app:
  ## Value to fully override guardian.name template
  nameOverride: ""
  ## Value to fully override guardian.fullname template
  fullnameOverride: ""

  image:
    repository: raystack/guardian
    pullPolicy: Always
    tag: latest
  container:
    args:
      - server
      - start
    livenessProbe:
      httpGet:
        path: /ping
        port: tcp
    readinessProbe:
      httpGet:
        path: /ping
        port: tcp

  migration:
    enabled: true
    args:
      - server
      - migrate

  service:
    annotations:
      projectcontour.io/upstream-protocol.h2c: tcp

  cron:
    enabled: true
    jobs:
      - name: "fetch-resources"
        schedule: "0 */2 * * *"
        restartPolicy: Never
        command: []
        args:
          - job
          - run
          - fetch_resources
      - name: "appeal-expiration-reminder"
        schedule: "0 9 * * *"
        restartPolicy: Never
        command: []
        args:
          - job
          - run
          - appeal_expiration_reminder
      - name: "appeal-expiration-revocation"
        schedule: "*/20 * * * *"
        restartPolicy: Never
        command: []
        args:
          - job
          - run
          - appeal_expiration_revocation

  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: contour
    hosts:
      - host: guardian.example.com
        paths:
          - path: /
            pathType: ImplementationSpecific
            backend:
              service:
                # name: backend_01
                port:
                  number: 80

  config:
    LOG_LEVEL: info
    AUTHENTICATED_USER_HEADER_KEY: x-authenticated-user-email
    NOTIFIER_PROVIDER: slack
    JOBS_FETCH_RESOURCES_ENABLED: false
    JOBS_FETCH_RESOURCES_INTERVAL: "0 */2 * * *"
    JOBS_REVOKE_EXPIRED_ACCESS_ENABLED: false
    JOBS_REVOKE_EXPIRED_ACCESS_INTERVAL: "*/20 * * * *"
    JOBS_EXPIRING_ACCESS_NOTIFICATION_ENABLED: false
    JOBS_EXPIRING_ACCESS_NOTIFICATION_INTERVAL: "0 9 * * *"

  secretConfig:
    ENCRYPTION_SECRET_KEY:
    NOTIFIER_ACCESS_TOKEN:
    DB_HOST: localhost
    DB_PORT:
    DB_NAME: guardian
    DB_USER: guardian
    DB_PASSWORD:
```

And install it with the helm command line along with the values file:

```sh
$ helm install my-release -f values.yaml raystack/guardian
```
