# Managing resources

## Collecting Resources

Guardian collects resources from the providers automatically as soon as it registered. It uses cronjob to fetch the resource data continuously.

Example resource object:

```javascript
{
  "id": 1,
  "provider_type": "google_bigquery",
  "provider_urn": "gcp-project-id",
  "type": "table",
  "urn": "gcp-project-id:dataset_name.table_name",
  "name": "table_name",
  "metadata": {
    "owners": [
      "owner@email.com"
    ]
  },
  "labels": {
    "key": "value"
  }
}
```

You can see all the resources by using this endpoint:

```text
GET /resources
Accept: application/json

Response:
[
  {
    "id": 1,
    "provider_type": "google_bigquery",
    "provider_urn": "gcp-project-id",
    "type": "table",
    "urn": "gcp-project-id:dataset_name.table_name",
    "name": "table_name",
    "metadata": {
      "owners": [
        "owner@email.com"
      ]
    },
    "labels": {
      "key": "value"
    }
  }
]
```

## Adding metadata to resources

Guardian also still allows user to add their own metadata or any additional information into the resources.

This can be useful when we configuring the approval policy, and we need information from metadata e.g. “owners” as the approvers.

Endpoint:

```text
PUT /resources/:id
Content-Type: application/json
Accept: application/json

Request Body:
{
  "details": {
    "owners": [
      “user@email.com”
    ],
    “key”: “value”
  }
}

Response:
{
  "id": 1,
  "provider_type": "google_bigquery",
  "provider_urn": "gcp-project-id",
  "type": "table",
  "urn": "gcp-project-id:dataset_name.table_name",
  "name": "table_name",
  "metadata": {
    "owners": [
      "user@email.com"
    ],
    “key”: “value”
  }
}
```

