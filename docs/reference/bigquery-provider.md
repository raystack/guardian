# Google BigQuery Provider

## 1. Config

#### Example

```yaml
type: google_bigquery
urn: gcp-project-id
credentials: <base64 encoded service account key json>
appeal:
  allow_active_access_extension_in: '7d'
resources:
  - type: dataset
    policy:
      id: bq_dataset_approval
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: READER
          - name: roles/bigquery.jobUser
            target: targetted-gcp-project-id
      - id: editor
        name: Editor
        permissions:
          - name: WRITER
  - type: table
    policy:
      id: bq_table_approval
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: roles/bigquery.dataViewer
          - name: roles/bigquery.jobUser
            target: targetted-gcp-project-id
```

### `BigQueryCredentials`

`string`. BigQuery credentials is a **base64 encoded** service account key json.

### `BigQueryResourceType`

- `dataset`
- `table`

### `BigQueryResourcePermission`

Fields ||
-|-
`target` | `string` <br> Target GCP project ID. If this field presents, the specified role in the `name` field will get applied to this GCP project ID
`name` | `string` <br> Required. GCP role name <br><br> **Note:** for `dataset` resource type, we are using legacy roles (`READER`, `WRITER`, or `OWNER`). [Read more...](https://cloud.google.com/bigquery/docs/reference/rest/v2/datasets#:~:text=Required.%20An%20IAM,back%20as%20%22OWNER%22.)
