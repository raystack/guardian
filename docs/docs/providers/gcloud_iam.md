# GCP
GCloud IAM provides a simple and consistent access control interface for all Google Cloud services. The Cloud IAM lets administrators authorize who can take action on specific resources, giving you full control and visibility to manage Google Cloud resources centrally.
### GCloud IAM Resources

- **Organization**: The organization resource is the hierarchical ancestor of folder and project resources. The IAM access control policies applied to the organization resource apply throughout the hierarchy of all resources in the organization. Organizational units help you manage users and apply common configurations or policies to users within the organization. 
- **Folder**: Folders are nodes in the Google Cloud resource hierarchy and can contain projects, other folders, or a combination of both. They can be seen as sub-organizations within the organization’s resources. Folders help you manage Google Cloud projects and apply common configurations or policies to projects.
- **Project**: The project resource is the base-level organizing entity. Organization and folder resources may contain multiple projects.  Projects play a crucial role in managing APIs, billing, and managing access to resources. In the context of identity management, projects are relevant because they are the containers for service accounts.

### GCloud IAM Users

The IAM allows user(Google account), Service account, Google group, Google Workspace account and Cloud identity domain. Currently, Guardian only supports **`user`** and **`service account`** as account types.

### Prerequisites

If a user/administrator wants to control access to an organization or a project, the user must have sufficient permissions for the same. With these permissions, the resource owner can grant and revoke other users/service accounts with selective access to these resources.

For registering GCloud IAM as a provider on Guardian, users must have a service account with IAM role: **`roles/iam.securityAdmin`** at the project/organization level.

### Authentication

Guardian requires **service account key** and the **resource name** of an administrator user in GCloud IAM. The Service Account key should be base64 encoded value.

```yaml
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: projects/gcp-project-id
```
## Access Management

Google Cloud IAM can be registered into Guardian in organization or project level by specifying the `credentials.resource_name` accordingly, `organizations/org-id` for an organization, and `projects/project-id` for a project. A provider instance, either it is an organzation or project, is considered as Guardian resource. Google Cloud predefined and custom roles can be selected as a role during appeal creation.

## Config

#### YAML Representation

```yaml
# project
type: gcloud_iam
urn: my-iam
allowed_account_types:
  - user
  - serviceAccount
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: projects/gcp-project-id
appeal:
  allow_active_access_extension_in: "7d"
resources:
  - type: project
    policy:
      id: my_policy
      version: 1
    roles:
      - id: role-1
        name: BigQuery
        permissions:
          - roles/bigquery.admin
          - roles/bigquery.dataEditor
          - roles/bigquery.dataOwner
      - id: role-2
        name: Custom
        permissions:
          - projects/integration/roles/project.iamManager
      - id: role-3
        name: Api gateway
        permissions:
          - roles/apigateway.admin
          - roles/apigateway.viewer
```

```yaml
# organization
type: gcloud_iam
urn: my-iam
allowed_account_types:
  - user
  - serviceAccount
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: organizations/gcp-org-id
appeal:
  allow_active_access_extension_in: "7d"
resources:
  - type: organization
    policy:
      id: my_policy
      version: 1
    roles:
      - id: role-1
        name: BigQuery
        permissions:
          - roles/bigquery.admin
          - roles/bigquery.dataEditor
          - roles/bigquery.dataOwner
```

### `GCloudIAMAccountType`

- `user`
- `serviceAccount`

### `GCloudIAMCredentials`

| Fields                |                                                                                                                                                               |
| :-------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `resource_name`       | `string` GCP Project ID in resource name format. Example: `projects/my-project-id`, `organizations/my-org-id`                                                 |
| `service_account_key` | `string` Service account key JSON that has [prerequisites permissions](#prerequisites). On provider creation, the value should be an base64 encoded JSON key. |

### `GCloudIAMResourceType`

- `project`
- `organization`

### `GCloudIAMResourceRoles`

A user defined roles grouping single or multiple GCloud roles.

### `GCloudIAMResourcePermission`

A Google Cloud predefined role name. These can be any roles defined under Gcloud project roles list. User defined roles group them together depending on the use case

