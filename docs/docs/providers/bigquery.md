# Big Query

BigQuery is an enterprise data warehouse tool for storing and querying massive datasets with super-fast SQL queries using the processing power of Google's infrastructure.

BigQuery has **datasets** that each one contains multiple **tables** which are used for storing the data. User also can run queries to read or even transform those data.
### BigQuery Resources

- **Dataset**: Datasets are top-level logical containers that are used to organize and control access to your BigQuery resources tables and views. Datasets are similar to schemas in other database systems.
- **Project**: Every dataset is associated with a project. To use Google Cloud, you must create at least one project. Projects form the basis for creating, enabling and using all Google Cloud services. A project can hold multiple datasets, and datasets with different locations can exist in the same project.
- **Folder**: Folders are an additional grouping mechanism above projects. Projects and folders inside a folder automatically inherit the access policies of their parent folder. Folders can be used to model different legal entities, departments, and teams within a company.
- **Organization**: The Organization resource represents an organization (for example, a company) and is the root node in the Google Cloud resource hierarchy. Using an Organization resource allows administrators to centrally control your BigQuery resources, rather than individual users controlling the resources they create.
- **Table**: A BigQuery table contains individual records organized in rows. Each record is composed of columns (also called fields).
Every table is defined by a schema that describes the column names, data types, and other information. You can specify the schema of a table when it is created, or you can create a table without a schema and declare the schema in the query job or load job that first populates it with data.

### BigQuery Users

BigQuery allows users, groups, and service accounts allowed to access the tables, views, and table data in a specific dataset. Currently, Guardian only supports **`user`** and **`service account`** as account types.

### Prerequisites

If a user/administrator wants to control access to a dataset or a table, the user must have sufficient permissions for the same. With these permissions, the resource owner can grant and revoke other users/service accounts with selective access to these resources.

For registering BigQuery as a provider on Guardian, users must have a service account with IAM role: **`roles/bigquery.dataOwner`** at the project level.



### Authentication

Guardian requires **service account key** and the **resource name** of an administrator user in BigQuery. The Service Account key should be base64 encoded value.

```yaml
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: projects/gcp-project-id
```


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
    filter: $resource.name endsWith transtaction
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
| `resource_name` | `string` This field contains the Project ID of the project containing the resources.<br/> Example: `projects/my-project-id` |
| `service_account_key` | `string` Service account key JSON that has [prerequisites permissions](#prerequisites). On provider creation, the value should be an base64 encoded JSON key. |

### `BigQueryResourceType`

- `dataset`
- `table`

### `BigQueryResourcePermission`

A Google Cloud predefined role name. 

For `dataset` resource type, we are using legacy roles. For more details [read here](https://cloud.google.com/bigquery/docs/access-control-basic-roles)
- `READER`
- `WRITER`
- `OWNER` 
