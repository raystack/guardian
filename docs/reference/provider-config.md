# Provider Configuration

A provider configuration is required when we want to register a provider instance to Guardian.

#### YAML representation
```yaml
type: string
urn: string
credentials: any
appeal: object
resources: []object
```

Fields ||
-|-
`type` | `string` <br> Required. Provider type<br><br> Possible values: `google_bigquery`, `metabase`
`urn` | `string` <br> Required. Provider instance identifier
`credentials` | `object` <br> Required. Credentials to setup connection and access the provider instance <br><br> Possible values: <br> - BigQuery: [`string(BigQueryCredentials)`](bigquery-provider.md#bigquerycredentials) <br> - Metabase: [`object(MetabaseCredentials)`](metabase-provider.md#metabasecredentials) 
`appeal` | [`object(AppealConfig)`](#appealconfig) <br> Required. Appeal options
`resources[]` | [`object(ResourceConfig)`](#resourceconfig) <br> Required. List of permission configurations for each resource type

### `AppealConfig`

Fields ||
-|-
`allow_permanent_access` | `boolean` <br> Set this to true if you want to allow users to have permanent access to the resources. Default: `false`
`allow_active_access_extension_in` | `string` <br> Duration before the access expiration date when the user allowed to create appeal to the same resource (extend their current access).

### `ResourceConfig`

Field ||
-|-
`type` | `string` <br> Required. <br><br> Possible values: <br> - BigQuery: [`string(BigQueryResourceType)`](bigquery-provider.md#bigqueryresourcetype) <br> - Metabase: [`string(MetabaseResourceType)`](metabase-provider.md#metabaseresourcetype)
`policy` | `object(id: string, version: int)` <br> Required. Approval policy config that want to be applied to this resource config. Example: `id: approval_policy_x, version: 1`
`roles[]` | [`object(RoleConfig)`](#roleconfig) <br> Required. List of resource permissions mapping

### `RoleConfig`

Fields ||
-|-
`id` | `string` <br> Required. Role identifier
`name` | `string` <br> Display name for role
`permissions[]` | `object` <br> Required. Set of permissions that will be granted to the requested resource <br><br> Possible values: <br> - BigQuery: [`object(BigQueryResourcePermission)`](bigquery-provider.md#bigqueryresourcepermission) <br> - Metabase: [`object(MetabaseResourcePermission)`](metabase-provider.md#metabaseresourcepermission)

## Providers

Here are the available providers in Guardian. Currently we only have Google BigQuery, but we will ad more soon.

| | Google BigQuery
|-|----------------
Provider type | `google_bigquery`
Credentials value | Base64 encrypted value of a service account key JSON
Available resource types | `dataset`, `table`

## Examples

- [Metabase](metabase-provider.md#example)
- [BigQuery](bigquery-provider.md#example)
- [Grafana](grafana-provider.md#example)
- [Tableau](tableau-provider.md#example)
