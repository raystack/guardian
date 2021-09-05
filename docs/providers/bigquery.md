# Big Query

## BigQuery

BigQuery is an enterprise data warehouse tool for storing and querying massive datasets with super-fast SQL queries using the processing power of Google's infrastructure.

BigQuery has **datasets** that each one contains multiple **tables** which are used for storing the data. User also can run queries to read or even transform those data.

## Authentication

A service account is required for Guardian to access BigQuery/GCP for fetching resources as well as managing permissions. The service account key is included during the provider registration. Read more on [Registering Provider](../guides/managing-providers.md#registering-provider).

## Access Management

In Guardian, user access can be given at the dataset or table level. To give more flexibility in terms of using BigQuery, Guardian also allow user to get permission at the GCP IAM level.

### BigQuery Resources

* [Dataset Access Control](https://cloud.google.com/bigquery/docs/dataset-access-controls)
* [Table Access Control](https://cloud.google.com/bigquery/docs/table-access-controls-intro)

### GCP IAM

* [IAM Permission](https://cloud.google.com/iam/docs/granting-changing-revoking-access)



## 1. Config

#### Example

```yaml
type: google_bigquery
urn: bg-resource-urn
credentials: 
  - service_account_key: <base64 encoded service account key json>
  - resource_name: projects/gcp-project-id
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

`string`. BigQuery credentials is a struct containing **base64 encoded** service account key json and resource name which should be having a **projects/** prefix followed by the gcp project id.

### `BigQueryResourceType`

* `dataset`
* `table`

### `BigQueryResourcePermission`

| Fields |  |
| :--- | :--- |
| `target` | `string`   Target GCP project ID. If this field presents, the specified role in the `name` field will get applied to this GCP project ID |
| `name` | `string`   Required. GCP role name    **Note:** for `dataset` resource type, we are using legacy roles \(`READER`, `WRITER`, or `OWNER`\). [Read more...](https://cloud.google.com/bigquery/docs/reference/rest/v2/datasets#:~:text=Required.%20An%20IAM,back%20as%20%22OWNER%22.) |

