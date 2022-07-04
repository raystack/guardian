# Appeal

#### JSON Representation

```json
{
  "id": 1,
  "resource_id": 1,
  "resource": {
    "id": 1,
    "provider_type": "bigquery",
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
  "role": "roles/viewer",
  "options": {
    "duration": "24h"
  },
  "details": {},
  "approvals": [
    {
      "id": 1,
      "name": "owner_approval",
      "appeal_id": 1,
      "status": "pending",
      "policy_id": "test-policy",
      "policy_version": 1,
      "approvers": [
        "john.doe@example.com"
      ],
      "created_at": "2021-10-26T09:29:48.838203Z",
      "updated_at": "2021-10-26T09:29:48.838203Z"
    }
  ],
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
  "created_at": "2021-10-26T09:29:48.838203Z",
  "updated_at": "2021-10-26T09:29:48.838203Z",
  "revoked_at": "0001-01-01T00:00:00Z"
}
```

### `Appeal`

| Field | Type | Description |
| :----- | :---- | :------ |
| `id` | `uint` | Unique identifier of appeal. |
| `resource_id` | `uint` | Resource identifier. |
| `resource` | [`object(Resource)`](resource.md#resource-1)| Complete resource information. |
| `role` | `string` |Permission type chosen by the creator to access the resource. |
| `options` | `object`| Options for the appeal. |
| `options.duration` | `string`| This field is for specifying how long the access would be granted until it automatically gets revoked by the system. Only time units like `h`, `m`, `s`, and `ms` are supported. Examples: `48h`, `48h30m`, `20m`. |
| `options.expiration_date` | `string`| Timestamp for when the appeal will expire. This field is automatically filled by Guardian by calculating the activation time plus the `duration`. |
| `details` | `object` |Additional information for the appeal. Details can be added from the appeal creation.
| `approvals` | [`[]object(Approval)`](#approval)| Approval steps applied for current appeal based on the applicable policy. |
| `policy_id` | `string`| Policy identifier |
| `policy_version` | `uint` | Policy version identifier. Used together with `policy_id` to reference to a policy.|
| `status` | `string` | Current status of the appeal. The initial status is `pending`. If the appeal creator canceled/removed the appeal while its on pending, the status is become `canceled`. After the approval steps completed, the status either become `active` or `rejected`. And if it gets expired or an admin revoked the status become `terminated`. |
| `account_type` | `string` | Type of the account based on the Provider of the selected `resource`. Default value is `user` |
| `account_id` | `string` |An account identifier related to `account_type` that will get the permission to the targetted resource once the appeal is approved. |
| `created_by` | `string`| Email address of the appeal creator.|
| `creator` | `object`| Creator user details information fetched from the configured identity manager as in the [Policy Config](policy.md). |
| `created_at` | `string`| Timestamp when the appeal created. |
| `updated_at` | `string` |Timestamp when the appeal last modified. |
| `revoked_at` | `string` |Timestamp when the appeal gets revoked. |
| `revoked_by` | `string` |Email address of the user who revoke the appeal. |
| `revoke_reason` | `string` | Reason filled by the revoking user to inform the appeal creator why the appeal gets revoked. |

### `Approval`

| Field | Type | Description |
| :----- | :---- | :------ |
| `id` | `uint` |Approval step unique identifier |
| `name` | `string`| Unique approval step name |
| `appeal_id` | `uint` |Appeal identifier |
| `status` | `string` ||
| `policy_id` | `string`| Policy identifier |
| `policy_version` | `uint`| Policy version identifier. Used together with `policy_id` to reference to a policy. |
| `approvers` | `[]string` |List of email address of eligible approvers if require manual approval. |
| `actor` | `string` |Email address of the approver who resolve the status of current approval step. |
| `reason` | `string` |Rejection reason filled by the actor if they rejecting current approval step. |
| `created_at` | `string` |Timestamp when the appeal created. |
| `updated_at` | `string` |Timestamp when the appeal last modified. |
