# Guardian

![test workflow](https://github.com/odpf/guardian/actions/workflows/test.yml/badge.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/odpf/guardian?logo=semantic-release)](Version)

Guardian is a data access management tool. It manages resources from various data providers along with the users’ access. Users required to raise an appeal in order to gain access to a particular resource. The appeal will go through several approvals before it is getting approved and granted the access to the user.

<p align="center"><img src="./docs/assets/overview.svg" /></p>

## Key Features

- **Provider Management**: Support various providers (currently only BigQuery, more coming up!) and multiple instances for each provider type
- **Resource Management**: Resources from a provider are managed in Guardian's database. There is also an API to update resource's metadata to add additional information.
- **Appeal-based access**: Users are expected to create an appeal for accessing data from registered providers. The appeal will get reviewed by the configured approvers before it gives the access to the user.
- **Configurable approval flow**: Approval flow configures what are needed for an appeal to get approved and who are eligible to approve/reject. It can be configured and linked to a provider so that every appeal created to their resources will follow the procedure in order to get approved.
- **External Identity Manager**: This gives the flexibility to use any third-party identity manager. User properties.

## Usage

Explore the following resoruces to get started with Guardian:
- [Guides](docs/guides) provides guidance on usage.
- [Concepts](docs/concepts) describes all important Guardian concepts including system architecture.
- [Reference](docs/reference) contains details about configurations and other aspects of Guardian.
- [Contribute](docs/contribute/contribution.md) contains resources for anyone who wants to contribute to Guardian.

## Running locally

Dependencies:
- Git 
- Go 1.15 or above
- PostgreSQL 13.2 or above

```sh
$ git clone git@github.com:odpf/guardian.git
$ cd guardian
$ go run main.go migrate
$ go run main.go serve
```

## Running tests

```sh
$ make test
```

## Contribute

Development of Guardian happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and
improvements. Read below to learn how you can take part in improving Guardian.

Read our [contributing guide](docs/contribute/contribution.md) to learn about our development process, how to propose
bugfixes and improvements, and how to build and test your changes to Guardian.

To help you get your feet wet and get you familiar with our contribution process, we have a list of
[good first issues](https://github.com/odpf/guardian/labels/good%20first%20issue) that contain bugs which have a relatively
limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/odpf/guardian/graphs/contributors).

## License

Guardian is [Apache 2.0](LICENSE) licensed.
