# Tableau

## Tableau

Tableau empowers everyone to see and understand the data. It is business intelligent for an entire organization. We can connect to any data source, be it a spreadsheet, database or bigdata. We can access data warehouses or cloud data as well.

### Tableau resources

* **Sites** In Tableau-speak, we use site to mean a collection of users, groups, and content \(workbooks, data sources\) that’s walled off from any other groups and content on the same instance of Tableau Server. Another way to say this is that Tableau Server supports multi-tenancy by allowing server administrators to create sites on the server for multiple sets of users and content. All server content is published, accessed, and managed on a per-site basis. Each site has its own URL and its own set of users \(although each server user can be added to multiple sites\). Each site’s content \(projects, workbooks, and data sources\) is completely segregated from content on other sites.
* **Projects** act as folder in tableau. A content resource \(workbooks and data sources\) can live in only project.
* **Workbooks** in tableau are a collection of views, metrics and data sources. Guardian supports access at all the levels i.e. workbook, metrics and data sources. Workbooks have options to show or hide tabs. If it is shown, permissions to the resources below are only **inherited** from the workbook level. If it is hidden, permissions can be given at the view/metric/data source level.
* **Views** are a visualization or viz that you create in Tableau. A viz might be a chart, a graph, a map, a plot, or even a text table. Access can be granted at view level only if the parent workbook has tabs option set to hidden.
* **Metrics** are new type of content that is fully integrated with Tableau's data and analytics platform through Tableau Server and Tableau Online. Metrics update automatically and display the most recent value. Access can be granted at metric level only if the parent workbook has tabs option set to hidden.
* **Data Sources** can be published to Tableau Server when your Tableau users want to share data connections they’ve defined. When a data source is published to the server, other users can connect to it from their own workbooks, as they do other types of data. When the data in the Tableau data source is updated, all workbooks that connect to it pick up the changes. Access can be granted at data source level only if the parent workbook has tabs option set to hidden.
* **Flows** are created to schedule tasks to run at a specific time or on a recurring basis. Access can be directly granted at a flow level.

### Tableau Users

Tableau allows to group users into groups and manage group level access to the resources. But, Guardian allows direct user level access to any resource.

## Authentication

Guardian requires **host**, **email**, **password** and **content url** of an administrator user in Tableau.

Example provider config for tableau:

```yaml
...
credentials:
  host: https://prod-apnortheast-a.online.tableau.com
  username: user@test.com
  password: password@123
  content_url: guardiantestsite
...
```

## Access Management

In Guardian, user access can be given at the workbook, views, metrics, data sources or flow level.



## 1. Config

### Example

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

## `Tableau Credentials`

| Fields | Deatils |
| :--- | :--- |
| `host` | `string`   Required. Tableau instance host.   Example: `https://prod-apnortheast-a.online.tableau.com` |
| `username` | `email`   Required. Email address of an account that has Administration permission. |
| `password` | `string`   Required. Account's password. |
| `content_url` | `string`   Required. Site's content url aka slug.   Example: In `https://10ay.online.tableau.com/#/site/MarketingTeam/workbooks` the content url is `MarketingTeam` |

## `Grafana Resource Type`

* `Workbook`
* `View`
* `Metric`
* `Data Source`
* `Flow`

## `Tableau Permissions`

| Fields | Permissions |
| :--- | :--- |
| `Workbook` | AddComment, ChangeHierarchy, ChangePermissions, Delete, ExportData, ExportImage, ExportXml, Filter, Read \(view\), ShareView, ViewComments, ViewUnderlyingData, WebAuthoring, and Write. |
| `View` | AddComment, ChangePermissions, Delete, ExportData, ExportImage, ExportXml, Filter, Read \(view\), ShareView, ViewComments, ViewUnderlyingData, WebAuthoring, and Write. |
| `Metric` | Read,Write,Delete,ChangeHierarchy,ChangePermissions. |
| `Data Source` | ChangePermissions, Connect, Delete, ExportXml, Read \(view\), and Write. |
| `Flow` | ChangeHierarchy, ChangePermissions, Delete, Execute, ExportXml \(Download\), Read \(view\), and Write. |
| `Site Roles` | Creator, Explorer, ExplorerCanPublish, ServerAdministrator, SiteAdministratorExplorer, SiteAdministratorCreator, Unlicensed, ReadOnly, or Viewer. |

## `Table Resource Permission`

| Fields | Type | Details |
| :--- | :--- | :--- |
| `urn` | Required.   `string` | Tableau Site Id. |
| `resources: type` | Required.   `string` | Must be one of `workbook, view, metric, datasource and flow`. |
| `resources: policy` | Required.   `string & string` | Must have id as policy name.   Must have a version number. |
| `resources: roles` | Required.   `string ,string & permissions` | Must have a role id .   Must have a role name.   Must have a list of permissions required. |
| `resources: roles: permissions` | Required.   `string & string` | Must have a name in format `<permission-name>:<permission-mode>` or just `<permission-name>` in case of site role .   `Optional:` If this is a site role, it should have a type attribute with value always equal to `site_role`. |

