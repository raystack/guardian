# CLI

`Guardian` is a command line tool used to interact with the main guardian service. Follow the [installation](../getting_started/installation) and [configuration](../getting_started/configuration) guides to set up the CLI tool for Guardian.

## List of Commands

Guardian CLI supports many commands. To get a list of all the commands, follow these steps.
Enter the following code into the terminal:

```text
$ guardian

# or

$ guardian --help
```

List of all availiable commands are as follows:

```text
CORE COMMANDS
  appeal      Manage appeals
  policy      Manage policies
  provider    Manage providers
  resource    Manage resources

ADDITIONAL COMMANDS
  completion  Generate shell completion scripts
  config      Manage client configurations
  help        Help about any command
  job         Manage jobs
  reference   Show command reference
  server      Server management
  version     Print version information
```

To know the usage of any of the core commands use the following syntax:

```text
$ guardian <command> <subcommand> --help
```

## Managing Policies

Policies are used to define governance rules of the data access.
Policies command allows us to list, create or update policies.

**What is inside?**

Enter the following code into the terminal:

```text
$ guardian policy
```

The output is the following:

```text
Available Commands:
  apply       Apply a policy config
  create      Create a new policy
  edit        Edit a policy
  init        Creates a policy template
  list        List and filter access policies
  plan        Show changes from the new policy
  view        View a policy
```

### Policy Init

This command is used to create a policy template with a given file name. Check [policy reference](../reference/policy.md) for more details on the policy configuration.

Syntax for making the policy initialization file.

```
$ guardian policy init --file=<output-name>
```

flags required

```
-f, --file string   File name for the policy config
```

#### Example Configurations

We can configure a policy file `policy.yaml` as shown below.

```yaml
# policy.yaml
id: my_policy
steps:
  - name: manager_approval
    description: Manager approval for sensitive data
    when: $appeal.resource.details.is_sensitive == true
    strategy: manual
    approvers:
      - $appeal.creator.manager_email
  - name: resource_owner_approval
    description: Approval from resource admin/owner
    strategy: manual
    approvers:
      - $appeal.resource.details.owner
```

Now, we can create a policy using the `create` command as given below.

### Create Policy

The create command is used to register a new policy. For this we have to define our policy file, which would be passed as a flag to the `create` command.

Policy has `version` to ensure each appeal has a reference to an applied policy when it's created. A policy is created with an initial `version` equal to `1`.

Usage `guardian policy create [flags]`

Flags required:

```
-f, --file string   Path to the policy config

$ guardian policy create --file=<file-path>
```

The output is the following:

```text
policy created with id: my_policy
```

### List Policies

To get a list and filter of all the avaliable access policies present in the Guardian database, use the `list` command as explained here.

Enter the following code into the terminal:

```text
$ guardian policy list
```

The output is the following:

```text
  ID             VERSION  DESCRIPTION                             STEPS
  my_policy       1        two step policy for tableau workbooks   manager_approval,resource_owner_approval
```

### Edit Policy

To update an existing policy present in the Guardian' database using a file, use the `edit` command as explained here. Updating a policy actually means creating a new policy with the same id but the version gets incremented by 1. Both the new and previous policies still can be used by providers.

Usage :

```
$ guardian policy edit --file=<file-path>
```

An example update of the `policy.yaml` file is given below:

```text
id: my_policy
steps:
  - name: supervisor_approval
    strategy: manual
    approvers:
    - $appeal.resource.details.supervisor
  - name: head_approval
    strategy: manual
    approvers:
    - $appeal.resource.details.owner
```

Now to update the policy defined here.

```text
$ guardian policies edit --file policy.yaml
```

The output is the following:

```text
policy updated
```

Note that on update of a policy it's version is also updated. We can verify this by listing all the policies.

```text
  ID             VERSION  DESCRIPTION                                       STEPS
  my_policy     2        two step policy for tableau workbooks             supervisor_approval,resource_owner_approval
```

### View Policy

View a policy. Display the ID, name, and other information about a policy.

Usgae:

```
$ guardian policy view <policy-id> --version=<policy-version>
```

Flags

```
-o, --output string    Print output with the selected format (default "yaml")
-v, --version string   Version of the policy
```

### Apply Policy

Apply a policy config.Create or edit a policy from a file.

Usage:

