# Introduction

Guardian is a data access management tool. It manages resources from various data providers along with the users’ access. Users required to raise an appeal in order to gain access to a particular resource. The appeal will go through several approvals before it is getting approved and granted the access to the user.

## Key Features

- **Provider Management**: Support various providers (currently only BigQuery, more coming up!) and multiple instances for each provider type
- **Resource Management**: Resources from a provider are managed in Guardian's database. There is also an API to update resource's metadata to add additional information.
- ** Appeal-based access**: Users are expected to create an appeal for accessing data from registered providers. The appeal will get reviewed by the configured approvers before it gives the access to the user.
- **Configurable approval flow**: Approval flow configures what are needed for an appeal to get approved and who are eligible to approve/reject. It can be configured and linked to a provider so that every appeal created to their resources will follow the procedure in order to get approved.
- **External Identity Manager**: This gives the flexibility to use any third-party identity manager. User properties.

## Usage

Explore the following resoruces to get started with Guardian:
- [Guides](/guides) provides guidance on usage.
- [Concepts](/concepts) describes all important Guardian concepts including system architecture.
- [Reference](/reference) contains details about configurations and other aspects of Guardian.
- [Contribute](/contribute/contribution.md) contains resources for anyone who wants to contribute to Guardian.
