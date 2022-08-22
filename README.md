# Guardian

![test workflow](https://github.com/odpf/guardian/actions/workflows/test.yaml/badge.svg)
![release workflow](https://github.com/odpf/guardian/actions/workflows/release.yaml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/odpf/guardian/badge.svg?branch=main)](https://coveralls.io/github/odpf/guardian?branch=main)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/odpf/guardian?logo=semantic-release)](Version)

Guardian is a on-demand access management tool. It manages resources from various data providers along with the users’ access. Users required to raise an appeal in order to gain access to a particular resource. The appeal will go through several approvals before it is getting approved and granted the access to the user.

<p align="center"><img src="./docs/static/assets/overview.svg" /></p>

## Key Features

- **Provider management**: Support various providers (currently only BigQuery, more coming up!) and multiple instances for each provider type
- **Resource management**: Resources from a provider are managed in Guardian's database. There is also an API to update resource's metadata to add additional information.
- **Appeal-based access**: Users are expected to create an appeal for accessing data from registered providers. The appeal will get reviewed by the configured approvers before it gives the access to the user.
- **Configurable approval flow**: Approval flow configures what are needed for an appeal to get approved and who are eligible to approve/reject. It can be configured and linked to a provider so that every appeal created to their resources will follow the procedure in order to get approved.
- **External identity managers**: This gives the flexibility to use any third-party identity manager. User properties.

## Documentation

Explore the following resoruces to get started with Guardian:

- [Guides](https://odpf.github.io/guardian/docs/guides/introduction) provides guidance on usage.
- [Concepts](https://odpf.github.io/guardian/docs/concepts/overview) describes all important Guardian concepts including system architecture.
- [Reference](https://odpf.github.io/guardian/docs/reference/api) contains details about configurations and other aspects of Guardian.
- [Contribute](https://odpf.github.io/guardian/docs/contribute/contribution) contains resources for anyone who wants to contribute to Guardian.

## Installation

Install Guardian on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. <br/>Refer this for [installations](https://odpf.github.io/guardian/docs/installation) and [configurations](https://odpf.github.io/guardian/docs/guides/configuration)

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/guardian/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`guardian` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/odpf/guardian/releases/latest) page:

```sh
brew install odpf/tap/guardian
```

To upgrade to the latest version:

```
brew upgrade guardian
```

Check for installed guardian version

```sh
guardian version
```

#### Linux

`guardian` is available as downloadable binaries from the [releases](https://github.com/odpf/guardian/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`guardian` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/odpf/guardian/releases/latest) page:

```
scoop bucket add guardian https://github.com/odpf/scoop-bucket.git
```

To upgrade to the latest version:

```
scoop update guardian
```

#### Docker

We provide ready to use Docker container images. To pull the latest image:

```
docker pull odpf/guardian:latest
```

To pull a specific version:

```
docker pull odpf/guardian:v0.3.2
```

## Usage

Guardian is purely API-driven. It is very easy to get started with Guardian. It provides CLI, HTTP and GRPC APIs for simpler developer experience.

#### CLI

Guardian CLI is fully featured and simple to use, even for those who have very limited experience working from the command line. Run `guardian --help` to see list of all available commands and instructions to use.

List of commands

```
guardian --help
```

Print command reference

```sh
guardian reference
```

#### API

Guardian provides a fully-featured GRPC and HTTP API to interact with Guardian server. Both APIs adheres to a set of standards that are rigidly followed. Please refer to [proton](https://github.com/odpf/proton/tree/main/odpf/guardian/v1beta1) for GRPC API definitions.

## Running locally

<details>
  <summary>Dependencies:</summary>

    - Git
    - Go 1.17 or above
    - PostgreSQL 13.2 or above

</details>

Clone the repo

```
git clone git@github.com:odpf/guardian.git
```

Install all the golang dependencies

```
make setup
```

Build guardian binary file

```
make build
```

Init server config. Customise with your local configurations.

```
make config
```

Run database migrations

```
./guardian server migrate -c config.yaml
```

Start guardian server

```
./guardian server start -c config.yaml
```

Initialise client configurations

```
./guardian config init
```

## Running tests

Running all unit tests

```sh
make test
```

Print code coverage

```
make coverage
```

## Contribute

Development of Guardian happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and
improvements. Read below to learn how you can take part in improving Guardian.

Read our [contributing guide](https://odpf.github.io/guardian/docs/contribute/contribution) to learn about our development process, how to propose
bugfixes and improvements, and how to build and test your changes to Guardian.

To help you get your feet wet and get you familiar with our contribution process, we have a list of
[good first issues](https://github.com/odpf/guardian/labels/good%20first%20issue) that contain bugs which have a relatively
limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/odpf/guardian/graphs/contributors).

## License

Guardian is [Apache 2.0](LICENSE) licensed.
