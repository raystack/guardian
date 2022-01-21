# Managing appeals

## Creating appeal

Appeals can be created in the following ways:
1. [Using `guardian appeal create` CLI command](#1-create-an-appeal-using-cli)
2. [Calling to `POST /api/v1beta1/appeals` API](#2-create-an-appeal-using-http-api)

Guardian creates access for users to resources with configurable approval flow. Appeal created by user with specifying which resource they want to access and also the role.

### Appeal Lifecycle

![](/assets/appeal-lifecycle.png)

#### Appeal Status

- `pending` \(initial status\): During this state, the appeal will evaluate approval steps one by one. The result from the approval steps evaluation will determine whether the appeal will be approved or rejected.
- `rejected`: The appeal has at least one failed approval step.
- `active`: The appeal has been approved. As long as the appeal is in this status, the user will have the access to the designated resource.
- `terminated`: An active access can be revoked by any authorized user at any time, or, if the appeal already exceeds the lifetime limit then it will automatically get revoked.
- `canceled`: The appeal canceled by the creator when the status was on pending.

#### Actions

- Approve: Called when all the approval steps are passed/approved.
- Reject: Called when there is one approval step that is rejected.
- Revoke: A manual action that is called by an authorized user intentionally to revoke an active access.
- Expire: If the appeal specifies the expiration policy then it will automatically get expired when it is already passed the lifetime limit.
- Recreate: Possible for appeals that are currently still active, rejected, or terminated. This action will create a new appeal based on the previous one. For the appeal coming from active status, there is a policy related to access extension.

### 1. Create an Appeal using CLI
```console
$ guardian appeal create --account-id=user@example.com --resource=a32b702a-029d-4d76-90c4-c3b8cc52941b --role=viewer
```

### 2. Create an Appeal using HTTP API
```console
$ curl --request POST '{{HOST}}/api/v1beta1/appeals' \
--header 'X-Auth-Email: user@example.com' \
--header 'Content-Type: application/json' \
--data-raw '{
  "account_id": "user@example.com",
  "resources": [
    {
      "id": "a32b702a-029d-4d76-90c4-c3b8cc52941b",
      "role": "viewer"
    }
  ]
}'
```

## Canceling Appeals

Appeal creator can cancel their appeal while it's status is still on `pending`.

Appeals can be canceled in the following ways:
1. [Calling to `PUT /api/v1beta1/appeals/:id/cancel` API](#1-cancel-an-appeal-using-http-api)

### 1. Cancel an Appeal using HTTP API
```console
$ curl --request PUT '{{HOST}}/api/v1beta1/appeals/{{appeal_id}}/cancel'
```

## Approving and Rejecting Appeals

![](/assets/approval-flow.png)

Appeals can be approved/rejected in the following ways:
1. [Using `guardian appeal approve/reject` CLI command](#1-approve-or-reject-an-appeal-using-cli)
2. [Calling to `POST /api/v1beta1/appeals/:id/approvals/:approval_step_name/` API](#2-approve-or-reject-an-appeal-using-http-api)

### 1. Approve or Reject an Appeal using CLI
#### Approve an Appeal
```console
$ guardian appeal approve --id={{appeal_id}} --step={{approval_step_name}}
```

#### Reject an Appeal
```console
$ guardian appeal reject --id={{appeal_id}} --step={{approval_step_name}} --reason={{rejection_message}}
```

### 2. Approve or Reject an Appeal using HTTP API
#### Approve an Appeal
```console
$ curl --request POST '{{HOST}}/api/v1beta1/appeals/{{appeal_id}}/approvals/{{approval_step_name}}' \
--header 'X-Auth-Email: user@example.com' \
--header 'Content-Type: application/json' \
--data-raw '{
    "action": "approve"
}'
```

#### Reject an Appeal
```console
$ curl --request POST '{{HOST}}/api/v1beta1/appeals/{{appeal_id}}/approvals/{{approval_step_name}}' \
--header 'X-Auth-Email: user@example.com' \
--header 'Content-Type: application/json' \
--data-raw '{
    "action": "reject",
    "reason": "{{rejection_message}}"
}'
```
