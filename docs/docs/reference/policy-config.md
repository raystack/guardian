# Policy Configurations

## Step config

| Field | Type | Description | Required | Default value |
| :--- | :--- | :--- | :--- | :--- |
| name | `string` |Step name | YES | - |
| description | `string` | Step description | NO | - |
| run\_if | `Expression` | Determines whether the step should be evaluated or it can be skipped. If it evaluates to be falsy, the step will automatically skipped. Otherwise, step become pending/blocked (normal). Accessible vars: `$appeal` | NO | -
| strategy | `string` | `auto` or `manual`. Determines if approval step is manual or automatic approval | YES | - |
| approvers | `string` | Determines approvers for manual approval. The evaluation should return string or []string that contains email address of the approvers. Accessible vars: `$appeal` | NO | - |
| approve_if | `string` | Expression to determines the resolution of the step if `approvers` field is not present. Accessible vars: `$appeal` | YES if `approvers` is empty | - |
| allow\_failed | `boolean` | If `true` and the step got rejected, it will mark the appeal status as skipped instead of rejected | NO | `false` |

### Variables

#### `$appeal`
   * Appeal object example:

     ```json
     {
        "id": 1,
        "resource_id": 1,
        "resource": {
          "id": 1,
          "provider_type": "google_bigquery",
          "provider_urn": "gcp-project-id",
          "type": "dataset",
          "urn": "gcp-project-id:dataset_name",
          "name": "dataset_name",
          "details": {
            "owners": [
              "owner@email.com",
              "another.owner@email.com"
            ],
            ...
          },
          "labels": {
            "key": "value"
          },
          "created_at": "2021-01-01T00:00:05.36851+07:00",
          "updated_at": "2021-01-01T00:00:05.36851+07:00"
        },
        "policy_id": "test-policy",
        "policy_version": 1,
        "status": "pending",
        "account_id": "user@email.com",
        "account_type": "user",
        "created_by": "user@email.com",
        "creator": {
          "id": 1,
          "email": "user@email.com",
          "full_name": "John Doe",
          "manager_email": "manager@email.com",
          ...
        },
        "role": "roles/viewer",
        "options": {
          "duration": "24h"
        },
        "created_at": "2021-10-26T09:29:48.838203Z",
        "updated_at": "2021-10-26T09:29:48.838203Z",
        "revoked_at": "0001-01-01T00:00:00Z",
        "details": {
          ...
        }
      }
     ```

   * Usage example
     * `$appeal.resource.id` =&gt; `1`
     * `$appeal.resource.details.owners` =&gt; `["owner@email.com", "another.owner@email.com"]`
     * `$appeal.resource.labels.key` =&gt; `"value"`
     * `$appeal.creator.manager_email` =&gt; `"manager@email.com"`

## Example

```yaml
id: bigquery_approval
steps:
  - name: supervisor_approval
    description: 'only will get evaluated if check_if_dataset_is_pii return true'
    when: $appeal.resource.details.is_pii
    strategy: manual
    approvers:
    - $appeal.creator.userManager
  - name: admin_approval
    description: approval from dataset admin/owner
    strategy: manual
    approvers:
    - $appeal.resource.details.owner
iam:
  provider: http
  config:
    url: http://localhost:5000/users/{user_id}
  profile_schema:
    id: user_id
    name: full_name
    email: email
    entity: company_name
    userManager: manager_email
```
