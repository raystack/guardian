# Overview

The following topics contains an overview of the importatnt concepts related to the Guardian tool.

## Guardian CLI

Guardian provides a command line interface which is used to start the Guardian service using `server` command and a lot of other features like creating and managing appeals, policies, providers and resources. It is not necessary to use the command line tool to interact with the Guardian server. GRPC/REST can also be used to interact with the server which is what CLI does internally for communication with the service.


## 1. Policies

Protecting access to IT systems and applications is critical to maintain the integrity of the data. For this purpose any resources.<br/><br/>
The Policy defined by the Guardian Admin for each of these resources is based on the principle of the strict need to have been approved either manually or automatically based on conditions defined within the policy. This ensures that the admins and the approvers have adequate control to restrict access to system and data.

Policy controls how users or accounts can get access to a resource. Policy is used by an appeal to determine the approval flow, get user's identity/profile for authorization, and decide whether it needs additional appeals. A Policy is attached to a resource type in the provider configurations, thus it should be the first thing to setup before registering a provider on Guardian.

## 2. Providers

Providers are third party services that store our data and these Providers help us draw different conclusions of the same from the analytics within the tool. Google BigQuery for instance is a cloud-based big data analytics web service for processing very large read-only data sets,using SQL-like syntax. Similarly Grafana and Metabase connect with every possible database for pulling up metrics that make sense of the massive amount of data & to monitor our apps with the help of customizable dashboards.

A Provider is the source of the resources(that is the data and the analytics) for which the Guardian users create an appeal. Provider instances need to be registered in Guardian so that Guardian can manage access to their resources.

Provider manages roles, resources, provider credentials and also points each resource type to a considered policy.
Once a provider configuration is registered, Guardian will immediately fetch the resources and store it in the database.

## 3. Resources

Resource in Guardian represents the data stored in the provider which organizes and controls access to this data.For example, in the [BigQuery](https://cloud.google.com/bigquery/docs/introduction) provider, a resource represents a [dataset](https://cloud.google.com/bigquery/docs/datasets-intro) or a [table](https://cloud.google.com/bigquery/docs/tables). One of Guardian's responsibility is to manage the access to resources along with a pre-defined approval flow so that the organisation's data is well secured and be timely available to all the stakeholders when required.

Guardian manages resources from multiple providers(currently supports **BigQuery, Google Cloud IAM, Grafana, Metabase, Tableau**)


## 4. Appeals

An appeal is essentially a request created by users to give them access to resources. In order to grant the access, an appeal has to be approved by approvers which is assigned based on the applied policy. Appeal contains information about the requested account, the creator, the selected resources, the specific role for accessing the resource, and options to determine the behaviour of the access e.g. permanent or temporary access.

Appeals are system permissions associated with an account for a particular resource, it represents the permissions either to view or to write the data associated with the resource. Access rights will be disabled once the tenure for which the appeal is accepted expires or it can also be revoked by the Guardian admin.

The user will be able to request for an extension for the same before its expiry or re-request with an appeal once the permissions are disabled.
An Appeal is the core part of Guardian which is created by the users to get access to a particular resource.

**Creating Appeal<br/> ** Guardian creates access for users to resources with configurable approval flow. Appeal created by user with specifying which resource they want to access and also the role.

#### Appeal Lifecycle

![](/assets/appeal-lifecycle.png)

#### Appeal Status

- **pending** \(initial status\): During this state, the appeal will evaluate approval steps one by one. The result from the approval steps evaluation will determine whether the appeal will be approved or rejected.
- **rejected**: The appeal has at least one failed approval step.
- **active**: The appeal has been approved. As long as the appeal is in this status, the user will have the access to the designated resource.
- **terminated**: An active access can be revoked by any authorized user at any time, or, if the appeal already exceeds the lifetime limit then it will automatically get revoked.
- **canceled**: The appeal canceled by the creator when the status was on pending.

#### Actions

- **Approve**: Called when all the approval steps are passed/approved.
- **Reject**: Called when there is one approval step that is rejected.
- **Revoke**: A manual action that is called by an authorized user intentionally to revoke an active access.
- **Expire**: If the appeal specifies the expiration policy then it will automatically get expired when it is already passed the lifetime limit.
- **Recreate**: Possible for appeals that are currently still active, rejected, or terminated. This action will create a new appeal based on the previous one. For the appeal coming from active status, there is a policy related to access extension.

#### Approving and Rejecting Appeal flow

![](/assets/approval-flow.png)

