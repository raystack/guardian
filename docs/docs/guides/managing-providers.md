# Managing providers

## Registering Provider

Provider instances are registered to Guardian through configurations. Each provider configuration provides how Guardian can interact with the provider, configuring appeal’s approval policy, and role mapping.

Provider config example:

```yaml
type: google_bigquery
urn: bq-resource-urn
credentials: 
  - service_account_key: {base64 encoded service account key json}
  - resource_name: projects/gcp-project-id
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: dataset
    policy:
      id: bigquery_dataset
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: READER
          - name: roles/customRole
            target: other-gcp-project-id
      - id: editor
        name: Editor
        permissions:
          - name: WRITER
  - type: table
    policy:
      id: bigquery_table
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: roles/bigQuery.dataViewer
      - id: editor
        name: Editor
        permissions:
          - name: roles/bigQuery.dataEditor
```

Check the [approval policy reference](https://github.com/odpf/guardian/tree/9710a699aed45f07a88283bef5f80e60db38d825/docs/reference/provider.md) for more details

To create provider configuration, you can use this endpoint:

```text
POST /providers
Accept: application/json

Request Body:
type: google_bigquery
urn: bg-resource-urn
credentials: 
  - service_account_key: {base64 encoded service account key json}
  - resource_name: projects/gcp-project-id
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: dataset
    policy:
      id: bigquery_dataset
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: READER
          - name: roles/customRole
            target: other-gcp-project-id
      - id: editor
        name: Editor
        permissions:
          - name: WRITER
  - type: table
    policy:
      id: bigquery_table
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: roles/bigQuery.dataViewer
      - id: editor
        name: Editor
        permissions:
          - name: roles/bigQuery.dataEditor

Response:
{
  "id": 1,
  "type": "google_bigquery",
  "urn": "pilotdata-integration",
  "config": {
    "type": "google_bigquery",
    "urn": "pilotdata-integration",
    "labels": null,
    "appeal": {
      "allow_permanent_access": false,
      "allow_active_access_extension_in": "7d"
    },
    "resources": [
      {
        "type": "dataset",
        "policy": {
          "id": "bigquery_dataset",
          "version": 1
        },
        "roles": [
          {
            "id": "viewer",
            "name": "Viewer",
            "permissions": [
              {
                "name": "READER"
              },
              {
                "name": "roles/customRole",
                "target": "other-gcp-project-id"
              }
            ]
          },
          {
            "id": "editor",
            "name": "Editor",
            "permissions": [
              {
                "name": "WRITER"
              }
            ]
          }
        ]
      },
      {
        "type": "table",
        "policy": {
          "id": "bigquery_table",
          "version": 1
        },
        "roles": [
          {
            "id": "viewer",
            "name": "Viewer",
            "permissions": [
              {
                "name": "roles/bigQuery.dataViewer"
              }
            ]
          },
          {
            "id": "editor",
            "name": "Editor",
            "permissions": [
              {
                "name": "roles/bigQuery.dataEditor"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

## Updating Provider Config

To update a provider configuration, you can use this endpoint:

```text
PUT /providers/:id
Accept: application/json

Request Body:
type: google_bigquery
urn: bg-resource-urn
credentials: 
  - service_account_key: {base64 encoded service account key json}
  - resource_name: projects/gcp-project-id
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: dataset
    policy:
      id: bigquery_dataset
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: READER
          - name: roles/customRole
            target: other-gcp-project-id
      - id: editor
        name: Editor
        permissions:
          - name: WRITER
  - type: table
    policy:
      id: bigquery_table
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: roles/bigQuery.dataViewer
      - id: editor
        name: Editor
        permissions:
          - name: roles/bigQuery.dataEditor

Response:
{
  "id": 1,
  "type": "google_bigquery",
  "urn": "pilotdata-integration",
  "config": {
    "type": "google_bigquery",
    "urn": "pilotdata-integration",
    "labels": null,
    "appeal": {
      "allow_permanent_access": false,
      "allow_active_access_extension_in": "7d"
    },
    "resources": [
      {
        "type": "dataset",
        "policy": {
          "id": "bigquery_dataset",
          "version": 1
        },
        "roles": [
          {
            "id": "viewer",
            "name": "Viewer",
            "permissions": [
              {
                "name": "READER"
              },
              {
                "name": "roles/customRole",
                "target": "other-gcp-project-id"
              }
            ]
          },
          {
            "id": "editor",
            "name": "Editor",
            "permissions": [
              {
                "name": "WRITER"
              }
            ]
          }
        ]
      },
      {
        "type": "table",
        "policy": {
          "id": "bigquery_table",
          "version": 1
        },
        "roles": [
          {
            "id": "viewer",
            "name": "Viewer",
            "permissions": [
              {
                "name": "roles/bigQuery.dataViewer"
              }
            ]
          },
          {
            "id": "editor",
            "name": "Editor",
            "permissions": [
              {
                "name": "roles/bigQuery.dataEditor"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

