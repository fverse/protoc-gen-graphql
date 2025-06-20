# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-06-20

### Added

- New CLI wrapper with `generate` and `init` commands for simplified usage
- `generate` command auto-includes bundled `options.proto`, eliminating manual proto path setup
- `init` command to initialize `options.proto` in your proto directory
- Embedded `options.proto` in the binary for zero-config usage

### Changed

- Changed options.proto package from `dieture` to empty package for cleaner option syntax
- Options now use short form: `(method)`, `(required)`, `(keep_case)`, `(skip)` instead of prefixed versions

### Migration

If upgrading from v0.1.x, your proto files using the old package will continue to work. For new projects, use:

```protobuf
import "protobuf/options/options.proto";

option (method) = { kind: "query" target: "client" };
```

## [0.1.1] - 2025-06-17

### Fixed

- Fixed empty message handling in GraphQL schema generation. Empty messages (e.g., `message Empty {}`) are now correctly handled - queries and mutations using `Empty` as input generate without input parameters instead of referencing a non-existent `IEmpty` type

## [0.1.0] - 2024-XX-XX

### Added

- Initial release
- GraphQL schema generation from Protocol Buffers
- Support for queries and mutations
- Input and output type generation
- Enum support
- Reachability analysis for selective type generation
