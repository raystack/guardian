# Introduction

Guardian is a data access management tool. It manages resources from various data providers along with the users’ access. Users required to raise an appeal in order to gain access to a particular resource. The appeal will go through several approvals before it is getting approved and granted the access to the user.

![](.gitbook/assets/overview.svg)

## Key Features

* **Provider Management**: Support various providers \(currently only BigQuery, more coming up!\) and multiple instances for each provider type
* **Resource Management**: Resources from a provider are managed in Guardian's database. There is also an API to update resource's metadata to add additional information.
* **Appeal-based access**: Users are expected to create an appeal for accessing data from registered providers. The appeal will get reviewed by the configured approvers before it gives the access to the user.
* **Configurable approval flow**: Approval flow configures what are needed for an appeal to get approved and who are eligible to approve/reject. It can be configured and linked to a provider so that every appeal created to their resources will follow the procedure in order to get approved.
* **External Identity Manager**: This gives the flexibility to use any third-party identity manager. User properties.

## Usage

Explore the following resoruces to get started with Guardian:

* [Guides](https://github.com/odpf/guardian/tree/f94d5782891f7dd5c3c12ca40834cd0d5b524163/guides/README.md) provides guidance on usage.
* [Concepts](https://github.com/odpf/guardian/tree/f94d5782891f7dd5c3c12ca40834cd0d5b524163/concepts/README.md) describes all important Guardian concepts including system architecture.
* [Reference](https://github.com/odpf/guardian/tree/f94d5782891f7dd5c3c12ca40834cd0d5b524163/reference/README.md) contains details about configurations and other aspects of Guardian.
* [Contribute](https://github.com/odpf/guardian/tree/f94d5782891f7dd5c3c12ca40834cd0d5b524163/contribute/contribution.md) contains resources for anyone who wants to contribute to Guardian.

