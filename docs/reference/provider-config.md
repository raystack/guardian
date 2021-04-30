# Provider configuration

A provider configuration is required when we want to register a provider instance to Guardian.

Field | Description | Required | Default value
------|-------------|----------|--------------
type | [Available provider types](#providers) | YES | -
urn | Provider instance identifier | YES | -
credentials | Credentials that will be used by Guardian to connect to the provider instance, check [Provider Credentials](#provider-credentials) | YES | -
appeal | Appeal options, check [Appeal Config](#appeal-config) | YES | -
resources | List of permission configurations for each available resource type, check [Resource Config](#resource-config) | YES | -

## Appeal Config

Field | Description | Required | Default value
------|-------------|----------|--------------
allow_permanent_access | Set this to `true` if you want to allow users to have permanent access to the resources | NO | `false`
allow_active_access_extension_in | Duration before the access expiration date when the user allowed to create appeal to the same resource (extend their current access). | NO | -

## Resource Config

Field | Description | Required | Default value
------|-------------|----------|--------------
type | Resource type in that provider, check  | YES | -
policy | Approval policy config want to be applied. Requires the `id` and the `version` of the policy | YES | -
roles | List of [Role config](#role-config)

## Role Config

Field | Description | Required | Default value
------|-------------|----------|--------------
id | Role ID. On the appeal creation, user is asked a role id for each resource access appeal | YES | -
name | Display name | NO | empty
permissions | Set of permissions want to given when the access granted | YES | -

## Providers

Here are the available providers in Guardian. Currently we only have Google BigQuery, but we will ad more soon.

| | Google BigQuery
|-|----------------
Provider type | `google_bigquery`
Credentials value | Base64 encrypted value of a service account key JSON
Available resource types | `dataset`, `table`

## Example

```yaml
type: google_bigquery
urn: gcp-project-id
credentials: credentials...
appeal:
  allow_permanent_access: true
  allow_active_access_extension_in: '7d'
resources:
  - type: dataset
    policy:
      id: bigquery_approval
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: roles/bigQuery.dataViewer
          - name: roles/customRole
            target: other-gcp-project-id
      - id: editor
        name: Editor
        permissions:
          - name: roles/bigQuery.dataEditor
  - type: table
    policy:
      id: bigquery_approval
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: roles/bigQuery.dataViewer
          - name: roles/customRole
            target: other-gcp-project-id
      - id: editor
        name: Editor
        permissions:
          - name: roles/bigQuery.dataEditor
```
