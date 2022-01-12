# Big Query

BigQuery is an enterprise data warehouse tool for storing and querying massive datasets with super-fast SQL queries using the processing power of Google's infrastructure.

BigQuery has **datasets** that each one contains multiple **tables** which are used for storing the data. User also can run queries to read or even transform those data.

## Prerequisites

1. A service account with `roles/bigquery.dataOwner` role at the project level

## Access Management

Access can be given at the dataset level or table level as those allowed to be managed through these BigQuery APIs:
- [Dataset Access Control](https://cloud.google.com/bigquery/docs/dataset-access-controls)
- [Table Access Control](https://cloud.google.com/bigquery/docs/table-access-controls-intro)

## Config

#### YAML Representation

```yaml
type: bigquery
urn: my-bigquery
allowed_account_types:
  - user
  - serviceAccount
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: projects/gcp-project-id
appeal:
  allow_active_access_extension_in: "7d"
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
          - roles/bigquery.jobUser
```

### `BigQueryAccountType`

- `user`
- `serviceAccount`

### `BigQueryCredentials`

| Fields | |
| :--- | :--- |
| `resource_name` | `string` GCP Project ID in resource name format. Example: `projects/my-project-id` |
| `service_account_key` | `string` Service account key JSON that has [prerequisites permissions](#prerequisites). On provider creation, the value should be an base64 encoded JSON key. |

### `BigQueryResourceType`

- `dataset`
- `table`

### `BigQueryResourcePermission`

A Google Cloud predefined role name. For `dataset` resource type, we are using legacy roles \(`READER`, `WRITER`, or `OWNER`\). [Read more...](https://cloud.google.com/bigquery/docs/reference/rest/v2/datasets#:~:text=Required.%20An%20IAM,back%20as%20%22OWNER%22.)
