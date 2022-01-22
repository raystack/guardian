# Managing Resources

Resource in Guardian represents the actual resource in the provider e.g. for BigQuery provider, a resource represents a dataset or a table. One of Guardian's responsibility is to manage the access to resources, and in order to do that Guardian needs to be able to manage the resources as well.

## Collecting Resources

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

## Updating Resources Metadata

Update a resource can be done in the following ways:
1. [Using `guardian resource set` CLI command](#1-update-a-resource-using-cli)
2. [Calling to `PUT /api/v1beta1/resources/:id` API](#2-update-a-resource-using-http-api)

Guardian allows users to add metadata to the resources. This can be useful when configuring the approval steps in the policy that needs information from metadata e.g. “owners” as the approvers.

### 1. Update a Resource using CLI
```console
$ guardian resource set --id={{resource_id}} --values=<key1>=<value1> --values=<key2>=<value2>
```
### 2. Update a Resource using HTTP CLI
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

## Listing Resources

Listing resources can be done in the following ways:
1. [Using `guardian resource list` CLI command](#1-list-resources-using-cli)
2. [Calling to `GET /api/v1beta1/resources` API](#2-list-resources-using-http-api)

### 1. List Resources using CLI
```console
$ guardian resource list --output=yaml
```

### 2. List Resources using HTTP API
```console
$ curl --request GET '{{HOST}}/api/v1beta1/resources'
```

## Viewing Resources

Viewing a resource can be done in the following ways:

1. [Using `guardian resource view` CLI command](#1-view-a-resource-using-cli)
2. [Calling to `GET /api/v1beta1/resources/:id` API](#2-view-a-resource-using-http-api)

### 1. View a Resource using CLI
```console
$ guardian resource view {{resource_id}}
```

### 2. View a Resource using HTTP API
```console
$ curl --request GET '{{HOST}}/api/v1beta1/resources/{{resource_id}}'
```
