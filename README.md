# protoc-gen-graphql

A Protocol Buffers compiler plugin that generates GraphQL schemas from your `.proto` files. Define your API once in protobuf, get type-safe GraphQL schemas with queries, mutations, and input types. Features smart type generation that only creates types you actually use, and multi-target support to generate different schemas for different clients (admin, public, internal).

## Installation

### Download Binary

**macOS (Apple Silicon)**

```bash
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-darwin-arm64 -o protoc-gen-graphql
chmod +x protoc-gen-graphql
```

**macOS (Intel)**

```bash
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-darwin-amd64 -o protoc-gen-graphql
chmod +x protoc-gen-graphql
```

**Linux (x86_64)**

```bash
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-linux-amd64 -o protoc-gen-graphql
chmod +x protoc-gen-graphql
```

**Linux (ARM64)**

```bash
curl -L https://github.com/fverse/protoc-graphql/releases/latest/download/protoc-gen-graphql-linux-arm64 -o protoc-gen-graphql
chmod +x protoc-gen-graphql
```

**Windows**

Download `protoc-gen-graphql-windows-amd64.exe` from [releases](https://github.com/fverse/protoc-graphql/releases/latest) and rename it to `protoc-gen-graphql.exe`.

**Optional:** Move to PATH for easier access:

```bash
# macOS/Linux
sudo mv protoc-gen-graphql /usr/local/bin/

# Then you can use without --plugin flag
protoc --graphql_out=./out your_file.proto
```

### Alternative: Install with Go

If you have Go installed:

```bash
go install github.com/fverse/protoc-graphql@latest
```

### Build from Source

```bash
git clone https://github.com/fverse/protoc-graphql.git
cd protoc-graphql
go build -o protoc-gen-graphql
```

**Prerequisites:** Protocol Buffers compiler (`protoc`) must be installed.

## Usage

### CLI Commands (Recommended)

The plugin includes a built-in CLI that simplifies usage by automatically handling the options.proto import:

```bash
# Generate GraphQL schema (auto-includes options.proto)
protoc-gen-graphql generate -o ./out user.proto

# Generate with options
protoc-gen-graphql generate --target=client --combine_output -o ./out user.proto product.proto

# Initialize options.proto in your project (optional, for manual protoc usage)
protoc-gen-graphql init

# Show help
protoc-gen-graphql help
```

#### Generate Command Options

| Option                     | Description                                        |
| -------------------------- | -------------------------------------------------- |
| `-o, --out <dir>`          | Output directory (default: current directory)      |
| `-I, --proto_path <path>`  | Additional proto import path (can be repeated)     |
| `--target <value>`         | Generate only RPCs for specific target             |
| `--keep_case`              | Preserve original field names                      |
| `--keep_prefix`            | Keep prefix in type names                          |
| `--combine_output`         | Merge all schemas into single file                 |
| `--output_filename <name>` | Custom output filename (use with --combine_output) |
| `--input_naming <value>`   | Input naming style: "suffix" or "prefix"           |
| `--affix <value>`          | Custom affix for input types                       |
| `--all`                    | Include types from imported proto files            |

#### Init Command

```bash
# Initialize in default location (./protobuf/options/)
protoc-gen-graphql init

# Initialize in custom proto directory
protoc-gen-graphql init ./protos
```

### Direct protoc Usage

You can also use the plugin directly with protoc:

```bash
protoc --plugin=protoc-gen-graphql=./protoc-gen-graphql \
  --graphql_out=./out \
  -I. -I./protobuf \
  user.proto
```

Options are added after `--graphql_out=` separated by commas:

```bash
protoc --plugin=protoc-gen-graphql=./protoc-gen-graphql \
  --graphql_out=target=client,combine_output:./out \
  user.proto
```

## Configuring Your Proto Files

### 1. Import Options

```protobuf
import "protobuf/options/options.proto";
```

When using the `generate` command, this import is automatically resolved. For direct protoc usage, run `protoc-gen-graphql init` first or include the proto path.

### 2. Annotate RPCs

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User) {
    option (method) = {
      kind: "query"        // "query" or "mutation"
      target: "client"     // "client", "admin", "internal", or "*"
    };
  }

  rpc CreateUser(CreateUserRequest) returns (User) {
    option (method) = {
      kind: "mutation"
      target: "admin"
    };
  }
}
```

### 3. Mark Required Fields (Optional)

```protobuf
message User {
  string name = 1;
  string email = 2 [(required) = true];  // Non-nullable in GraphQL (!)
}
```

### 4. Preserve Field Casing (Optional)

```protobuf
message Config {
  string api_key = 1;                      // Becomes "apiKey"
  string API_KEY = 2 [(keep_case) = true]; // Stays "API_KEY"
}
```

## Complete Example

**user.proto**

```protobuf
syntax = "proto3";
import "protobuf/options/options.proto";

