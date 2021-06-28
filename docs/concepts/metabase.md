# Metabase

Metabase is a data visualization tool that lets you connect to external databases and create charts and dashboards based on the data from the databases. Guardian supports access management to the following resources in Metabase:

1. Database
2. Collection

Metabase itself manages its user access on group-based permissions, while Guardian lets each individual user have access directly to the resources.

<!-- TODO: add graph for metabase resources hierarchy/relation -->

# Authentication
Guardian requires **email** and **password** of an administrator user in Metabase.

Example provider config for metabase:
```yaml
...
credentials:
  host: http://localhost:12345
  user: administrator@email.com
  password: password123
...
```
Read more about metabase provider configuration [here](../reference/metabase-provider.md#config).


# Metabase Access Creation

Guardian creates a group that has only one permission type to one resource in Metabase
Example: If a user wants to have **read** access to the **Product** database (id=99), Guardian will create a group called **database_99_read**, grant it with **read** permission only to the **Product** database, and then add the user to that group.

The group naming convention is:
```
<resource_type>_<resource_id>_<permission_type/role>
```

<!-- TODO: add sequence graph how guardian create access in metabase -->
