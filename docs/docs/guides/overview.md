# Overview

The following topics will describe how to setup and use Guardian.

## Guardian CLI

Guardian provides a command line interface which is used to start the Guardian service using `server` command and a lot of other features like creating and managing appeals, policies, providers and resources. It is not necessary to use the command line tool to interact with the Guardian server. GRPC/REST can also be used to interact with the server which is what CLI does internally for communication with the service.


## 1. Policies

Protecting access to IT systems and applications is critical to maintain the integrity of the data. For this purpose any resources 
The Policy defined by the Guardian Admin for each of these resources is based on the principle of the strict need to have been approved either manually by the approvers or automatically based on conditions defined within the policy. This ensures that the admins and the approvers have adequate control to restrict access to system and data.
The first step required to onboard your resources to Guardian is configuring the approval policy.

[Know more](./managing-policies.mdx)

## 2. Providers

Providers are third party services that store our data and these Providers help us draw different conclusions of the same from the analytics within the tool. Google BigQuery for instance is a cloud-based big data analytics web service for processing very large read-only data sets,using SQL-like syntax. Similarly Grafana and Metabase connect with every possible database for pulling up metrics that make sense of the massive amount of data & to monitor our apps with the help of customizable dashboards.A Provider is the source of the resources(viz the data and the analytics) for which the Guardian users create an appeal. Provider instances need to be registered in Guardian so that Guardian can manage access to their resources.

[Know more](./managing-providers.mdx)

## 3. Resources

Guardian manages resources from multiple providers.

[Know More](./managing-resources.mdx)

## 4. Appeals

Appeals are system permissions associated with an account for a particular resource, it represents the permissions either to view or to write the data associated with the resource .Access rights will be disabled or removed once the tenure for which the appeal is accepted expires. The user will be able to request for an extension for the same before its expiry or re-request with an appeal once the permissions are disabled.
An Appeal is the core part of Guardian which is created by the users to get access to a particular resource.

[Know more](./managing-appeals.mdx)
