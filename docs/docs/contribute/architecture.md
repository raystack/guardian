# Architecture

Basic building blocks of Guardian are

- Guardian CLI
- Guardian Service
- Guardian Database
- Guardian Providers
- Jobs

#### Guardian CLI

Guardian CLI will be able to can start a service that controls all that Guardian has to offer. Guardian CLI uses GRPC to communicate with the guardian service for almost all the operations.

#### Guardian Service

Guardian service exposes few REST endpoints that can be used with simple curl request for registering or updating a provider, creating and granting/revoking appeals or checking the status of a appeal, creating policies etc.

#### Jobs
You can run [jobs](../reference/jobs.md) using `guardian` cli command to perform one time actions. You can also run them periodically using cronjob through [helm chart](../guides/deployment.md#use-the-helm-chart). 
These jobs support in keeping the list of resources up to date, revoking expired grants, notifying users about expiring grants, etc.

#### Guardian Database

Provider once registered needs to be stored somewhere as a source of truth. Guardian uses postgres as a storage engine to store the provider details, all the resource details which fall under the provider, all the policy information and appeals etc.

#### Guardian Provider

Guardian itself doesn't govern how a appeal will be executed. It only provides the building blocks. A provider for any resource type needs to be build and integrated with Guardian in order to support it's access flow. Any provider has 4 components

- Config - This defines the permissions and credential configuration of a resource.
- Client - This defines the client configurations and methods.
- Provider - This deals with the interaction with a provider.
- Resource - This defines all the resources and functions associated with them.

## Providers Supported

### 1. Grafana Provider

### 2. Metabase Provider

### 3. Bigquery Provider

### 4. Tableau Provider
