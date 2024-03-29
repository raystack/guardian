# Guardian

![test workflow](https://github.com/raystack/guardian/actions/workflows/test.yaml/badge.svg)
![release workflow](https://github.com/raystack/guardian/actions/workflows/release.yaml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/raystack/guardian/badge.svg?branch=main)](https://coveralls.io/github/raystack/guardian?branch=main)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/raystack/guardian?logo=semantic-release)](Version)

Guardian is a tool for extensible and universal data access with automated access workflows and security controls across data stores, analytical systems, and cloud products.

<p align="center"><img src="./docs/static/assets/overview.svg" /></p>

## Key Features

- **Provider management**: Support various providers (currently only BigQuery, more coming up!) and multiple instances for each provider type
- **Resource management**: Resources from a provider are managed in Guardian's database. There is also an API to update resource's metadata to add additional information.
- **Appeal-based access**: Users are expected to create an appeal for accessing data from registered providers. The appeal will get reviewed by the configured approvers before it gives the access to the user.
- **Configurable approval flow**: Approval flow configures what are needed for an appeal to get approved and who are eligible to approve/reject. It can be configured and linked to a provider so that every appeal created to their resources will follow the procedure in order to get approved.
- **External identity managers**: This gives the flexibility to use any third-party identity manager. User properties.

## Documentation

Explore the following resoruces to get started with Guardian:

- [Guides](https://guardian.vercel.app/docs/tour/introduction) provides guidance on usage.
- [Concepts](https://guardian.vercel.app/docs/concepts/overview) describes all important Guardian concepts including system architecture.
- [Reference](https://guardian.vercel.app/docs/reference/api) contains details about configurations and other aspects of Guardian.
- [Contribute](https://guardian.vercel.app/docs/contribute/contribution) contains resources for anyone who wants to contribute to Guardian.

## Installation

Install Guardian on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. <br/>Refer this for [installations](https://guardian.vercel.app/docs/installation) and [configurations](https://guardian.vercel.app/docs/tour/configuration)

#### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/raystack/guardian/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`guardian` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/raystack/guardian/releases/latest) page:

```sh
brew install raystack/tap/guardian
```

To upgrade to the latest version:

```
brew upgrade guardian
```

#### Linux

`guardian` is available as downloadable binaries from the [releases](https://github.com/raystack/guardian/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`guardian` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/raystack/guardian/releases/latest) page:

```
scoop bucket add guardian https://github.com/raystack/scoop-bucket.git
```

To upgrade to the latest version:

```
scoop update guardian
```

#### Docker

We provide ready to use Docker container images. To pull the latest image:

```
docker pull raystack/guardian:latest
```

To pull a specific version:

```
docker pull raystack/guardian:v0.8.0
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

Guardian provides a fully-featured GRPC and HTTP API to interact with Guardian server. Both APIs adheres to a set of standards that are rigidly followed. Please refer to [proton](https://github.com/raystack/proton/tree/main/raystack/guardian/v1beta1) for GRPC API definitions.

## Contribute

Development of Guardian happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and
improvements. Read our [contributing guide](https://guardian.vercel.app/docs/contribute/contribution) to learn about our development process, how to propose
bugfixes and improvements, and how to build and test your changes to Guardian.

To help you get your feet wet and get you familiar with our contribution process, we have a list of
[good first issues](https://github.com/raystack/guardian/labels/good%20first%20issue) that contain bugs which have a relatively
limited scope. This is a great place to get started.

This project exists thanks to all the [contributors](https://github.com/raystack/guardian/graphs/contributors).

## License

Guardian is [Apache 2.0](LICENSE) licensed.
