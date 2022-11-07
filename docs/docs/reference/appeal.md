# Appeal

#### JSON Representation

```json
{
  "id": "49d3d948-d5f5-4f8a-affc-8547bc02ec4f",
  "resource_id": "60999a98-b037-4a7e-8e9f-1999bc3be9cb",
  "resource": {
    "id": "60999a98-b037-4a7e-8e9f-1999bc3be9cb",
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
    "expiration_date": "2024-01-01T00:00:05.36851+07:00",
    "duration": "24h"
  },
  "details": {},
  "description": "This is a test appeal",
  "labels": {
    "key": "value"
  },
  "approvals": [
    {
      "id": "c6d2e6f1-5767-49ba-8eef-8fb8f0006f3a",
      "name": "owner_approval",
      "appeal_id": "d95dde82-5719-48f9-b92b-9bd216499a77",
      "status": "pending",
      "actor": "john.doe@example.com",
      "reason": "LGTM",
      "policy_id": "test-policy",
      "policy_version": 1,
      "approvers": [
        "john.doe@example.com"
      ],
      "created_at": "2021-10-26T09:29:48.838203Z",
      "updated_at": "2021-10-26T09:29:48.838203Z"
    }
  ],
  "grant": {
    "id": "ecd81395-7879-476f-b39b-cbf38d707b07",
    "status": "active",
    "status_in_provider": "active",
    "account_id": "user@email.com",
    "account_type": "user",
    "resource_id": "3d87367a-8cd6-4f6c-aee0-4bb29b82e9ff",
    "role": "viewer",
    "permissions": [
      "READER"
    ],
    "is_permanent": false,
    "expiration_date": "2024-01-01T00:00:05.36851+07:00",
    "appeal_id": "49d3d948-d5f5-4f8a-affc-8547bc02ec4f",
    "source": "appeal",
    "created_by": "user@email.com",
    "owner": "owner@email.com",
    "created_at": "2021-10-26T09:29:48.838203Z",
    "updated_at": "2021-10-26T09:29:48.838203Z"
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
  "created_at": "2021-10-26T09:29:48.838203Z",
  "updated_at": "2021-10-26T09:29:48.838203Z",
  "revoked_at": "0001-01-01T00:00:00Z"
}
```

### Appeal

