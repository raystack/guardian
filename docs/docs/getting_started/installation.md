# Installation
Guardian installation is quite simple. You can install Guardian on macOS, Windows, Linux, OpenBSD, FreeBSD, and on any machine. There are several approaches to installing Guardian.

1. Using a [pre-compiled binary](#binary-cross-platform)
2. Installing with [package manager](#homebrew-installation)
3. Installing from [source](#building-from-source)

---
### Binary (Cross-platform)
The client and server binaries are downloadable at the releases tab. Currently the installer is not available. Download the appropriate version for your platform from the [releases](https://github.com/odpf/guardian/releases) page. Once downloaded, the binary can be run from anywhere. You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account. Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

---
### Homebrew Installation

```sh
# Install guardian (requires homebrew installed)
$ brew install odpf/taps/guardian

# Upgrade guardian (requires homebrew installed)
$ brew upgrade guardian

# Check for installed guardian version
$ guardian version
```
---
### Building from source

#### Prerequisites
Guardian requires the following dependencies:
- Golang (version 1.17 or above)
- Git

#### Build
Run either of the following commands to compile Guardian from source

```sh
$ git clone git@github.com:odpf/guardian.git  (Using SSH Protocol) Or
$ git clone https://github.com/odpf/guardian.git (Using HTTPS Protocol)
```

```
# Install all the golang dependencies
$ make install

# Check all build commands available
$ make help

# Build Guardian binary file       
$ make build
```
---
### Verifying the installation​

To verify Guardian is properly installed, run `guardian --help` on your system. You should see help output. If you are executing it from the command line, make sure it is on your PATH or you may get an error about Guardian not being found.

```
$ guardian --help
```
