# Introduction

Welcome to introductory guide to Guardian! This guide is the best place to start with Guardian. We cover what Guardian is, what problems it can solve, how does it work, and how you can get started using it. If you are familiar with the basics of Guardian, the guide provides a more detailed reference of available features.


## What is Guardian? 

Guardian is a on-demand access management tool with automated access workflows and security controls across data stores, analytical systems, and cloud products. It allows you to manage secure and compliant self-service data access for multiple resources with multiple stakeholders. Users are required to raise an appeal in order to gain access to a particular resource. The appeal will go through several approvals before it is getting approved and granted the access to the user.

![](/assets/overview.svg)

## The Problem we aim to solve!

Organizational teams in charge of authenticating user identities and managing access to corporate resources requires the IT staff to provision access manually. The longer it takes for a user to gain access to crucial business data, the less productive that user will be. On the flip side, failing to revoke the access rights of employees who have left the organization or transferred to different departments can have serious security consequences. To close this window of exposure and risk, IT staff must de-provision access to corporate data as quickly as possible. Manual provisioning and de-provisioning of access is  labor-intensive and prone to human error or oversights. Especially for large organizations, it is not an efficient or sustainable way to manage user identities and access.

Guardian is designed to ensure that our enterprise has robust security controls in place while streamlining authentication procedures to increase user productivity. In conclusion, Guardian is an essential part of ensuring that employees are both empowered to deliver that value and prevented from damaging the business’s reputation or security. 

## How does it work?

Resource administrators needs to register a data provider on Guardian along with an access policy. The policy defines all the steps that a request needs to pass though before access is granted for a resource. The appeal can either be manually approved or automatically checked according to conditions defined in the policy before passing it to the next step of the policy. The policy can be also have multiple approvers associated with an appeal.

Users are required to raise an appeal in order to gain access to a particular resource. The appeal will go through all the approvals/steps defined for that particular resource before it gets approved and the access is granted to the user.

## Key Features

- **Provider Management**: Support various providers \(currently only BigQuery, Metabase, Grafana and Tableau, with more coming up!\) and multiple instances for each provider type.
- **Resource Management**: Resources from a provider are managed in Guardian's database. There is also an API to update resource's metadata to add additional information.
- **Appeal-based access**: Users are expected to create an appeal for accessing data from registered providers. The appeal will get reviewed by the configured approvers before it gives the access to the user.
- **Configurable approval flow**: Approval flow configures what are needed for an appeal to get approved and who are eligible to approve/reject. It can be configured and linked to a provider so that every appeal created to their resources will follow the procedure in order to get approved.
- **External Identity Manager**: This gives the flexibility to use any third-party identity manager. User properties.

## Usage

Explore the following resoruces to get started with Guardian:

- [Guides](./guides/overview.md) provides guidance on usage.
- [Concepts](./concepts/architecture.md) describes all important Guardian concepts including system architecture.
- [Reference](./reference/glossary.md) contains details about configurations and other aspects of Guardian.
- [Contribute](./contribute/contribution.md) contains resources for anyone who wants to contribute to Guardian.

## Where to go from here

See the [installation](../docs/getting_started/installation) page to install the `Guardian` CLI.

Next, we recommend completing the `Guides`. The tour provides an overview of most of the existing functionality of Guardian and takes approximately 30 minutes to complete.

After completing the tour, check out the remainder of the documentation in the `reference` and `concepts` section for your specific areas of interest. We've aimed to provide as much documentation as we can for the various components of Guardian to give you a full understanding of Guardian's surface area.

Finally, follow the project on [GitHub](https://github.com/odpf/guardian), and contact us if you'd like to get involved.
