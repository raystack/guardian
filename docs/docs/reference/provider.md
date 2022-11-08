# Provider

A provider configuration is required when we want to register a provider instance to Guardian.

#### YAML representation

```yaml
id: "fcbfd47a-7dc4-4d3a-aff1-97ea7b205ac4"
type: "bigquery"
urn: "test-bq-urn"
config: 
  type: "bigquery"
  urn: "test-bq-urn"
  appeal:
    allow_permanent_access: false
    allow_active_access_extension_in: 24h
  resources:
    - type: "dataset"
      filter: $resource.name == 'playground'
      policy:
        id: "my-policy"
        version: 1
      roles:
        id: "viewer"
        name: "Viewer"
        permissions:
          - "READER"
  allowed_account_types:
    - user
created_at: "2021-10-26T09:29:48.838203Z"
updated_at: "2022-10-26T07:41:52.676004Z"
```

### `Provider`

| Field      | Type                                      | Description                                |
|------------|-------------------------------------------|--------------------------------------------|
| id         | `string`                                  | Provider unique identifier                 |
| type       | `string`                                  | Provider type                              |
| urn        | `string`                                  | Unique provider URN                        |
| config     | [object(ProviderConfig)](#providerconfig) | Provider Configuration                     |
| created_at | `string`                                  | Timestamp when the provider created.       |
| updated_at | `string`                                  | Timestamp when the provider last modified. |

### `ProviderConfig`
| Field | Type | Description | Required | 
| :----- | :---- | :------ | :------ | 
| `type`| `string` | This field conatains the name of the Resource Provider<br/><br/> Possible values can be:<br/> - BigQuery : **bigquery** <br/> - Google Cloud Storage : **gcs** <br/> - Tableau : **tableau** <br/> - Grafana : **grafana** <br/> - Metabase : **metabase** <br/> - Google Cloud IAM : **gcloud_iam** <br/> - No-Op : **noop**| Yes |
| `urn`| `string` | Provider instance identifier   | Yes | 
| `allowed_account_types` | `[string]` | Optional. List of allowed account types. Each provider could have different account types, but `user` account type is applicable for any provider type | No | 
| `credentials` | `object`| Credentials required to setup connection and access the provider <br/> <br/>  Possible values: <br/> BigQuery: [object(BigQuery)](../providers/bigquery.md#bigquerycredentials) <br/> Google Cloud Storage : [object(GCS)](../providers/gcs#gcs-credentials) <br/> Metabase: [object(Metabase)](../providers/metabase.md#metabasecredentials) <br/>Tableau: [object(Tableau)](../providers/tableau.md#tableau-credentials)<br/>Grafana:[object(Grafana)](../providers/grafana.md#grafanacredentials)<br/>Google Cloud IAM: [object(GCloudIAM)](../providers/gcloud_iam.md#gcloudiamcredentials)<br/>No-Op: `Nil`| Yes | 
| `appeal`      | [`object(AppealConfig)`](provider.md#appealconfig) | Contains details of the tenure for which an access for a resource is provided. Contains two fields `allow_permanent_access` and `allow_active_access_extension_in` for permanent access and time before which the user can appeal for an extention | Yes | 
| `resources` | [`[object(ResourceConfig)]`](provider.md#resourceconfig) | Contains the configurations for each resource . The fields `type` and `policy` stores the type of resource and the policy associated with it. `Roles` conatins the role (say Viewer, Editor, Writer) which the resource supports | Yes |
| `parameters` | [`object(ProviderParameter)`](#providerparameter) | Optional. Contains the parameters for the provider. | No |


### `AppealConfig`

| Field | Type | Description | Required | 
| :----- | :---- | :------ | :------ | 
| `allow_permanent_access`| `boolean` | Set this to true if you want to allow users to have permanent access to the resources. Default: false | No |
| `allow_active_access_extension_in` | `string` | Duration before the access expiration date when the user allowed to create appeal to the same resource \(extend their current access\). | No |

### `ResourceConfig`

| Field     | Type | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Required | 
|:----------| :--------- |:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------| 
| `type`    | `string` | Possible values for the Resource Type:<br/> - BigQuery: [string(BigQuery)](../providers/bigquery.md#bigqueryresourcetype) <br/> - Google Cloud Storage: [string(GCS)](../providers/gcs#gcs-resource-types)<br/> - Metabase: [string(Metabase)](../providers/metabase.md#metabaseresourcetype) <br/> - Graffana: [string(Graffana)](../providers/grafana.md#grafanaresourcetype) <br/> - Tableau: [string(Tableau)](../providers/tableau.md#grafana-resource-type) <br/> - Google Cloud IAM: [string(GCloudIAM)](../providers/gcloud_iam.md#gcloudiamresourcetype) <br/> - No-Op: [string(No-Op)](../providers/noop) | Yes      |
| `filter`  | `string` | Filter condition to add a specific set of resources match with condition. Example: `filter: $resource.name endsWith audit`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | No       |
| `policy`  | `object(id: string, version: int)` | Approval policy config that want to be applied to this resource config. Example: `id: approval_policy_x, version: 1`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | Yes      |
| `roles[]` | [`object(Role)`](provider.md#role) | List of resource permissions mapping                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | Yes      |

### `Role`

| Field | Type | Description | Required | 
| :----- | :---- | :------ | :------ | 
| `id` | `string` | Role identifier| Yes |
| `name`| `string` | Display name for role| |
| `permissions[]` | `object or string` | Set of permissions that will be granted to the requested resource.<br/> Possible values for Resource Permissions :<br/> - BigQuery: [object(BigQuery)](../providers/bigquery.md#bigqueryresourcepermission) <br/>- Google Cloud Storage: [object(GCS)](../providers/gcs#gcs-resource-permission) <br/>- Metabase: [object(Metabase)](../providers/metabase.md#metabaseresourcepermission) <br/>- Grafana: [object(Grafana)](../providers/grafana.md#grafanaresourcepermission) <br/>- Tableau: [object(Tableau)](../providers/tableau.md#table-resource-permission) <br/> - Google Cloud IAM: object(GCloudIAM) <br/>- No-Op : `Nil`| Yes |

### `ProviderParameter`

| Field | Type | Description | Required |
| :----- | :---- | :------ | :------ |
| `key` | `string` | The key is unique identifier for the parameter | Yes |
| `label` | `string` | The label is used to display the parameter in the UI | Yes |
| `required` | `boolean` | Indicates whether the parameter is required or not | Yes |
| `description` | `string` | The description of the parameter | No |