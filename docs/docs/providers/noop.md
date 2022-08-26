# No Op

Using a No-op provider, Guardian users can take advantage of policy workflow without adding resources to this provider in Guardian. Users can call the Guardian APIs for approval workflows and appeal management. This can also allow users to locally test Guardian easily without configuring an actual provider.

## Provider Configurations
#### YAML Representation
```yaml
type: noop
urn: tes-noop-URN
credentials: nil
resources:
  - type: noop
    policy:
      id: my_policy
      version: 1
    roles:
      - id: test_role
        name: testRole
```

**`Allowed Account Types`** `user`<br/>
**`Credentials`** `Must be nil`<br/>
**`Resources Type`** `noop`<br/>
**`ResourcePermissions`**: `Should be empty`

