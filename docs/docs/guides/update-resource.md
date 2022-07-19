import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Update resource

We will try to update a resource information in this example exercise. Let's say we want to add owner's information to the `playground` dataset.

To update the resource metadata we will need it's id which is generated when the Provider registered it on the Guardian. First we need to check the `resource_id` of that `playground` dataset.

To get the list of all the resources use any of the following methods:

#### Getting the `resource_id` for the resource:

1. Using `guardian resource list` CLI command
2. Calling to `GET /api/v1beta1/resources` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```bash
$ guardian resource list --output=yaml
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```bash
$ curl --request GET '{{HOST}}/api/v1beta1/resources'
```

  </TabItem>
</Tabs>

You can use `resource_id` to get the resource details by any of the following commands:

1. Using `guardian resource view` CLI command
2. Calling to `GET /api/v1beta1/resources/:id` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```bash
$ guardian resource view {{resource_id}}
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```bash
$ curl --request GET '{{HOST}}/api/v1beta1/resources/{{resource_id}}'
```

  </TabItem>
</Tabs>

To update the resource metadata with this information add this to the resource file or request body

```yaml
details:
  owner: owner.guy@company.com
```

Use any of these commands to update the owner details:

1. Using `guardian resource set` CLI command
2. Calling to `PUT /api/v1beta1/resources/:id` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```bash
$ guardian resource set <resource_id> -f resource.yaml
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```bash
$ curl --request PUT '{{HOST}}/api/v1beta1/resources/{{resource_id}}' \
--header 'Content-Type: application/json' \
--data-raw '{
    "details": {
        "owner": "owner.guy@company.com"
    }
}'
```

  </TabItem>
</Tabs>