```
$ guardian policy apply --file=<file-path>
```

Flags
-f, --file string Path to the policy config

### Plan Policy

Show changes from the new policy. This will not actually apply the policy config.

Usage:

```
$ guardian policy plan --file=<file-path>
```

flags

```
-f, --file string   Path to the policy config
```

## Managing Providers

Providers command allows us to list, create or update providers.

- **What is inside?**

Enter the following code into the terminal:

```text
$ guardian providers
```

The output is the following:

```text
Available Commands:
  apply       Apply a provider
  create      Register a new provider
  edit        Edit a provider
  init        Creates a provider template
  list        List and filter providers
  plan        Show changes from the new provider
  view        View a provider details
```

### Provider Init

This command is used to creates a provider template. Following this define the provider's config file, which would be passed as a flag to the `create` command. Check [provider reference](../reference/provider.md#providerconfig) for more details on the configuration.

Syntax for making the initialization file.

```
guardian provider init [flags]
```

Flags required :

```
-f, --file string   File name for the policy config
```

#### Example Configurations

We can configure a `provider.yaml` file for tableau provider as shown below.

```yaml
type: tableau
urn: 691acb66-27ef-4b4f-9222-f07052e6ffg0
labels:
  entity: gojek
  landscape: id
credentials:
  host: https://prod-apnortheast-a.online.tableau.com
  username: user@test.com
  password: password@123
  content_url: guardiantestsite
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: metric
    policy:
      id: policy_20
      version: 1
    roles:
      - id: read
        name: Read
        permissions:
          - name: Read:Allow
      - id: write
        name: Write
        permissions:
          - name: Write:Allow
```

To register a new provider use the `create` command as shown below.

### Create Provider

The create command is used to register a new provider on the Guardian database.

Usage ` guardian provider create [flags]`

Flags required and usage:

```text
-f, --file string   Path to the provider config

$ guardian providers create --file provider.yaml
```

Output is of the following form:

```text
provider created with id: 26
```

### List Providers

To get a list of all the providers present in the Guardian' database, use the `list` command as explained here.

Enter the following code into the terminal:

```text
$ guardian providers list
```

The output is the following:

```text
  ID  TYPE     URN
  21  tableau  691acb66-27ef-4b4f-9222-f07052e6ffc2
  22  tableau  691acb66-27ef-4b4f-9222-f07052e6ffc8
  26  tableau  691acb66-27ef-4b4f-9222-f07052e6ffg0
  24  tableau  691acb66-27ef-4b4f-9222-f07052e6ffd0
```

### Edit Provider

To update an existing provider present in the Guardian' database, use the `edit` command as explained here. Update the `provider.yaml` file with required changes.

Usage : `$ guardian provider edit <provider-id> --file <file-path>`

Flags required :

```
-f, --file string   Path to the provider config
```

The output is the following:

```text
provider updated
```

### Plan Provider

This command is to show the changes from the new provider. This command will not actually apply the provider config.

Usage : `$ guardian provider plan [flags]`

Flags required :

```
-f, --file string   Path to the provider config
```

### View Provider

This command is used to view a provider details. Displays the ID, name, and other information about a provider.

Usage : `$ guardian provider view <provider-id> [flags]`

Flags required :

```
-o, --output string   Print output with the selected format (default "yaml")
```

### Apply Provider

Apply a provider. It is used to create or edit a provider from a file.

Usage : `$ guardian provider apply [flags]`

Flags required :

```
-f, --file string   Path to the provider config

$ guardian provider apply --file <file-path>
```

## Managing Resources

Resources command allows us to list and set metadata of resources.

- **What is inside?**

Enter the following code into the terminal:

```text
$ guardian resources
```

The output is the following:

```text
Available Commands:
  list        List resources
  set         Store new metadata for a resource
  view        View a resource details
```

### List Resources

It fetches the list of all the resources in the Guardian's database.

Enter the following code into the terminal:

```text
$ guardian resource list
$ guardian resource list --provider-type=bigquery --type=dataset
$ guardian resource list --details=key1.key2:value --details=key1.key3:value
```

List resources flags

```
-D, --deleted                Show deleted resources
-d, --details stringArray    Filter by details object values. Example: --details=key1.key2:value
-n, --name string            Filter by name
-T, --provider-type string   Filter by provider type
-U, --provider-urn string    Filter by provider urn
-t, --type string            Filter by type
-u, --urn string             Filter by urn
```

