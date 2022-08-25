import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Manage appeal

Note: Approve/reject still not supported from the CLI currently.

#### Appeals can be approved/rejected in the following ways:

1. Using `guardian appeal approve/reject` CLI command
2. Calling to `POST /api/v1beta1/appeals/:id/approvals/:approval_step_name/` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

#### Approve an Appeal

```bash
$ guardian appeal approve --id={{appeal_id}} --step={{approval_step_name}}
```

#### Reject an Appeal

```bash
$ guardian appeal reject --id={{appeal_id}} --step={{approval_step_name}} --reason={{rejection_message}}
```

  </TabItem>
  <TabItem value="http" label="HTTP">

#### Approve an Appeal

```bash
$ curl --request POST '{{HOST}}/api/v1beta1/appeals/{{appeal_id}}/approvals/{{approval_step_name}}' \
--header 'X-Auth-Email: user@example.com' \
--header 'Content-Type: application/json' \
--data-raw '{
    "action": "approve"
}'
```

#### Reject an Appeal

```bash
$ curl --request POST '{{HOST}}/api/v1beta1/appeals/{{appeal_id}}/approvals/{{approval_step_name}}' \
--header 'X-Auth-Email: user@example.com' \
--header 'Content-Type: application/json' \
--data-raw '{
    "action": "reject",
    "reason": "{{rejection_message}}"
}'
```

  </TabItem>
</Tabs>
