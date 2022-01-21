# Managing policies

Policy controls how users or accounts can get access to a resource. Policy used by appeal to determine the approval flow, get creator's identity/profile, and decide whether it needs additional appeals. Policy is attached to a resource type in the provider config, thus a policy should be the first thing to setup before creating a provider and getting its resources.

## Creating Policies

Policies can be created in the following ways:
1. [Using `guardian policy create` CLI command](#1-create-a-policy-using-cli)
2. [Calling to `POST /api/v1beta1/policies` API](#2-create-a-policy-using-http-api)

Policy has `version` to ensure each appeal has a reference to an applied policy when it's created. A policy is created with an initial `version` equal to `1`.

#### Example
```yaml
# policy.yaml
id: my_policy
steps:
  - name: manager_approval
    description: Manager approval for sensitive data
    when: $appeal.resource.details.is_sensitive == true
    strategy: manual
    approvers:
      - $appeal.creator.manager_email
  - name: resource_owner_approval
    description: Approval from resource admin/owner
    strategy: manual
    approvers:
      - $appeal.resource.details.owner
```

Check [policy reference](../reference/policy.md) for more details on the policy configuration

### 1. Create a Policy using CLI

```console
$ guardian policy create --file=policy.yaml
```

### 2. Create a Policy using HTTP API

```console
$ curl --request POST '{{HOST}}/api/v1beta1/policies' \
--header 'Content-Type: application/json' \
--data-raw '{
  "id": "my_policy",
  "steps": [
    {
      "name": "manager_approval",
      "description": "Manager approval for sensitive data",
      "when": "$appeal.resource.details.is_sensitive == true",
      "strategy": "manual",
      "approvers": [
        "$appeal.creator.manager_email"
      ]
    },
    {
      "name": "resource_owner_approval",
      "description": "Approval from resource admin/owner",
      "strategy": "manual",
      "approvers": [
        "$appeal.resource.details.owner"
      ]
    }
  ]
}'
```
## Updating Policy

Policies can be updated in the following ways:
1. [Using `guardian policy edit` CLI command](#1-update-a-policy-using-cli)
2. [Calling to `PUT /api/v1beta1/policies/:id` API](#2-update-a-policy-using-http-api)

Updating a policy actually means creating a new policy with the same `id` but the `version` gets incremented by `1`. Both the new and previous policies still can be used by providers.

### 1. Update a Policy using CLI

```console
$ guardian policy edit --file=policy.yaml
```

### 2. Update a Policy using HTTP API

```console
$ curl --request PUT '{{HOST}}/api/v1beta1/policies/{{policy_id}}' \
--header 'Content-Type: application/json' \
--data-raw '{
  "steps": [
    {
      "name": "manager_approval",
      "description": "Manager approval for sensitive data",
      "when": "$appeal.resource.details.is_sensitive == true",
      "strategy": "manual",
      "approvers": [
        "$appeal.creator.manager_email"
      ]
    },
    {
      "name": "resource_owner_approval",
      "description": "Approval from resource admin/owner",
      "strategy": "manual",
      "approvers": [
        "$appeal.resource.details.owner"
      ]
    }
  ]
}'
```

## Listing Policies

Listing policies can be done in the following ways:
1. [Using `guardian policy list` CLI command](#1-list-policies-using-cli)
2. [Calling to `GET /api/v1beta1/policies` API](#2-list-policies-using-http-api)

### 1. List Policies using CLI
```console
$ guardian policy list --output=yaml
```

### 2. List Policies using HTTP API
```console
$ curl --request GET '{{HOST}}/api/v1beta1/policies'
```

## Viewing Policies

Viewing a policy can be done in the following ways:

1. [Using `guardian policy view` CLI command](#1-view-a-policy-using-cli)
2. [Calling to `GET /api/v1beta1/policies/:id/versions/:version` API](#2-view-a-policy-using-http-api)

### 1. View a Policy using CLI
```console
$ guardian policy view {{policy_id}} --version={{policy_version}}
```

### 2. View a Policy using HTTP API
```console
$ curl --request GET '{{HOST}}/api/v1beta1/policies/{{policy_id}}/versions/{{policy_version}}'
```
