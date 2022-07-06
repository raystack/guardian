# GCP

## Prerequisites

1. A service account with `roles/iam.securityAdmin` role at the project/organization level

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
