
## License
[LICENSE](LICENSE)

# Tierceron

[![GitHub release](https://img.shields.io/github/release/trimble-oss/tierceron-core.svg?style=flat-square)](https://github.com/trimble-oss/tierceron-core/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/trimble-oss/tierceron-core)](https://goreportcard.com/report/github.com/trimble-oss/tierceron-core)
[![PkgGoDev](https://img.shields.io/badge/go.dev-docs-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/trimble-oss/tierceron-core)

## What is it?
Tierceron-core is an interface definition project in support of the tierceron hive.  It provides a list of supported events emitted by the hive kernel.

## Why❓
Tierceron hive utilizes trcsh to run as a service that
delivers secrets securely to running [go plugin](https://pkg.go.dev/plugin) implementation of microservices.

## Key Features
* Configurations delivered in memory from vault to trcsh kernel directly to running tierceron plugin services via an 
* Kernel services can be remotely managed via trcsh scripts running in a release pipeline.  These scripts require authentication in order to run


## Getting started
If you are a contributor, please have a look on the [getting started](GETTING_STARTED.MD) file. Here you can check the information required and other things before providing a useful contribution.

## Trusted Committers
- [Joel Rieke](mailto:joel_rieke@trimble.com)
- [David Mkrtychyan](mailto:david_mkrtychyan@trimble.com)
- [Karnveer Gill](mailto:karnveer_gill@trimble.com)

## Contributing
Contributions are always welcome, no matter how large or small. Before contributing, please read the [code of conduct](CODE_OF_CONDUCT.MD).

See [Contributing](CONTRIBUTING.MD).

## Code review
Check the [code review](CODE_REVIEW.MD) information to find out how a **Pull Request** is evaluated for this project and what other coding standards you should consider when you want to contribute.

## Current effort
Titrating.  Tierceron can do a lot of things.  Some features are very easy to set up and use, others not so much.  Contributions welcomed!
