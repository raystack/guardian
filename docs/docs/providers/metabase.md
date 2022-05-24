# Metabase

Metabase is a data visualization tool that lets you connect to external databases and create charts and dashboards based on the data from the databases. Guardian supports access management to the following resources in Metabase:

1. Database
2. Collection

Metabase itself manages its user access on group-based permissions, while Guardian lets each individual user have access directly to the resources.

## Authentication

Guardian requires **email** and **password** of an administrator user in Metabase.

Example provider config for metabase:

```yaml
---
credentials:
  host: http://localhost:12345
  user: administrator@email.com
  password: password123
```

## Metabase Access Creation

Guardian creates a group that has only one permission type to one resource in Metabase Example: If a user wants to have **view** access to **database X** \(id=1\), Guardian will create a group called **database_1_view**, grant it with **view** permission only to **database X**, and then add the user to that group.

The group naming convention is:

```text
<resource_type>_<resource_id>_<permission_type/role>
```

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
  allow_active_access_extension_in: "7d"
resources:
  - type: database
    policy:
      id: policy_id
      version: 1
    roles:
      - id: read
        name: Read
        permissions:
          - schemas:all
      - id: query
        name: SQL Query
        permissions:
          - schemas:all
          - native:write
  - type: collection
    policy:
      id: policy_id
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - read
      - id: editor
        name: Editor
        permissions:
          - write
  - type: table
    policy:
      id: policy_id
      version: 1
    roles:
      - id: viewer
        name: Viewer
        permissions:
          - all
  - type: group
    policy:
      id: policy_id
      version: 1
    roles:
      - id: member
        name: Member
        permissions:
          - member
```

### `MetabaseCredentials`

| Fields     |                                                                                  |
| :--------- | :------------------------------------------------------------------------------- |
| `host`     | `string` Required. Metabase instance host Example: `http://localhost:12345`      |
| `username` | `email` Required. Email address of an account that has Administration permission |
| `password` | `string` Required. Account's password                                            |

### `MetabaseResourceType`

- `database`
- `table`
- `collection`
- `group`

### `MetabaseResourcePermission`

| Type               | Details                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| :----------------- |:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Required. `string` | Metabase permission mapping **Possible values:** - `database`: `schemas:all` \(read table\), `native:write` \(run SQL query\)  **Note**: Metabase requires `schemas:all` permission for `native:write` to be able to work - `collection`: `read`, `write`  **Note**: Metabase table requires `all` permission to read table, no write permission on table level **Note**: Metabase group requires no specific permission to be a member of group ||
