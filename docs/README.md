# Introduction

Guardian is a data access management tool. It allows you to manage secure and compliant self-service data access for multiple resources with multiple stakeholders.

![](.gitbook/assets/overview%20%281%29.svg)

## How does it work?

Resource administrators needs to register a data provider on Guardian along with a access policy. The policy defines all the steps that a request needs to pass though before access is granted for a resource. A step can be a person approving the user request or an automated check before passing it to the next step of the policy.

Users are required to raise an appeal in order to gain access to a particular resource. The appeal will go through all the approvals/steps defined for that particular resource before it gets approved and the access is granted to the user.

## Key Features

* **Provider Management**: Support various providers \(currently only BigQuery, Metabase, Grafana and Tableau, with more coming up!\) and multiple instances for each provider type.
* **Resource Management**: Resources from a provider are managed in Guardian's database. There is also an API to update resource's metadata to add additional information.
* **Appeal-based access**: Users are expected to create an appeal for accessing data from registered providers. The appeal will get reviewed by the configured approvers before it gives the access to the user.
* **Configurable approval flow**: Approval flow configures what are needed for an appeal to get approved and who are eligible to approve/reject. It can be configured and linked to a provider so that every appeal created to their resources will follow the procedure in order to get approved.
* **External Identity Manager**: This gives the flexibility to use any third-party identity manager. User properties.

## Usage

Explore the following resoruces to get started with Guardian:

* [Guides](https://odpf.gitbook.io/guardian/guides/overview) provides guidance on usage.
* [Concepts](https://odpf.gitbook.io/guardian/concepts/architecture) describes all important Guardian concepts including system architecture.
* [Reference](https://odpf.gitbook.io/guardian/reference/overview) contains details about configurations and other aspects of Guardian.
* [Contribute](https://odpf.gitbook.io/guardian/contribute/contribution) contains resources for anyone who wants to contribute to Guardian.

