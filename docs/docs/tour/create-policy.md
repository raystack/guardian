import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Create a policy

### Pre-Requisites

1. [Setting up server](./configuration.md#starting-the-server)
2. [Setting up the CLI](./configuration.md#client-configuration) (if you want to create policy using CLI)

### Example Policy

```yaml
id: my-first-policy
steps:
  - name: resource_owner_approval
    description: approval from resource owner
    strategy: manual
    approvers:
      - $appeal.resource.details.owner
  - name: admin_approval
    description: approval from admin (John Doe)
    strategy: manual
    approvers:
      - john.doe@company.com
appeal:
  - duration_options:
    - name: 1 day
      value: 24h
    - name: 1 week
      value: 98h
  - allow_on_behalf: false
```

Check [policy reference](../reference/policy.md) for more details on the policy configuration.<br/>

**Explanation of this Policy example**<br/>
When a Guardian user creates an appeal to the BigQuery resource (Playground here), this policy will applied, and the approvals required to approve that appeal are in the order as follows: <br/>

1. Approval from the resource owner ( this information is contained in the resource details object), and
2. Approval from John Doe as an admin

#### Policies can be created in the following ways:

1. Using `guardian policy create` CLI command
2. Calling to `POST /api/v1beta1/policies` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```bash
$ guardian policy create --file=<path to the policy.yaml file>
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```bash
$ curl --request POST '{{HOST}}/api/v1beta1/policies' \
--header 'Content-Type: application/json' \
--data-raw '{
  "id": "my-first-policy",
  "steps": [
    {
      "name": "resource_owner_approval",
      "description": "Approval from Resource owner",
      "strategy": "manual",
      "approvers": [
        "$appeal.resource.details.owner"
      ]
    },
    {
      "name": "admin_approval",
      "description": "Approval from the Admin (John Doe)",
      "strategy": "manual",
      "approvers": [
        "john.doe@company.com"
      ]
    }
  ],
   "appeal": {
        "duration_options": [
            {
                "name": "1 Day",
                "value": "24h"
            },
            {
                "name": "3 Day",
                "value": "72h"
            }
        ],
        "allow_on_behalf": true
    }
}'
```

  </TabItem>
</Tabs>

**Note** : For using the CLI tool, create a Policy.yaml file using the example configurations shown above and provide the path to it here.
