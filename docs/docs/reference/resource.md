# Resource

#### JSON Representation

```json
{
  "id": 1,
  "provider_type": "bigquery",
  "provider_urn": "my-bigquery",
  "type": "table",
  "urn": "gcp-project-id:dataset_name.table_name",
  "name": "table_name",
  "details": {
    ...
  },
  "created_at": "2021-10-26T09:29:48.838203Z",
  "updated_at": "2021-10-26T09:29:48.838203Z"
}
```

### `Resource`

| Fields | |
| :--- | :--- |
| `id` | `uint` Resource unique identifier |
| `provider_type` | `string` Type of the provider that manages this resource | 
| `provider_urn` | `string` Provider URN |
| `type` | `string` Type of the resource according to `provider_type` |
| `urn` | `string` Resource URN |
| `name` | `string` Display name |
| `details` | `object` Additional information of the resource that can be updated from Guardian |
| `created_at` | `string` Timestamp when the resource created |
| `updated_at` | `string` Timestamp when the resource last modified |
