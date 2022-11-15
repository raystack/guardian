# Data Catalog Policy tag

Policy tags enable you to control who can view sensitive columns in BigQuery tables. In Data Catalog, you can add or remove policy tags to columns directly on the table entry details page.

### Policy tag Resources

- **Tag**: Users that need access to columns protected with policy tags need the Fine-Grained Reader role. This role is assigned individually on every policy tag.

### BigQuery Users

BigQuery allows users, groups, and service accounts allowed to access the Fine-Grained Reader role on policy-tag. Currently, Guardian only supports **`user`** and **`service account`** as account types.

### Prerequisites

If a user/administrator wants to access to columns protected with policy tags need the Fine-Grained Reader role, the user must have `Fine-Grained Reader` permissions for the policy-tag associated with column.

For registering Data Catalog Policy tag as a provider on Guardian, users must have a service account with IAM roles: **`roles/bigquery.dataOwner, roles/datacatalog.categoryAdmin`** at the project level.



### Authentication

Guardian requires **service account key** and the **resource name** of an administrator user in BigQuery. The Service Account key should be base64 encoded value.

```yaml
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: projects/gcp-project-id/locations/taxonomy-location
```


## Access Management

Access can be given at the policy-tag level as those allowed to be managed through these DataCatalog APIs:

## Config

#### YAML Representation

```yaml
type: policy_tag
urn: my-policy_tag
allowed_account_types:
  - user
  - serviceAccount
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: projects/gcp-project-id/locations/us
appeal:
  allow_active_access_extension_in: "7d"
resources:
  - type: tag
    policy:
      id: my_policy
      version: 1
    roles:
      - id: fineGrainReader
        name: Fine Grain Reader
        permissions:
          - roles/datacatalog.categoryFineGrainedReader
```

### `AccountType`

- `user`
- `serviceAccount`

### `Credentials`

| Fields | |
| :--- | :--- |
| `resource_name` | `string` This field contains the Project ID of the project containing the resources.<br/> Example: `projects/my-project-id` |
| `service_account_key` | `string` Service account key JSON that has [prerequisites permissions](#prerequisites). On provider creation, the value should be an base64 encoded JSON key. |

### `ResourceType`

- `tag`

### `PolicyTagResourcePermission`

A Google Cloud predefined role name. 

For `tag` resource type and column access we are using following permission
- `roles/datacatalog.categoryFineGrainedReader`