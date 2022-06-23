# API

## Managing Policies

Policy controls how users or accounts can get access to a resource. Policy used by appeal to determine the approval flow, get creator's identity/profile, and decide whether it needs additional appeals. Policy is attached to a resource type in the provider config, thus a policy should be the first thing to setup before creating a provider and getting its resources.

### Creating Policies

Policies can be created by calling with a **`POST`** Method on **`{{HOST}}/api/v1beta1/policies`** 

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [Policy](../reference/policy.md#policy-1) |


##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response | CreatePolicyResponse |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

** Here is an example below: **

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

Policy has `version` to ensure each appeal has a reference to an applied policy when it's created. A policy is created with an initial `version` equal to `1`.

Check [policy reference](../reference/policy.md) for more details on the policy configuration.

### Updating Policy
Updating a policy actually means creating a new policy with the same `id` but the `version` gets incremented by `1`. Both the new and previous policies still can be used by providers.

Policies can be updated by using the **`PUT`** Method on **`{{HOST}}/api/v1beta1/policies/:id`** 

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id   | path | |Yes| String|
| body | body |  | Yes | [Policy](../reference/policy.md#policy-1) |


##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **
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
### Listing Policies

To get the list of all the policies created by the user, use the ** `GET` ** Method on **`{{HOST}}/api/v1beta1/policies`**

##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

```console
$ curl --request GET '{{HOST}}/api/v1beta1/policies'
```

### Viewing Policies

Viewing a policy can be done by the ** `GET`** Method on **`{{HOST}}/api/v1beta1/policies/:id/versions/:version`**

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| version | path | | Yes | string |

##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

```console
$ curl --request GET '{{HOST}}/api/v1beta1/policies/{{policy_id}}/versions/{{policy_version}}'
```

## Managing Providers

Provider manages roles, resources, provider credentials and also points each resource type to a considered policy.

### Registering Providers

Once a provider config is registered, Guardian will immediately fetch the resources and store it in the database.

Providers can be created by calling to **`POST`** Method **`{{HOST}}/api/v1beta1/providers`**

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [Provider](../reference/provider.md#provider-configurations) |


##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **

```console
$ curl --request POST '{{HOST}}/api/v1beta1/providers' \
--header 'Content-Type: application/json' \
--data-raw '{
  "type": "bigquery",
  "urn": "gcp-project-id-bigquery",
  "allowed_account_types": [
    "user",
    "serviceAccount"
  ],
  "credentials": {
    "service_account_key": "{{base64 encoded service account key json}}",
    "resource_name": "projects/gcp-project-id"
  },
  "appeal": {
    "allow_permanent_access": false,
    "allow_active_access_extension_in": "168h"
  },
  "resources": [
    {
      "type": "dataset",
      "policy": {
        "id": "my_policy",
        "version": 1
      },
      "roles": [
        {
          "id": "viewer",
          "name": "Viewer",
          "permissions": [
            "READER"
          ]
        },
        {
          "id": "editor",
          "name": "Editor",
          "permissions": [
            "WRITER"
          ]
        }
      ]
    },
    {
      "type": "table",
      "policy": {
        "id": "my_policy",
        "version": 1
      },
      "roles": [
        {
          "id": "viewer",
          "name": "Viewer",
          "permissions": [
            "roles/bigquery.dataViewer"
          ]
        },
        {
          "id": "editor",
          "name": "Editor",
          "permissions": [
            "roles/bigquery.dataEditor"
          ]
        }
      ]
    }
  ]
}'
```

Check [provider reference](../reference/provider.md) for more details on Provider schema.

### Updating Providers

Providers can be updated by calling to **`PUT`** Method **`{HOST}}/api/v1beta1/providers/:id`**

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| provider_id   | path | |Yes| String|
| body | body |  | Yes | [Provider](../reference/provider.md#provider-configurations) |


##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **
```console
$ curl --request PUT '{{HOST}}/api/v1beta1/providers/{{provider_id}}' \
--header 'Content-Type: application/json' \
--data-raw '{
  "allowed_account_types": [
    "user"
  ]
}'
```

### Listing Providers

To get the list of all the providers avaliable, call the **`GET`** Method on **`{{HOST}}/api/v1beta1/providers`**

##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **
```console
$ curl --request GET '{{HOST}}/api/v1beta1/providers'
```

### Viewing Providers

To see the details of a particular provider by id, call the **`GET`** Method on **`{{HOST}}/api/v1beta1/providers/:id`** with the following parameters:

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| provider_id   | path | |Yes| String|


##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **
```console
$ curl --request GET '{{HOST}}/api/v1beta1/providers/{{provider_id}}'
```

### Listing Roles for a Resource Type

Listing roles can be done by calling to **`GET`** Method **`{{HOST}}/api/v1beta1/providers/:id/resources/:resource_type/roles`**

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| resource_type | path | | Yes | string |

##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **
```console
$ curl --request GET '{{HOST}}/api/v1beta1/providers/{{provider_id}}/resources/{{resource_type}}/roles'
```

## Managing Resources

Resource in Guardian represents the actual resource in the provider e.g. for BigQuery provider, a resource represents a dataset or a table. One of Guardian's responsibility is to manage the access to resources, and in order to do that Guardian needs to be able to manage the resources as well.

#### Collecting Resources

Guardian collects resources from the provider automatically as soon as it registered. While in parallel, Guardian also has a job for continously syncing resources.

#### Example
```json
{
  "id": "a32b702a-029d-4d76-90c4-c3b8cc52941b",
  "provider_type": "bigquery",
  "provider_urn": "gcp-project-id-bigquery",
  "type": "table",
  "urn": "gcp-project-id:dataset_name.table_name",
  "name": "table_name",
  "details": {
    "is_sensitive": false,
    "owner": [
      "john.doe@example.com",
      "john.smith@example.com"
    ]
  }
}
```

### Updating Resources Metadata

Guardian allows users to add metadata to the resources. This can be useful when configuring the approval steps in the policy that needs information from metadata e.g. “owners” as the approvers.

Update a resource can be done by calling to **`PUT`** Method **`{{HOST}}/api/v1beta1/resources/:id`**

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| resource_id   | path | |Yes| String|
| body | body |  | Yes | TODO |


##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **
```console
$ curl --request PUT '{{HOST}}/api/v1beta1/resources/{{resource_id}}' \
--header 'Content-Type: application/json' \
--data-raw '{
    "details": {
        "key1": "value1",
        "key2": "value2"
    }
}'
```

### Listing Resources

To get the list of all the resources availiable, call the **`GET`** Method on **`{{HOST}}/api/v1beta1/resources`**

##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **

```console
$ curl --request GET '{{HOST}}/api/v1beta1/resources'
```

### Viewing Resources

To see the details of a particular resource by id, call the **`GET`** Method on **`{{HOST}}/api/v1beta1/resources/:id`** using the following parameters:

##### Parameters - TODO

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id   | path | |Yes| String|


##### Responses - TODO

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

** Here is an example below: **
```console
$ curl --request GET '{{HOST}}/api/v1beta1/resources/{{resource_id}}'
```
---

## Managing Appeals

An appeal is essentially a request created by users to give them access to resources. In order to grant the access, an appeal has to be approved by approvers which is assigned based on the applied policy. Appeal contains information about the requested account, the creator, the selected resources, the specific role for accessing the resource, and options to determine the behaviour of the access e.g. permanent or temporary access.

### Creating Appeals

Guardian creates access for users to resources with configurable approval flow. Appeal created by user with specifying which resource they want to access and also the role.

Appeals can be created by calling the **`POST`** Method on **`{{HOST}}/api/v1beta1/appeals`** using the following parameters:

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [Appeal](../reference/appeal.md#appeal-1)|


##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |


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

### Canceling Appeals

#### Appeal creator can cancel their appeal while it's status is still on `pending`.

Appeals can be canceled by calling the **`PUT`** Method on **`{{HOST}}/api/v1beta1/appeals/:id/cancel`** endpoint using the following parameters:

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id   | path | | Yes | String|

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |

```console
$ curl --request PUT '{{HOST}}/api/v1beta1/appeals/{{appeal_id}}/cancel'
```

### Approving and Rejecting Appeals

Appeals can be approved/rejected by calling the **`POST`** Method on **`{{HOST}}/api/v1beta1/appeals/:id/approvals/:approval_step_name/`** endpoint with the following parameters:

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path | | Yes | String|
| approval_name | path || Yes | String| 
| body | body |  | Yes | [Approval](../reference/appeal.md#approval) |


##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | v1beta1CreatePolicyResponse |
| default | An unexpected error response. | rpcStatus |


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

## Models

#### rpcStatus

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| message | string |  | No |
| details | [ [protobufAny](#protobufany) ] |  | No |

#### protobufAny

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| typeUrl | string |  | No |
| value | byte |  | No |