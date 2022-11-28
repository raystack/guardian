# Shield

Shield is a cloud native role-based authorization aware server and reverse-proxy system. With Shield, you can assign roles to users or groups of users to configure policies that determine whether a particular user has the ability to perform a certain action on a given resource.. Guardian supports access management to the following resources in Shield:

1. Team
2. Project
3. Organization


## Compatible version of shield : 
<= v0.4.1
## Authentication

Guardian requires authentication email of an administrator user having access to all Organizations in Shield.

Example Credential config for shield provider:

```yaml
---
credentials:
  host: http://localhost:12345
  auth_email: "guardian_test@test.com"
```

Example provider config for shield provider:
## Config
```



#### Example

```yaml
type: shield
urn: shield-provider-urn
credentials:
  host: http://localhost:12345
  auth_email: "guardian_test@test.com"
allowed_account_types:
  - user
resources:
  - type: team
    policy:
      id: policy_id
      version: 1
    roles:
      - id: member
        name: Member
        permissions:
          - users
      - id: admin
        name: Admin
        permissions:
          - admins
  - type: project
    policy:
      id: policy_id
      version: 1
    roles:
      - id: admin
        name: Admin
        permissions:
          - admins
  - type: organization
    policy:
      id: policy_id
      version: 1
    roles:
      - id: admin
        name: Admin
        permissions:
          - admins
```

### `ShieldCredentials`

| Fields       |                                                                                               |
|:-------------|:----------------------------------------------------------------------------------------------|
| `host`       | `string` Required. Shield instance host Example: `http://localhost:12345`                     |
| `auth_email` | `email` Required. Email address of an account that has Organization Administration permission |
### `ShieldResourceType`

- `team`
- `project`
- `organization`

### `ShieldResourcePermission`

| Type               | Details                                                                                                                                                                                                              |
| :----------------- |:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Required. `string` | Shield permission mapping **Possible values:** - <br/>`team`: `users` \(Member of team\), `admins` \(admin of team\)   <br/>`project`:` admins` (Admin of project)<br/> `organization`:`admins` (Admin of Org) <br/> ||
