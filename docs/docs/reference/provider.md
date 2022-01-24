# Provider Configurations

A provider configuration is required when we want to register a provider instance to Guardian.

#### YAML representation

```yaml
type: string
urn: string
credentials: any
appeal: object
resources: []object
```

| Fields        |                                                                                                                                                                                          |
| :------------ | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`        | `string` Required. Provider type Possible values: `google_bigquery`, `metabase`                                                                                                          |
| `urn`         | `string` Required. Provider instance identifier                                                                                                                                          |
| `allowed_account_types` | `[]string` Optional. List of allowed account types. Each provider could have different account types, but `user` account type is applicable for any provider type |
| `credentials` | `object` Required. Credentials to setup connection and access the provider |
| `appeal`      | [`object(AppealConfig)`](provider.md#appealconfig) Required. Appeal options                                                                                                       |
| `resources[]` | [`object(ResourceConfig)`](provider.md#resourceconfig) Required. List of permission configurations for each resource type                                                         |

### `AppealConfig`

| Fields                             |                                                                                                                                                  |
| :--------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------- |
| `allow_permanent_access`           | `boolean` Set this to true if you want to allow users to have permanent access to the resources. Default: `false`                                |
| `allow_active_access_extension_in` | `string` Duration before the access expiration date when the user allowed to create appeal to the same resource \(extend their current access\). |

### `ResourceConfig`

| Field     |                                                                                                                                                                   |
| :-------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`    | `string` Required. Possible values: - BigQuery: `string(BigQueryResourceType)` - Metabase: `string(MetabaseResourceType)`                                         |
| `policy`  | `object(id: string, version: int)` Required. Approval policy config that want to be applied to this resource config. Example: `id: approval_policy_x, version: 1` |
| `roles[]` | [`object(Role)`](provider.md#role) Required. List of resource permissions mapping                                                                          |

### `Role`

| Fields          |                                                                                                                                                                                                                                                                                                                 |
| :-------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`            | `string` Required. Role identifier                                                                                                                                                                                                                                                                              |
| `name`          | `string` Display name for role                                                                                                                                                                                                                                                                                  |
| `permissions[]` | `object or string` Required. Set of permissions that will be granted to the requested resource Possible values: - BigQuery: `object(BigQueryResourcePermission)` - Metabase: `object(MetabaseResourcePermission)` - Grafana: `object(GrafanaResourcePermission)` - Tableau: `object(TableauResourcePermission)` |

## Providers

Here are the available providers in Guardian. Currently we only have Google BigQuery, but we will ad more soon.

| Google BigQuery          |                                                                                                                                                     |
| :----------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------- |
| Provider type            | `google_bigquery`                                                                                                                                   |
| Credentials value        | 1) service_account_key - Base64 encrypted value of a service account key JSON 2) resource_name - Name of bigquery project appended with "projects/" |
| Available resource types | `dataset`, `table`                                                                                                                                  |
