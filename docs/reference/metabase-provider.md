# Metabase Provider

## 1. Config

#### Example

```yaml
type: metabase
urn: my-metabase
credentials:
  host: http://localhost:12345
  user: administrator@email.com
  password: password123
appeal:
  allow_active_access_extension_in: '7d'
resources:
  - type: database
    policy:
      id: policy_id
      version: 1
    roles:
      - id: read
        name: Read
        permissions:
          - name: schemas:all
      - id: query
        name: SQL Query
        permissions:
          - name: schemas:all
          - name: native:write
  - type: collection
    policy:
      id: policy_id
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - name: read
      - id: editor
        name: Editor
        permissions:
          - name: write
```

### `MetabaseCredentials`

Fields ||
-|-
`host` | `string` <br> Required. Metabase instance host <br> Example: `http://localhost:12345`
`username` | `email` <br> Required. Email address of an account that has Administration permission
`password` | `string` <br> Required. Account's password

### `MetabaseResourceType`

- `database`
- `collection`

### `MetabaseResourcePermission`

Fields ||
-|-
`name` | `string` <br> Required. Metabase permission mapping <br><br> **Possible values:** <br> - `database`: `schemas:all` (read table), `native:write` (run SQL query) <br> **Note**: Metabase requires `schemas:all` permission for `native:write` to be able to work <br> - `collection`: `read`, `write`
