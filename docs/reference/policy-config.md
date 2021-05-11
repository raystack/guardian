# Policy Configurations

## Step config

| Field | Description | Required | Default value |
| :--- | :--- | :--- | :--- |
| name | Step name | YES | - |
| description | Step description | NO | - |
| approvers | Object path from [these variables](policy-config.md#variables), or list of approver emails | NO | - |
| conditions | List of conditions. An approval step will be considered as successful if all conditions are passed | YES if `approvers` is empty | - |
| allow\_failed | If `true` and the conditions failed, it will mark the appeal status as skipped instead of rejected | NO | `false` |
| dependencies | List of dependency step name | NO | - |

### Variables

1. `$resource`: the requested resource object
   * Resource object example:

     ```javascript
     {
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
             "additional_info": "..."
         },
         "labels": {
             "key": "value"
         },
         "created_at": "2021-01-01T00:00:05.36851+07:00",
         "updated_at": "2021-01-01T00:00:05.36851+07:00"
     }
     ```

   * Usage example
     * `$resource.id` =&gt; `1`
     * `$resource.details.owners` =&gt; `["owner@email.com", "another.owner@email.com"]`
     * `$resource.labels.key` =&gt; `"value"`
2. `$user_approvers`: fetch to third-party service to resolve user's approvers
   * Usage example

     * appeal creator: `user@email.com`
     * approvers: `$user_approvers`
     * configured third-party service URL: `http://localhost:5000/user-approvers`

     Guardian will fetch to `http://localhost:5000/user-approvers?user=user@email.com` and expecting the response body to be like this:

     ```javascript
     {
         "emails": [
           "approver1@email.com",
           "approver2@email.com"
         ]
     }
     ```

     Given the response, Guardian will set the approvers to `approver1@email.com` and `approver2@email.com` for that particular approval step.

## Example

```yaml
id: bigquery_approval
steps:
  - name: check_if_dataset_is_pii
    description: pii dataset needs additional approval from the team lead
    conditions:
    - field: $resource.details.is_pii
      match:
        eq: true
    allow_failed: true
  - name: supervisor_approval
    description: 'only will get evaluated if check_if_dataset_is_pii return true'
    dependencies: [check_if_dataset_is_pii]
    approvers: $user.profile.team_leads.[].email
  - name: admin_approval
    description: ...
    approvers: $resource.details.owner
```

