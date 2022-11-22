# Architecture

Basic building blocks of Guardian are

- Guardian CLI
- Guardian Service
- Guardian Database
- Guardian Providers
- Scheduler

### Overview

![Architecture Diagram](/assets/guardian-architecture.svg "GuardianArchitecture")

### Guardian CLI

`Guardian` is a command line tool used to interact with the main Guardian service and basic scaffolding job
specifications. It can be used to

- Manage appeals
- Manage policies
- Manage providers
- Manage grants
- Manage resources
- Start Guardian server

### Guardian Service

Guardian cli can start a service that controls and orchestrates all that Guardian has to
offer. Guardian cli uses GRPC to communicate with the Guardian service for almost all the
operations that takes `host` as the flag. Service also exposes few REST endpoints
that can be used with simple curl request for manage appeals, policies, providers, grants and resources.


### Guardian Database

Guardian uses postgres as a storage engine to store data such as appeals, policies, 
providers, grants and resources.

### Guardian Providers

A Provider is the source of the resources(that is the data and the analytics) for which the Guardian users create an appeal. 
Provider instances need to be registered in Guardian so that Guardian can manage access to their resources.

Provider manages roles, resources, provider credentials and also points each resource type to a considered policy. 
Once a provider configuration is registered, Guardian will immediately fetch the resources and store it in the database.

### Scheduler

Guardian have multiple schedulers that are responsible for fetching the resources from the providers and storing it in the database.
Scheduler also manages the lifecycle of the grants. For example, if access is about to expire, scheduler will notify the user before the access expires.

