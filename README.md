# protoc-gen-graphql

A protoc plugin that generates GraphQL schema from Protocol Buffer (`.proto`) files. It analyzes your proto definitions and produces type-safe GraphQL schemas with support for queries, mutations, input types, and enums.

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [Releases page](https://github.com/fverse/protoc-graphql/releases).

**Quick install script (macOS/Linux):**

```bash
# Auto-detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"

VERSION=$(curl -s https://api.github.com/repos/fverse/protoc-graphql/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L "https://github.com/fverse/protoc-graphql/releases/download/${VERSION}/protoc-gen-graphql-${OS}-${ARCH}" -o protoc-gen-graphql
chmod +x protoc-gen-graphql
sudo mv protoc-gen-graphql /usr/local/bin/
```

**Manual download:**

```bash
# macOS (Apple Silicon)
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-darwin-arm64 -o protoc-gen-graphql

# macOS (Intel)
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-darwin-amd64 -o protoc-gen-graphql

# Linux (amd64)
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-linux-amd64 -o protoc-gen-graphql

# Linux (arm64)
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-linux-arm64 -o protoc-gen-graphql

# Then make executable and move to PATH
chmod +x protoc-gen-graphql
sudo mv protoc-gen-graphql /usr/local/bin/
```

For Windows, download `protoc-gen-graphql-windows-amd64.exe` and add it to your PATH.

**Verify checksums:**

Each release includes a `checksums.txt` file. Verify your download:

```bash
sha256sum -c checksums.txt
```

### Install with Go

```bash
go install github.com/fverse/protoc-graphql@latest
```

### Build from Source

```bash
git clone https://github.com/fverse/protoc-graphql.git
cd protoc-graphql
go build -o protoc-gen-graphql
```

### Prerequisites

- Protocol Buffers compiler (`protoc`)

## Usage

### Basic Usage

```bash
protoc --plugin=protoc-gen-graphql=./protoc-gen-graphql \
  --graphql_out=.:./out \
  your_file.proto
```

### With Options

```bash
protoc --plugin=protoc-gen-graphql=./protoc-gen-graphql \
  --graphql_out=target=client,combine_output:./out \
  your_file.proto
```

### Plugin Options

| Option           | Description                                                             | Example            |
| ---------------- | ----------------------------------------------------------------------- | ------------------ |
| `target`         | Filter RPCs by target (e.g., `admin`, `client`, `internal`, `all`, `*`) | `target=client`    |
| `keep_case`      | Preserve original field casing (default: converts to camelCase)         | `keep_case`        |
| `keep_prefix`    | Keep prefix in type names                                               | `keep_prefix=true` |
| `combine_output` | Combine all schemas into a single `schema.graphql` file                 | `combine_output`   |
| `all`            | Generate schema for all files including imports                         | `all=true`         |

### Example Command

```bash
protoc --plugin=protoc-gen-graphql=./protoc-gen-graphql \
  --graphql_out=target=client,keep_case,combine_output:./out \
  -I. -I./protobuf \
  hello.proto
```

## Proto File Configuration

### Import Options

Add the options import to your proto file:

```protobuf
import "protobuf/options/options.proto";
```

### Method Options

Configure GraphQL generation for each RPC method:

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (UserResponse) {
    option (method) = {
      kind: "query"           // "query" or "mutation"
      target: "client"        // Target client: "admin", "client", "internal", "*", "all"
      skip: false             // Skip this method in generation
      gql_input: {
        param: "id"           // GraphQL input parameter name
        type: "String"        // Override input type
        optional: true        // Make input optional
      }
      gql_output: "[User]"    // Override output type (supports arrays)
    };
  }

  rpc CreateUser(CreateUserRequest) returns (UserResponse) {
    option (method) = {
      kind: "mutation"
      target: "admin"
    };
  }
}
```

### Field Options

Control individual field behavior:

```protobuf
message User {
  string first_name = 1;                    // Converts to "firstName" in GraphQL
  string API_KEY = 2 [(keep_case) = true];  // Keeps as "API_KEY"
  string email = 3 [(required) = true];     // Marks as non-nullable (!)
}
```

## Generated Output

### Input Proto

```protobuf
message User {
  string name = 1;
  string email = 2 [(required) = true];
  repeated string roles = 3;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
}

service UserService {
  rpc GetUsers(Empty) returns (UsersResponse) {
    option (method) = { kind: "query" target: "*" };
  }

  rpc CreateUser(CreateUserRequest) returns (User) {
    option (method) = { kind: "mutation" target: "admin" };
  }
}
```

### Generated GraphQL Schema

```graphql
# Code generated by protoc-gen-graphql. DO NOT EDIT
# protoc-gen-graphql v0.1

type User {
  name: String
  email: String!
  roles: [String]
}

input ICreateUserRequest {
  name: String
  email: String
}

type Query {
  getUsers: UsersResponse!
}

type Mutation {
  createUser(input: ICreateUserRequest!): User!
}
```

## Type Mapping

| Proto Type                           | GraphQL Type                      |
| ------------------------------------ | --------------------------------- |
| `string`                             | `String`                          |
| `int32`, `int64`, `sint32`, `sint64` | `Int`                             |
| `float`, `double`                    | `Float`                           |
| `bool`                               | `Boolean`                         |
| `bytes`                              | `String`                          |
| `enum`                               | `enum`                            |
| `message`                            | `type` (output) / `input` (input) |
| `repeated T`                         | `[T]`                             |
| `optional T`                         | `T` (nullable)                    |

## Selective Type Generation

The plugin performs reachability analysis to generate only the types that are actually used:

- **Output types**: Only messages reachable from RPC response types are generated as GraphQL `type`
- **Input types**: Only messages reachable from RPC request types are generated as GraphQL `input`
- **Enums**: Only enums referenced by reachable types are included

This keeps your generated schema clean and focused.

## Multi-Target Generation

Generate different schemas for different clients:

```bash
# Generate client-facing schema
protoc --graphql_out=target=client,combine_output:./out/client ...

# Generate admin schema
protoc --graphql_out=target=admin,combine_output:./out/admin ...

# Generate all RPCs
protoc --graphql_out=target=all,combine_output:./out/full ...
```

## License

See [LICENSE](LICENSE) for details.
