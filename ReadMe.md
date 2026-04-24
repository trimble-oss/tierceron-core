
## License
[LICENSE](LICENSE)

# Tierceron-core

[![GitHub release](https://img.shields.io/github/release/trimble-oss/tierceron-core.svg?style=flat-square)](https://github.com/trimble-oss/tierceron-core/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/trimble-oss/tierceron-core)](https://goreportcard.com/report/github.com/trimble-oss/tierceron-core)
[![PkgGoDev](https://img.shields.io/badge/go.dev-docs-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/trimble-oss/tierceron-core)

## What is it?
Tierceron-core is the shared Go module for Tierceron runtime, transport, and flow plumbing. It packages the reusable APIs, plugin coordination types, filesystem abstractions, statistics protobufs, and utility layers that higher-level Tierceron services import.

## What is in this repo?
- `api`: REST, gRPC, and SOAP client support, endpoint definitions, TLS configuration, retry handling, and example code.
- `core`: plugin event types, chat and command channels, configuration context, data-flow statistics, plugin synchronization, and core configuration helpers.
- `flow`: the flow processing framework for initializing and running Tierceron data pipelines.
- `statsdk`: generated protobuf and gRPC definitions for statistics exchange.
- `trcshfs`: shell-facing filesystem abstractions and `trcshio` helpers.
- `bitlock`, `util`, and `util/mlock`: shared locking and utility helpers.
- `atrium`, `buildopts`, `prod`, and `shell`: build options and runtime support packages used by dependent Tierceron components.

## Key Features
- Multi-protocol API caller support for REST, gRPC, and SOAP integrations.
- Shared kernel types for plugin lifecycle, chat exchange, command routing, and data-flow statistics.
- Flow orchestration primitives for building reusable processing pipelines.
- Shared protobuf and filesystem packages used across Tierceron services.


## Getting started
To work on the module locally:

- Run `go test ./...` from the repository root.
- See [api/README.md](api/README.md) for the standalone API caller package documentation.
- See [api/TLS_USAGE.md](api/TLS_USAGE.md) for certificate and TLS configuration details.
- Contributors should start with [GETTING_STARTED.MD](GETTING_STARTED.MD).

## Trusted Committers
- [Joel Rieke](mailto:joel_rieke@trimble.com)
- [David Mkrtychyan](mailto:david_mkrtychyan@trimble.com)
- [Karnveer Gill](mailto:karnveer_gill@trimble.com)
- [Meghan Bailey](mailto:meghan_bailey@trimble.com)

## Contributing
Contributions are always welcome, no matter how large or small. Before contributing, please read the [code of conduct](CODE_OF_CONDUCT.MD).

See [Contributing](CONTRIBUTING.MD).

## Code review
Check the [code review](CODE_REVIEW.MD) information to find out how a **Pull Request** is evaluated for this project and what other coding standards you should consider when you want to contribute.

## Security
Please review [SECURITY.md](SECURITY.md) for vulnerability reporting guidance.
