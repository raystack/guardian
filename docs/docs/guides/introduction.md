# Introduction

This tour introduces you to Guardian schema registry. Along the way you will learn how to manage schemas, enforce rules, serialise and deserialise data using Guardian clients.

### Prerequisites

This tour requires you to have Guardian CLI tool installed on your local machine. You can run `guardian version` to verify the installation. Please follow [installation](../getting_started/installation) and [configuration](../getting_started/configuration) guides if you do not have it installed already.

Guardian CLI and clients talks to Guardian server to publish and fetch policies, appeals and resources. Please make sure you also have a Guardian server running. You can also run server locally with `Guardian server start` command. For more details check deployment guide.

### Help

At any time you can run the following commands.

```
# Check the installed version for Guardian cli tool
$ guardian version

# See the help for a command
$ guardian --help
```

The list of all availiable commands are as follows:

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

Help command can also be run on any sub command with syntax `guardian <command> <subcommand> --help` Here is an example for the same.

```
$ guardian policy --help
```

Check the reference for Guardian cli commands.

```
$ guardian reference
```

### Background for this tutorial

We have 1 BigQuery project named `my-bq-project` and we want to manage user access using Guardian. We allow user to have access to datasets and tables with either `Viewer`, `Editor`, or `Owner` roles. We will also be defining certain rules to manage the approval, updating resource metadata, creating an appeal, approving and rejecting the appeal in this example guide.