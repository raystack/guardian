# Managing policies

Access can be given to a user after it passed a set of approval steps that we call policy. Guardian lets you configure the policy based on your own governance rules of the data access.

## Creating Policy

Creating a policy is the first step required for setting up Guardian in your environment. It is a dependency for the next step which is setting up provider configuration.

Policy created by providing a configuration of the approval flow itself. Here’s the example of the configuration:

```yaml
id: bigquery_dataset
steps:
  - name: check_if_dataset_is_pii
    description: pii dataset needs additional approval from the team lead
    approve_if:
    - field: $resource.details.is_pii
      match:
        eq: true
    allow_failed: true
  - name: supervisor_approval
    description: 'only will get evaluated if check_if_dataset_is_pii return true'
    dependencies: [check_if_dataset_is_pii]
    approvers: $user_approvers
  - name: admin_approval
    description: ...
    approvers: $resource.details.owner
```

Check [policy reference](../reference/policy-config.md) for more details on the policy configuration

To create a policy, you can use this endpoint

```text
POST /policies
Accept: application/json

Request Body:
id: bigquery_dataset
steps:
  - name: check_if_dataset_is_pii
    description: pii dataset needs additional approval from the team lead
    approve_if:
    - field: $resource.details.is_pii
      match:
        eq: true
    allow_failed: true
  - name: supervisor_approval
    description: 'only will get evaluated if check_if_dataset_is_pii return true'
    dependencies: [check_if_dataset_is_pii]
    approvers: $user_approvers
  - name: admin_approval
    description: ...
    approvers: $resource.details.owner

Response:
{
  "id": "bigquery_dataset",
  "version": 1,
  "description": "",
  "steps": [
    {
      "name": "check_if_dataset_is_pii",
      "description": "pii dataset needs additional approval from the team lead",
      "approve_if": [
        {
          "field": "$resource.details.is_pii",
          "match": {
            "eq": true
          }
        }
      ],
      "allow_failed": true,
      "dependencies": null,
      "approvers": ""
    },
    {
      "name": "supervisor_approval",
      "description": "only will get evaluated if check_if_dataset_is_pii return true",
      "approve_if": null,
      "allow_failed": false,
      "dependencies": [
        "check_if_dataset_is_pii"
      ],
      "approvers": "$user_approvers"
    },
    {
      "name": "admin_approval",
      "description": "...",
      "approve_if": null,
      "allow_failed": false,
      "dependencies": null,
      "approvers": "$resource.details.owner"
    }
  ],
  "labels": null,
  "created_at": "2021-05-04T08:05:07.691557+07:00",
  "updated_at": "2021-05-04T08:05:07.691557+07:00"
}
```

## Updating Policy

In Guardian, we keep track of the policy changes using the policy version which generated after you create/update a policy. This version tracking will help Guardian in case you updated the policy, the existing appeals that were created using the previous version would still be able to retrieve the matching policy version.

By updating a policy, Guardian will automatically bump up the version of that particular policy. For example, if the current version of policy `bigquery_dataset` is `1`, the version will automatically get increased to `2` when it gets updated.

To update a policy, you can use this endpoint:

```text
PUT /policies/:id
Accept: application/json

Request Body:
id: bigquery_dataset
steps:
  - name: check_if_dataset_is_pii
    description: pii dataset needs additional approval from the team lead
    approve_if:
    - field: $resource.details.is_pii
      match:
        eq: true
    allow_failed: true
  - name: supervisor_approval
    description: 'only will get evaluated if check_if_dataset_is_pii return true'
    dependencies: [check_if_dataset_is_pii]
    approvers: $user_approvers
  - name: admin_approval
    description: ...
    approvers: $resource.details.owners


Response:
{
  "id": "bigquery_dataset",
  "version": 2,
  "description": "",
  "steps": [
    {
      "name": "check_if_dataset_is_pii",
      "description": "pii dataset needs additional approval from the team lead",
      "approve_if": [
        {
          "field": "$resource.details.is_pii",
          "match": {
            "eq": true
          }
        }
      ],
      "allow_failed": true,
      "dependencies": null,
      "approvers": ""
    },
    {
      "name": "supervisor_approval",
      "description": "only will get evaluated if check_if_dataset_is_pii return true",
      "approve_if": null,
      "allow_failed": false,
      "dependencies": [
        "check_if_dataset_is_pii"
      ],
      "approvers": "$user_approvers"
    },
    {
      "name": "admin_approval",
      "description": "...",
      "approve_if": null,
      "allow_failed": false,
      "dependencies": null,
      "approvers": "$resource.details.owners"
    }
  ],
  "labels": null,
  "created_at": "2021-05-04T08:05:07.691557+07:00",
  "updated_at": "2021-05-04T08:05:07.691557+07:00"
}
```

