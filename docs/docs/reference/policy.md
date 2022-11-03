# Policy

#### YAML Representation

```yaml
id: bigquery_approval
version: 1
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
appeal_config:
  duration_options:
    - name: 1 Day
      value: 24h
    - name: 3 Days
      value: 72h
    - name: Permanent
      value: 0h
iam:
  provider: http
  config:
    url: http://localhost:5000/users/{user_id}
  schema:
    id: user_id
    name: full_name
    email: email
    entity: company_name
    userManager: manager_email
requirements:
  - on:
      provider_type: bigquery
      role: writer
    appeals:
      - resource:
          id: 99
        role: roles/bigquery.jobUser
        policy:
          id: auto_approval
          version: 1
```

### `Policy`

| Field           | Type                                    | Description                                                           | Required |
|:----------------|:----------------------------------------|:----------------------------------------------------------------------|:---------|
| `id`            | `string`                                | Policy unique identifier                                              | YES      |
| `version`       | `uint`                                  | Auto increment value. Keeping the                                     | NO       |
| `steps`         | [`[]object(Step)`](#step)               | Sequence of approval steps                                            | YES      |
| `appeal_config` | [`object(PolicyAppealConfig)`](#policyappealconfig)            | appeal configuration options                                          | NO       |
| `iam`           | [`object(IAM)`](#iam)                   | Identity manager configuration for client and identity/creator schema | NO       |
| `requirements`  | [`[]object(Requirement)`](#requirement) | Additional appeals                                                    | YES      |

### `Step`

| Field | Type | Description | Required |
| :----- | :---- | :------ | :------ |
| `name` | `string` | Approval step identifier | YES |
| `description` | `string` | Approval step description | NO |
| `when` | [`Expression`](#expression) | Determines whether the step should be evaluated or it can be skipped. If it evaluates to be falsy, the step will automatically skipped. Otherwise, step become pending/blocked (normal). | NO |
| `strategy` | `string` | Execution behaviour of the step. Possible values are `auto` or `manual` | YES |
| `rejection_reason` | `string` | This fills `Approval.Reason` if current approval step gets rejected based on `ApproveIf` expression. If `strategy=manual`, this field ignored. | NO |
| `approvers` | `[]string` | List of email or [`Expression`](#expression) string. The `Expression` is expected to return an email address or list of email addresses. | YES if `strategy` is `manual` | |
| `approve_if` | [`Expression`](#expression) | Determines the automatic resolution of current step when `strategy` is `auto` | YES if `strategy` is `auto` |
| `allow_failed` | `boolean` | If `true`, and current step is rejected, it will mark the appeal status as skipped instead of rejected | NO |

### `PolicyAppealConfig`

| Field            | Type                      | Description              | Required |
|:-----------------|:--------------------------|:-------------------------|:---------|
| `duration_options` | [[]object(DurationOptions)](#durationoptions) | list of duration options | NO       | 
| `allow_permanent_access`| `boolean` | Set this to true if you want to allow users to have permanent access to the resources. Default: false | No |
| `allow_active_access_extension_in` | `string` | Duration before the access expiration date when the user allowed to create appeal to the same resource \(extend their current access\).<br/> Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. Reference: [ParseDuration](https://pkg.go.dev/time#ParseDuration) | No |

### `DurationOptions`

| Field | Type   | Description                                                                                                                                                                                                                              | Required |
|:------|:-------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------|
| name  | string | name of duration                                                                                                                                                                                                                         | YES      |
| value | string | actual value of duration such as `24h`, `72h`. value will be `0h` in case of permanent duration. <br/> Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. Reference: [ParseDuration](https://pkg.go.dev/time#ParseDuration) | YES      |

### `IAM`

| Field | Type | Description | Required |
| :----- | :---- | :------ | :------ |
| `provider` | `string` | Identity manager type. Supported types are `http` and `shield` | YES |
| `config` | `object`| Client configuration according to the `provider` type | YES |
| `schema` | `map<string,string>` | User (appeal creator) profile details schema to be shown in the `creator` field in an appeal | NO |

### `Requirement`

| Field | Type | Description | Required |
| :----- | :---- | :------ | :------ |
| `on` | `object` | Criteria or conditions based on the current appeal to check before creating additional appeals | YES |
| `on.provider_type` | `string` | Criteria for the provider type of the current appeal's selected resource. Regex supported | NO |
| `on.provider_urn` | `string` | Criteria for the provider URN of the current appeal's selected resource. Regex supported | NO |
| `on.resource_type` | `string` | Criteria for the resource type of the current appeal's selected resource. Regex supported | NO |
| `on.resource_urn` | `string` | Criteria for the resource type of the current appeal's selected resource. Regex supported | NO |
| `on.role` | `string` | Criteria for the role of the current appeal. Regex supported | NO |
| `appeals` | `[]object` | List of additional appeals that will automatically created when `on` criteria is fulfilled| YES |
| `appeals[].resource` | `object` | Resource selector | YES |
| `appeals[].resource.id` | uint | Resource selector using the resource unique identifier | NO |
| `appeals[].resource.provider_type` | `string` | Resource selector using `provider_type`, `provider_urn`, `type`, and `urn`. Required if `appeals[].resource.id` is not present | NO |
| `appeals[].resource.provider_urn` | `string` | Resource selector using `provider_type`, `provider_urn`, `type`, and `urn`. Required if `appeals[].resource.id` is not present | NO |
| `appeals[].resource.type` | `string` | Resource selector using `provider_type`, `provider_urn`, `type`, and `urn`. Required if `appeals[].resource.id` is not present | NO |
| `appeals[].resource.urn` | `string` | Resource selector using `provider_type`, `provider_urn`, `type`, and `urn`. Required if `appeals[].resource.id` is not present | NO |
| `appeals[].role` | `string` | Role/permission to be assigned to the `account_id` of the current appeal to access the resource specified in the `resource` selector field | YES |
| `appeals[].policy` | `object` | Policy selector to be used for overriding the original policy linked to the resource specified in the `resource` selector field | NO |
| `appeals[].policy.id` | `string` | Policy identifier | YES |
| `appeals[].policy.version` | `uint` | Policy version identifier. Used together with `appeals[].policy.id` to reference to a policy | YES |

### `Expression`

Expression is an evaluatable statement intented to make the policy highly flexible. Guardian
uses https://github.com/antonmedv/expr to parse expressions. There's also some accessible variables specific to Guardian
use cases:

#### Variables

1. `$appeal`: [`Appeal`](appeal.md#appeal-1)

   Usage example:
    * `$appeal.resource.id` =&gt; `1`
    * `$appeal.resource.details.owners` =&gt; `["owner@email.com", "another.owner@email.com"]`
    * `$appeal.resource.labels.key` =&gt; `"value"`
    * `$appeal.creator.manager_email` =&gt; `"manager@email.com"`
