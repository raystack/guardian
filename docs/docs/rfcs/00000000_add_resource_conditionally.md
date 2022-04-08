- Feature Name: Add resources conditionally
- Status: in-progress
- Start Date: 2022-04-08
- Authors: Vikash & Sushmith

# Summary

Resource in Guardian represents the actual resource in the provider e.g. for BigQuery provider, a resource represents a
dataset or a table. Guardian collects resources from the provider automatically as soon as it registered. While in
parallel, Guardian also has a job for continuously syncing resources.

In current design it fetch all the database/table/resources in configure provider, there is no way to filter out on
resources.

# Technical Design

Add a scope of condition in provider configuration to filter out on resources.

#### Using expression

Expression can be used to check whether Guardian should add the resources or not.

```yaml
resources:
  add_if: $resource.urn endsWith 'keyset'
  type: tables
  policy:
    id: keyset_table_policy
```

#### Migration

We can edit or add new providers based on new configuration. For already added resources whom we want to make conditional need to soft-delete it from Guardian db.