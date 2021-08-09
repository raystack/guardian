# Grafana Provider

## 1. Config

#### Example

```yaml
type: grafana
urn: 1
labels:
 entity: gojek
 landscape: id
credentials:
  host: http://localhost:4000
  username: admin@localhost
  password: password
appeal:
 allow_permanent_access: true
 allow_active_access_extension_in: "7d"
resources:
  - type: dashboard
    policy:
      id: policy_x
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: view
      - id: editor
        name: Editor
        permissions:
          - name: edit
      - id: admin
        name: Admin
        permissions:
          - name: admin
```

### `GrafanaCredentials`

Fields ||
-|-
`host` | `string` <br> Required. Grafana instance host. <br> Example: `http://localhost:3000`
`username` | `email` <br> Required. Email address of an account that has Administration permission.
`password` | `string` <br> Required. Account's password.

### `GrafanaResourceType`

- `folders`
- `dashboards` - Direct dashboard level access via Guardian.

### `GrafanaResourcePermission`

Fields ||
-|-
`urn` | `int` <br> Required. <br> Grafana Organisation Id.
`resources: type` | `string` <br> Required. <br> Must be `dashboard`.
`resources: roles` | `string` <br> Required. <br> Must have id one of `viewer`, `editor` or `admin`. <br> Must have name one of `Viewer`, `Editor` or `Admin`. <br> Must have permissions one of `view`, `edit` or `admin`.