| Field                     | Type                                         | Description                                                                                                                                                                                                                                                                                                                             |
|---------------------------|----------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `id`                      | `string`                                     | Unique identifier of appeal.                                                                                                                                                                                                                                         |
| `resource_id`             | `string`                                     | Resource identifier.                                                                                                                                                                                                                                                                                                            |
| `resource`                | [`object(Resource)`](resource.md#resource-1) | Complete resource information.                                                                                                                                                                                                                                                                                                          |
| `role`                    | `string`                                     | Permission type chosen by the creator to access the resource.<br/>Example: `roles/viewer`                                                                                                                                                                                                                                                                              |
| `options`                 | [`object(AppealOptions)`](#appealoptions)  | Options for the appeal.                                                                                                                                                                                                                                                                                                                 |
| `details`                 | `object`                                     | Additional information for the appeal. Details can be added from the appeal creation.                                                                                                                                                                                                                                                   |
| `description`             | `string`                                     | Description of the appeal.                                                                                                                                                                                                                                                                                                              |
| `approvals`               | [`[]object(Approval)`](#approval)            | Approval steps applied for current appeal based on the applicable policy.                                                                                                                                                                                                                                                               |
| `grant`                   | [`object(Grant)`](#grant)                  | Grant created after the appeal is approved.                                                                                                                                                                                                                                                                                             |
| `policy_id`               | `string`                                     | Policy identifier                                                                                                                                                                                                                                                                                                                       |
| `policy_version`          | `uint`                                       | Policy version identifier. Used together with `policy_id` to reference to a policy.                                                                                                                                                                                                                                                     |
| `status`                  | `string`                                     | Current status of the appeal. The initial status is `pending`. If the appeal creator canceled/removed the appeal while its on pending, the status is become `canceled`. After the approval steps completed, the status either become `active` or `rejected`. And if it gets expired or an admin revoked the status become `terminated`. <br/>Reference: [Appeal Status](/docs/concepts/overview#appeal-status) |
| `account_type`            | `string`                                     | Type of the account based on the Provider of the selected `resource`. Default value is `user`                                                                                                                                                                                                                                           |
| `account_id`              | `string`                                     | An account identifier related to `account_type` that will get the permission to the targetted resource once the appeal is approved.                                                                                                                                                                                                     |
| `created_by`              | `string`                                     | Email address of the appeal creator.                                                                                                                                                                                                                                                                                                    |
| `creator`                 | `object`                                     | Creator user details information fetched from the configured identity manager as in the [Policy Config](policy.md).                                                                                                                                                                                                                     |
| `created_at`              | `string`                                     | Timestamp when the appeal created.                                                                                                                                                                                                                                                                                                      |
| `updated_at`              | `string`                                     | Timestamp when the appeal last modified.                                                                                                                                                                                                                                                                                                |
| `revoked_at`              | `string`                                     | Timestamp when the appeal gets revoked.                                                                                                                                                                                                                                                                                                 |
| `revoked_by`              | `string`                                     | Email address of the user who revoke the appeal.                                                                                                                                                                                                                                                                                        |
| `revoke_reason`           | `string`                                     | Reason filled by the revoking user to inform the appeal creator why the appeal gets revoked.                                                                                                                                                                                                                                            |

### AppealOptions

| Field           | Type       | Description                                                                                                                                                                                                                                    |
|-----------------|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| expiration_date | `dateTime` | Timestamp when the appeal expires                                                                                                                                                                                                              |
| duration        | `string`   | actual value of duration such as `24h`, `72h`. value will be `0h` in case of permanent duration. <br/> Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. Reference: [ParseDuration](https://pkg.go.dev/time#ParseDuration)       |

### Approval

| Field            | Type       | Description                                                                         |
|------------------|------------|-------------------------------------------------------------------------------------|
| `id`             | `string`   | Approval step unique identifier                                                     |
| `name`           | `string`   | Unique approval step name                                                           |
| `appeal_id`      | `uint`     | Appeal identifier                                                                   |
| `status`         | `string`   | The status of approval step <br/>Reference: [Approval Status](#approval-status) |
| `policy_id`      | `string`   | Policy identifier                                                                   |
| `policy_version` | `uint`     | Policy version identifier. Used together with `policy_id` to reference to a policy. |
| `approvers`      | `[]string` | List of email address of eligible approvers if require manual approval.             |
| `actor`          | `string`   | Email address of the approver who resolve the status of current approval step.      |
| `reason`         | `string`   | Rejection reason filled by the actor if they rejecting current approval step.       |
| `created_at`     | `string`   | Timestamp when the appeal created.                                                  |
| `updated_at`     | `string`   | Timestamp when the appeal last modified.                                            |

### Grant

| Field                | Type       | Description                                                                         |
|----------------------|------------|-------------------------------------------------------------------------------------|
| `id`                 | `string`   | Grant unique identifier                                                             |
| `status`             | `string`   | The status of grant <br/>Reference: [Grant Status](#grant-status)                   |
| `status_in_provider` | `string`   | The status of grant in the provider <br/>Reference: [Grant Status](#grant-status)   |
| `account_id`         | `string`   | An account identifier related to `account_type` that will get the permission to the targetted resource once the appeal is approved. |
| `account_type`       | `string`   | Type of the account based on the Provider of the selected `resource`. Default value is `user` |
| `resource_id`        | `string`   | Resource identifier                                                                 |
| `role`               | `string`   | Role identifier                                                                     |
| `permissions`        | `[]string` | List of permissions granted to the account                                          |
| `is_permanent`       | `bool`     | Indicates if the grant is permanent or not                                          |
| `expiration_date`    | `string`   | Timestamp when the grant expires                                                    |
| `appeal_id`          | `string`   | Appeal identifier                                                                   |
| `source`             | `string`   | Source of the grant <br/>Reference: [Grant Source](#grant-source)                   |
| `owner`              | `string`   | Email address of the user who created the grant                                     |
| `created_at`         | `string`   | Timestamp when the grant created.                                                   |
| `updated_at`         | `string`   | Timestamp when the grant last modified.                                             |

### Approval Status

- `pending` (initial status): During this state the approvers will determine whether the appeal will be approved or rejected
- `blocked`: The step is approved is blocked by prior step(s)
- `skipped`: The step is approved is skipped due to prior step are rejected
- `approved`: The step is approved by approvers
- `rejected`: The step is rejected by approvers

### Grant Status

- `active`: The grant is active and valid
- `inactive`: The grant is expired or revoked

### Grant Source

- `appeal`: The grant is created from an appeal
- `import`: The grant is imported from the provider