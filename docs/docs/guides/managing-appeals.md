# Managing appeals

## Creating appeal

This is the main use case of Guardian, to manage access approval for user to a particular resource. Appeal is created by a user with specifying which resource they want to access along with some other appeal options.

### Appeal Lifecycle

![](/assets/appeal-lifecycle.png)

#### Request statuses

- Pending \(initial status\): During this state, the appeal will evaluate approval steps one by one. The result from the approval steps evaluation will determine whether the appeal will be approved or rejected.
- Rejected: The appeal has at least one failed approval step.
- Active: The appeal has been approved. As long as the appeal is in this status, the user will have the access to the designated resource.
- Terminated: An active access can be revoked by any authorized user at any time, or, if the appeal already exceeds the lifetime limit then it will automatically get revoked.

#### Actions

- Approve: Called when all the approval steps are passed/approved.
- Reject: Called when there is one approval step that is rejected.
- Revoke: A manual action that is called by an authorized user intentionally to revoke an active access.
- Expire: If the appeal specifies the expiration policy then it will automatically get expired when it is already passed the lifetime limit.
- Recreate: Possible for appeals that are currently still active, rejected, or terminated. This action will create a new appeal based on the previous one. For the appeal coming from active status, there is a policy related to access extension.

To create an appeal, you can use this endpoint:

```text
POST /appeals
Content-Type: application/json
Accept: application/json

Request Body:
{
   "user": "user@email.com",
   "resources": [
       {
           "id": 1,
           "role": "viewer",
           "options": {
             "duration": "24h"
           }
       }
   ]
}

Response:
[
  {
    "id": 1,
    "resource_id": 1,
    "policy_id": "bigquery_dataset",
    "policy_version": 1,
    "status": "pending",
    "user": "user@email.com",
    "role": "viewer",
    "options": {
      "expiration_date": "2021-05-10T09:49:26.402189+07:00"
    },
    "labels": null,
    "approvals": [
      {
        "id": 11,
        "name": "check_if_dataset_is_pii",
        "appeal_id": 11,
        "status": "pending",
        "actor": null,
        "policy_id": "bigquery_dataset",
        "policy_version": 1
      },
      {
        "id": 12,
        "name": "supervisor_approval",
        "appeal_id": 11,
        "status": "pending",
        "actor": null,
        "policy_id": "bigquery_dataset",
        "policy_version": 1,
        "approvers": [
          "john.doe@email.com"
        ]
      },
      {
        "id": 13,
        "name": "admin_approval",
        "appeal_id": 11,
        "status": "pending",
        "actor": null,
        "policy_id": "bigquery_dataset",
        "policy_version": 1,
        "approvers": [
          "owner@email.com"
        ]
      }
    ]
  }
]
```

## Approving/Rejecting appeal

![](/assets/approval-flow.png)

Completing an appeal to gain the access to the designated resource could consist of multiple approvals, depending on the [approval policy](../reference/policy.md) applied to the designated resource. In Guardian, it called approval steps. Approval steps are determined during the appeal creation. For approval step without approvers, Guardian will evaluate it and resolve the status immediately. But for one with approvers, an action is required to approve/reject that particular approval step.

That action is can be done by using this endpoint:

```text
PUT /appeals/:id/approvals/:step_name
Content-Type: application/json
Accept: application/json

Params:
id: 1
step_name: supervisor_approval

Request Body:
{
    "action": "approve"
}

Response:
{
  "id": 1,
  "resource_id": 1,
  "policy_id": "bigquery_dataset",
  "policy_version": 1,
  "status": "pending",
  "user": "user@email.com",
  "role": "viewer",
  "options": {
    "expiration_date": "2021-05-10T09:49:26.402189+07:00"
  },
  "labels": null,
  "approvals": [
    {
      "id": 11,
      "name": "check_if_dataset_is_pii",
      "appeal_id": 11,
      "status": "pending",
      "actor": null,
      "policy_id": "bigquery_dataset",
      "policy_version": 1
    },
    {
      "id": 12,
      "name": "supervisor_approval",
      "appeal_id": 11,
      "status": "approved", // this approval step’s status is updated to ‘approved’
      "actor": null,
      "policy_id": "bigquery_dataset",
      "policy_version": 1,
      "approvers": [
        "john.doe@email.com"
      ]
    },
    {
      "id": 13,
      "name": "admin_approval",
      "appeal_id": 11,
      "status": "pending",
      "actor": null,
      "policy_id": "bigquery_dataset",
      "policy_version": 1,
      "approvers": [
        "owner@email.com"
      ]
    }
  ]
}
```
