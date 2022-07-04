# Provider

A provider configuration is required when we want to register a provider instance to Guardian.

#### YAML representation

```yaml
type: string
urn: string
credentials: object
appeal: object
resources: []object
```

### `ProviderConfig`
| Field | Type | Description | Required | 
| :----- | :---- | :------ | :------ | 
| `type`| `string` | This field conatains the name of the Resource Provider<br/><br/> Possible values can be:<br/> - BigQuery : **google_bigquery** <br/> - Tableau : **tableau** <br/> - Grafana : **grafana** <br/> - Metabase : **metabase** <br/> - Google Cloud IAM : **gcloud_iam**| Yes |
| `urn`| `string` | Provider instance identifier   | Yes | 
| `allowed_account_types` | `[string]` | Optional. List of allowed account types. Each provider could have different account types, but `user` account type is applicable for any provider type | No | 
| `credentials` | `object`| Credentials required to setup connection and access the provider <br/> <br/>  Possible values: <br/> BigQuery: [object(BigQuery)](../providers/bigquery.md#bigquerycredentials) <br/> Metabase: [object(Metabase)](../providers/metabase.md#metabasecredentials) <br/>Tableau: [object(Tableau)](../providers/tableau.md#tableau-credentials)<br/>Grafana:[object(Grafana)](../providers/grafana.md#grafanacredentials)<br/>Google Cloud IAM: [object(GCloudIAM)](../providers/gcloud_iam.md#gcloudiamcredentials)| Yes | 
| `appeal`      | [`object(AppealConfig)`](provider.md#appealconfig) | Contains details of the tenure for which an access for a resource is provided. Contains two fields `allow_permanent_access` and `allow_active_access_extension_in` for permanent access and time before which the user can appeal for an extention | Yes | 
| `resources` | [`[object(ResourceConfig)]`](provider.md#resourceconfig) | Contains the configurations for each resource . The fields `type` and `policy` stores the type of resource and the policy associated with it. `Roles` conatins the role (say Viewer, Editor, Writer) which the resource supports | Yes |


### `AppealConfig`

| Field | Type | Description | Required | 
| :----- | :---- | :------ | :------ | 
| `allow_permanent_access`| `boolean` | Set this to true if you want to allow users to have permanent access to the resources. Default: false | No |
| `allow_active_access_extension_in` | `string` | Duration before the access expiration date when the user allowed to create appeal to the same resource \(extend their current access\). | No |

### `ResourceConfig`

| Field | Type | Description | Required | 
| :----- | :--------- | :---------- | :------ | 
| `type`    | `string` | Possible values for the Resource Type:<br/> - BigQuery: [string(BigQuery)](../providers/bigquery.md#bigqueryresourcetype) <br/> - Metabase: [string(Metabase)](../providers/metabase.md#metabaseresourcetype) <br/> - Graffana: [string(Graffana)](../providers/grafana.md#grafanaresourcetype) <br/> - Tableau: [string(Tableau)](../providers/tableau.md#grafana-resource-type) <br/> - Google Cloud IAM: [string(GCloudIAM)](../providers/gcloud_iam.md#gcloudiamresourcetype) | Yes|
| `policy`  | `object(id: string, version: int)` | Approval policy config that want to be applied to this resource config. Example: `id: approval_policy_x, version: 1` |Yes|
| `roles[]` | [`object(Role)`](provider.md#role) |List of resource permissions mapping|Yes|

### `Role`

| Field | Type | Description | Required | 
| :----- | :---- | :------ | :------ | 
| `id` | `string` | Role identifier| Yes |
| `name`| `string` | Display name for role| |
| `permissions[]` | `object or string` | Set of permissions that will be granted to the requested resource.<br/> Possible values for Resource Permissions :<br/> - BigQuery: [object(BigQuery)](../providers/bigquery.md#bigqueryresourcepermission) <br/>- Metabase: [object(Metabase)](../providers/metabase.md#metabaseresourcepermission) <br/>- Grafana: [object(Grafana)](../providers/grafana.md#grafanaresourcepermission) <br/>- Tableau: [object(Tableau)](../providers/tableau.md#table-resource-permission) <br/> - Google Cloud IAM: object(GCloudIAM)| Yes |