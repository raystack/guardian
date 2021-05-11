# Managing Providers

## Registering Provider

Provider instances are registered to Guardian through configurations. Each provider configuration provides how Guardian can interact with the provider, configuring appeal’s approval policy, and role mapping.

Provider config example: 
```yaml
type: google_bigquery
urn: gcp-project-id
credentials: {base64 service account key json}
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

Check the [approval policy reference](../reference/provider-config.md) for more details

To create provider configuration, you can use this endpoint:

```
POST /providers
Accept: application/json

Request Body:
type: google_bigquery
urn: gcp-project-id
credentials: {base64 service account key json}
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

```
PUT /providers/:id
Accept: application/json

Request Body:
type: google_bigquery
urn: gcp-project-id
credentials: {base64 service account key json}
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