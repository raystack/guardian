import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Create BigQuery Provider

We are going to register a Google Cloud Bigquery provider with a dataset named `Playground` in this example.

### Pre-Requisites

A service account with `roles/bigquyer.dataOwner` role granted to the Google cloud project. Encode the Service Account key in the base 64 format, get the project id from the BigQuery Project you've created and provide the details in the `Credentials` object given below.

### Example Provider Configuartion

```yaml
type: bigquery
urn: my-first-bigquery-provider
credentials:
  resource_name: projects/<<my-bq-project-id>> # projects/<<gcp project id>>
  service_account_key: <<base-64 encoded service account key json>> # Encode the service account key in base 64 form
resources:
  - type: table
    policy:
      id: my-first-policy
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - roles/bigquery.dataViewer
      - id: editor
        name: Editor
        permissions:
          - roles/bigquery.dataViewer
      - id: owner
        name: Owner
        permissions:
          - roles/bigquery.dataOwner
  - type: dataset
    policy:
      id: my-first-policy
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - READER
      - id: editor
        name: Editor
        permissions:
          - WRITER
      - id: owner
        name: Owner
        permissions:
          - OWNER
```

Check [BigQuery](../providers/bigquery.md) provider reference for more details.

**Explanation for this Provider Configuration**<br/>

Here we are registering the Bigquery provider with the Service account credentials. These credentials are used by Guardian server to communicate to the provider (in this case is `my-bq-project` bigquery) to retrieve the available resources (table and datasets) as well as managing access to them.

Here we are registering a BigQuery Provider with two types of resources `Dataset` and `Table` Each Resource has a policy along with its version attached to it.

We have configured the resource type `table` with this policy [`my-first-policy@1`](create-policy#example-policy). Every appeal created to for this resource type under `my-first-bigquery-provider` provider, will have approval steps according to the policy defined here [`my-first-policy@1`](./create-policy.md#example-policy).

The `Roles` field is used to define what type of permission a user have, be it `Editor`,`Viewrer` or `Owner` in the BigQuery dataset.

To check all the available roles for a particular resource type use the API `{{HOST}}/api/v1beta1/providers/:id/resources/:resource_type/roles` with the `GET` Method.

### Registering the BigQuery Provider

#### Providers can be created in the following ways:

1. Using `guardian provider create` CLI command
2. Calling to `POST /api/v1beta1/providers` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```bash
$ guardian provider create --file=provider.yaml
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```json
$ curl --request POST '{{HOST}}/api/v1beta1/providers' \
--header 'Content-Type: application/json' \
--data-raw '{
  "type": "bigquery",
  "urn": "my-first-bigquery-provider",
  "credentials": {
    "service_account_key": "{{base64 encoded service account key json}}",
    "resource_name": "projects/my-bq-project"
  },
  "resources": [
    {
      "type": "table",
      "policy": {
        "id": "my-first-policy",
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
            "roles/bigquery.dataViewer"
          ]
        },
        {
          "id": "owner",
          "name": "Owner",
          "permissions": [
            "roles/bigquery.dataOwner"
          ]
        }
      ]
    },
    {
      "type": "dataset",
      "policy": {
        "id": "my-second-policy",
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
          "name": "EDITOR",
          "permissions": [
            "WRITER"
          ]
        },
        {
          "id":"owner",
          "name":"OWNER",
          "permissions":[
            "OWNER"
          ]
        }
      ]
    }
  ]
}'
```

  </TabItem>
</Tabs>