message User {
  string id = 1;
  string name = 2;
  string email = 3 [(required) = true];
  repeated string roles = 4;
}

message GetUserRequest {
  string id = 1;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User) {
    option (method) = { kind: "query" target: "*" };
  }

  rpc CreateUser(CreateUserRequest) returns (User) {
    option (method) = { kind: "mutation" target: "admin" };
  }
}
```

**Generate:**

```bash
protoc-gen-graphql generate -o ./out user.proto
```

**Generated schema.graphql**

```graphql
type User {
  id: String
  name: String
  email: String!
  roles: [String]
}

input IGetUserRequest {
  id: String
}

input ICreateUserRequest {
  name: String
  email: String
}

type Query {
  getUser(input: IGetUserRequest!): User!
}

type Mutation {
  createUser(input: ICreateUserRequest!): User!
}
```

## Selective Type Generation

The plugin intelligently generates only the types you actually use:

- **Reachability Analysis**: Traces which messages are reachable from your RPC methods
- **Output Types**: Only messages used in RPC responses become GraphQL `type`
- **Input Types**: Only messages used in RPC requests become GraphQL `input`
- **Enums**: Only enums referenced by reachable types are included
- **Clean Schemas**: No unused types cluttering your generated schema

This means if you have 100 message types but only use 10 in your RPCs, only those 10 (plus their dependencies) are generated.

## Type Mapping

| Proto Type                   | GraphQL Type                  |
| ---------------------------- | ----------------------------- |
| string                       | String                        |
| int32, int64, sint32, sint64 | Int                           |
| float, double                | Float                         |
| bool                         | Boolean                       |
| bytes                        | String                        |
| enum                         | enum                          |
| message                      | type (output) / input (input) |
| repeated T                   | [T]                           |
| optional T                   | T (nullable)                  |

## Advanced Usage

### Multi-Target Schemas

Generate different schemas for different clients:

```bash
# Public API
protoc-gen-graphql generate --target=client --combine_output -o ./out/client user.proto

# Admin API
protoc-gen-graphql generate --target=admin --combine_output -o ./out/admin user.proto

# Internal services
protoc-gen-graphql generate --target=internal --combine_output -o ./out/internal user.proto
```

### Custom Input/Output Types

```protobuf
rpc GetUsers(GetUsersRequest) returns (GetUsersResponse) {
  option (method) = {
    kind: "query"
    target: "client"
    gql_input: {
      param: "filter"      // Rename input parameter
      type: "UserFilter"   // Custom input type
      optional: true       // Make it optional
    }
    gql_output: "[User]"   // Return array instead of wrapper
  };
}
```

### Skip RPCs

```protobuf
rpc InternalMethod(Request) returns (Response) {
  option (method) = { skip: true };  // Won't appear in schema
}
```

### All Method Options

```protobuf
option (method) = {
  kind: "query"           // "query" or "mutation"
  target: "client"        // Target audience
  skip: false             // Skip generation
  gql_input: {
    param: "id"           // Parameter name
    type: "ID"            // Override type
    optional: true        // Make optional
  }
  gql_output: "[User]"    // Override output type
};
```

## License

See [LICENSE](LICENSE) for details.
