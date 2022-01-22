# Managing Providers

Provider manages roles, resources, provider credentials and also points each resource type to a considered policy.

## Registering Providers

Providers can be created in the following ways:
1. [Using `guardian provider create` CLI command](#1-register-a-provider-using-cli)
2. [Calling to `POST /api/v1beta1/providers` API](#2-register-a-provider-using-http-api)

Once a provider config is registered, Guardian will immediately fetch the resources and store it in the database.

#### Example
```yaml
# provider.yaml
type: bigquery
urn: gcp-project-id-bigquery
allowed_account_types:
  - user
  - serviceAccount
credentials:
  service_account_key: {{base64 encoded service account key json}}
  resource_name: projects/gcp-project-id
appeal:
  allow_permanent_access: false
  allow_active_access_extension_in: "168h"
resources:
  - type: dataset
    policy:
      id: my_policy
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - READER
      - id: editor
        name: Editor
        permissions:
          - WRITER
  - type: table
    policy:
      id: my_policy
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - roles/bigquery.dataViewer
      - id: editor
        name: Editor
        permissions:
          - roles/bigquery.dataEditor
```

Check [provider reference](../reference/provider.md) for more details.

### 1. Register a Provider using CLI
```console
$ guardian provider create --file=provider.yaml
```

### 2. Register a Provider using HTTP API
```console
$ curl --request POST '{{HOST}}/api/v1beta1/providers' \
--header 'Content-Type: application/json' \
--data-raw '{
  "type": "bigquery",
  "urn": "gcp-project-id-bigquery",
  "allowed_account_types": [
    "user",
    "serviceAccount"
  ],
  "credentials": {
    "service_account_key": "{{base64 encoded service account key json}}",
    "resource_name": "projects/gcp-project-id"
  },
  "appeal": {
    "allow_permanent_access": false,
    "allow_active_access_extension_in": "168h"
  },
  "resources": [
    {
      "type": "dataset",
      "policy": {
        "id": "my_policy",
        "version": 1
      },
      "roles": [
        {
          "id": "viewer",
          "name": "Viewer",
          "permissions": [
            "READER"
          ]
        },
        {
          "id": "editor",
          "name": "Editor",
          "permissions": [
            "WRITER"
          ]
        }
      ]
    },
    {
      "type": "table",
      "policy": {
        "id": "my_policy",
        "version": 1
      },
      "roles": [
        {
          "id": "viewer",
          "name": "Viewer",
          "permissions": [
            "roles/bigquery.dataViewer"
          ]
        },
        {
          "id": "editor",
          "name": "Editor",
          "permissions": [
            "roles/bigquery.dataEditor"
          ]
        }
      ]
    }
  ]
}'
```

## Updating Providers

Providers can be updated in the following ways:
1. [Using `guardian provider edit` CLI command](#1-update-a-provider-using-cli)
2. [Calling to `PUT /api/v1beta1/providers/:id` API](#2-update-a-provider-using-http-api)

### 1. Update a Provider using CLI
```console
$ guardian provider edit {{provider_id}} --file=provider.yaml
```

### 2. Update a Provider using HTTP API
```console
$ curl --request PUT '{{HOST}}/api/v1beta1/providers/{{provider_id}}' \
--header 'Content-Type: application/json' \
--data-raw '{
  "allowed_account_types": [
    "user"
  ]
}'
```

## Listing Providers

Listing providers can be done in the following ways:
1. [Using `guardian provider list` CLI command](#1-list-providers-using-cli)
2. [Calling to `GET /api/v1beta1/providers` API](#2-list-providers-using-http-api)

### 1. List Providers using CLI
```console
$ guardian provider list --output=yaml
```

### 2. List Providers using HTTP API
```console
$ curl --request GET '{{HOST}}/api/v1beta1/providers'
```

## Viewing Providers

Viewing a provider can be done in the following ways:

1. [Using `guardian provider view` CLI command](#1-view-a-provider-using-cli)
2. [Calling to `GET /api/v1beta1/providers/:id` API](#2-view-a-provider-using-http-api)

### 1. View a Provider using CLI
```console
$ guardian provider view {{provider_id}}
```

### 2. View a Provider using HTTP API
```console
$ curl --request GET '{{HOST}}/api/v1beta1/providers/{{provider_id}}'
```

## Listing Roles for a Resource Type

Listing roles can be done in the following ways:

1. [Calling to `GET /api/v1beta1/providers/:id/resources/:resource_type/roles` API](#1-list-roles-using-http-api)

### 1. List Roles using HTTP API
```console
$ curl --request GET '{{HOST}}/api/v1beta1/providers/{{provider_id}}/resources/{{resource_type}}/roles'
```