The output is the following form:

```text
ID    PROVIDER                              TYPE        URN                                   NAME
3552  tableau                               view        8a48df6d-bb5c-438f-a038-35149011e1b5  Flight Delays
      691acb66-27ef-4b4f-9222-f07052e6ffc2
4704  tableau                               metric      a408051f-c394-4a73-8f33-7bf7ba001d99  my-test-metric-ishan
      691acb66-27ef-4b4f-9222-f07052e6ffc2
3792  tableau                               workbook    7c940f8b-34c7-44af-9998-b95deef54edd  Regional
      691acb66-27ef-4b4f-9222-f07052e6ffc8
3802  tableau                               view        3bd3acd1-0681-458b-9566-0519ba844519  Overview
      691acb66-27ef-4b4f-9222-f07052e6ffc8
3807  tableau                               view        703c58f2-5b7f-46ba-bf96-9f4b473e4da8  Commission Model
      691acb66-27ef-4b4f-9222-f07052e6ffc8
5614  tableau                               view        7342fec1-4092-4bd4-abf4-8e531fe0f8ad  Stocks
      691acb66-27ef-4b4f-9222-f07052e6ffd0
```

### Set Resource

Store new metadata for a resource

```
$ guardian resource set <resource-id> --filePath=<file-path>
```

Flags

```
-f, --file string   updated resource file path
```

### View Resource

View a resource details

```
$ guardian resource view <resource-id> --output=json --metadata=true
```

Flags

```
-m, --metadata   Set if you want to see metadata, default: false
```

## Managing Appeals

Appeals command allows us to list, create, list and reject appeal.

**What is inside?**

Enter the following code into the terminal:

```text
$ guardian appeals
```

The output is the following:

```text
Available Commands:
  approve     Approve an approval step
  cancel      Cancel an appeal
  create      Create a new appeal
  list        List and filter appeals
  reject      Reject an approval step
  revoke      Revoke an active access/appeal
  status      Approval status of an appeal
```

### Create Appeal

It helps us to create a new appeal.

** Here are some examples given below: **

```
$ guardian appeal create
$ guardian appeal create --account=<account-id> --type=<account-type> --resource=<resource-id> --role=<role>
```

Flags

```
-a, --account string    Email of the account to appeal
-d, --duration string   Duration of the access
-R, --resource string   ID of the resource
-r, --role string       Role to be assigned
-t, --type string       Type of the account
```

```text
$ guardian appeals create --resource-id 5624 --role write --user test-user@email.com --options.duration "24h"
```

The output is the following:

```text
appeal created with id: 13
```

### List Appeals

This command helps us to list appeals in the Guardian's database, we can also filter the appeals with some additional queries flags.

To filter the list of appeals by Role, Status or Account use the following flags:

```
-a, --account string       Filter by account
-r, --role string          Filter by role
-s, --status stringArray   Filter by status(es)
```

** Here are some examples below: **

```text
$ guardian appeal list
$ guardian appeal list --status=pending
$ guardian appeal list --role=viewer
```

The output is of the form :

```text
  ID  USER                  RESOURCE ID  ROLE   STATUS
  13  test-user@email.com   5624         write  pending
```

### Approve Appeals

It's used to approve an appeal.
Approve an approval step

```
$ guardian appeal approve <appeal-id> --step=<step-name>
```

flags

```

-s, --step string   Name of approval step
```

### Reject Appeals

It's used to reject an appeal.

Reject an approval step

```
$ guardian appeal reject <appeal-id> --step=<step-name>
```

### Revoke Appeal

Revoke an active access/appeal

```
$ guardian appeal revoke <appeal-id>
$ guardian appeal revoke <appeal-id> --reason=<reason>
```

flags

```
-r, --reason string   Reason of the revocation
```

### Check Appeal Status

Status value of an appeal can either be one of these `pending`, `canceled`, `active`,`rejected`,`terminated`. To check the current Approval status of the appeal use the following command:

```
$ guardian appeal status <appeal-id>
```

### Cancel Appeal

Cancel an appeal. **Appeal creator can cancel their appeal while it's status is still on `pending`**

```
$ guardian appeal cancel <appeal-id>
```
