# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
