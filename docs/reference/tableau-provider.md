# Tableau Provider

### 1. Config

#### Example

```yaml
type: tableau
urn: 691acb66-27ef-4b4f-9222-f07052e6ffd0
labels:
  entity: gojek
  landscape: id
credentials:
  host: https://prod-apnortheast-a.online.tableau.com
  username: test@email.com
  password: password@123
  content_url: guardiantestsite
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: workbook
    policy:
      id: policy_1
      version: 1
    roles:
      - id: read
        name: Read
        permissions:
          - name: Read:Allow
          - name: ViewComments:Allow
          - name: ViewUnderlyingData:Allow
          - name: Filter:Allow
          - name: Viewer
            type: site_role
      - id: write
        name: Write
        permissions:
          - name: Write:Allow
          - name: AddComment:Allow
          - name: Creator
            type: site_role
      - id: admin
        name: Admin
        permissions:
          - name: ChangeHierarchy:Allow
          - name: ChangePermissions:Allow
          - name: Delete:Allow
          - name: ServerAdministrator
            type: site_role
      - id: export
        name: Export
        permissions:
          - name: ExportData:Allow
          - name: ExportImage:Allow
          - name: ExportXml:Allow
          - name: SiteAdministratorExplorer
            type: site_role
      - id: other
        name: Other
        permissions:
          - name: ShareView:Allow
          - name: WebAuthoring:Allow
          - name: ExplorerCanPublish
            type: site_role
  - type: flow
    policy:
      id: policy_2
      version: 1
    roles:
      - id: read
        name: Read
        permissions:
          - name: Read:Allow
          - name: Viewer
            type: site_role
      - id: write
        name: Write
        permissions:
          - name: Write:Allow
          - name: Creator
            type: site_role
      - id: admin
        name: Admin
        permissions:
          - name: ChangeHierarchy:Allow
          - name: ChangePermissions:Allow
          - name: Delete:Allow
          - name: ServerAdministrator
            type: site_role
      - id: export
        name: Export
        permissions:
          - name: ExportXml:Allow
          - name: SiteAdministratorExplorer
            type: site_role
      - id: other
        name: Other
        permissions:
          - name: Execute:Allow
          - name: ExplorerCanPublish
            type: site_role   
```

### `Tableau Credentials`

Fields | Deatils|
-|-
`host` | `string` <br> Required. Tableau instance host. <br> Example: `https://prod-apnortheast-a.online.tableau.com`
`username` | `email` <br> Required. Email address of an account that has Administration permission.
`password` | `string` <br> Required. Account's password.
`content_url` | `string` <br> Required. Site's content url aka slug. <br> Example: In `https://10ay.online.tableau.com/#/site/MarketingTeam/workbooks` the content url is `MarketingTeam`

### `Grafana Resource Type`

- `Workbook`
- `View`
- `Metric`
- `Data Source`
- `Flow`

### `Tableau Permissions`

Fields |Permissions|
-|-
`Workbook` | AddComment, ChangeHierarchy, ChangePermissions, Delete, ExportData, ExportImage, ExportXml, Filter, Read (view), ShareView, ViewComments, ViewUnderlyingData, WebAuthoring, and Write.
`View` | AddComment, ChangePermissions, Delete, ExportData, ExportImage, ExportXml, Filter, Read (view), ShareView, ViewComments, ViewUnderlyingData, WebAuthoring, and Write.
`Metric` | Read,Write,Delete,ChangeHierarchy,ChangePermissions.
`Data Source` | ChangePermissions, Connect, Delete, ExportXml, Read (view), and Write.
`Flow` | ChangeHierarchy, ChangePermissions, Delete, Execute, ExportXml (Download), Read (view), and Write.
`Site Roles` | Creator, Explorer, ExplorerCanPublish, ServerAdministrator, SiteAdministratorExplorer, SiteAdministratorCreator, Unlicensed, ReadOnly, or Viewer.



### `Table Resource Permission`

Fields |Type| Details|
-|-|-
`urn` | Required. <br> `string` | Tableau Site Id.
`resources: type` | Required. <br> `string` | Must be one of `workbook, view, metric, datasource and flow`.
`resources: policy` | Required. <br> `string & string` | Must have id as policy name. <br> Must have a version number.
`resources: roles` | Required. <br> `string ,string & permissions` | Must have a role id . <br> Must have a role name. <br> Must have a list of permissions required.
`resources: roles: permissions` | Required. <br> `string & string` | Must have a name in format `<permission-name>:<permission-mode>` or just `<permission-name>` in case of site role . <br> `Optional:` If this is a site role, it should have a type attribute with value always equal to `site_role`.
