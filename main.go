package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fverse/protoc-graphql/internal"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	// Handle CLI commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("protoc-gen-graphql %s\n", internal.Version)
			os.Exit(0)
		case "generate", "gen":
			runGenerate()
			return
		case "init":
			runInit()
			return
		case "help", "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	// Default: run as protoc plugin (reads from stdin)
	runAsPlugin()
}

func runAsPlugin() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading proto: %v\n", err)
		os.Exit(1)
	}

	var request pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(data, &request); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing proto: %v\n", err)
		os.Exit(1)
	}

	plugin := internal.New(&request)
	plugin.Execute()
	plugin.SetSupportOptionalField()

	defer plugin.Info("Codegen completed")

	output, err := proto.Marshal(plugin.Response)
	if err != nil {
		plugin.Error(err, "error serializing output")
	}

	_, err = os.Stdout.Write(output)
	if err != nil {
		plugin.Error(err, "error writing output")
	}
}

func printHelp() {
	fmt.Println(`protoc-gen-graphql - Generate GraphQL schemas from Protocol Buffers

Usage:
  protoc-gen-graphql [command] [options]

Commands:
  generate, gen    Generate GraphQL schema from proto files (recommended)
  init             Initialize options.proto in your proto directory
  help             Show this help message

Generate Command:
  protoc-gen-graphql generate [options] <proto_files...>

  Options:
    -o, --out <dir>          Output directory (default: current directory)
    -I, --proto_path <path>  Additional proto import path (can be repeated)
    --target <value>         Set the target (e.g., "admin", "client", "3")
    --keep_case              Keep original field casing
    --keep_prefix            Keep prefix in type names
    --combine_output         Combine all schemas into one file
    --output_filename <name> Custom output filename (use with --combine_output)
    --input_naming <value>   Input naming style: "suffix" or "prefix"
    --affix <value>          Custom affix for input types

Init Command:
  protoc-gen-graphql init [proto_directory]

  Creates options.proto in <proto_directory>/options/options.proto
  Default proto_directory: ./protobuf

  Options:
    --force                  Overwrite existing options.proto

Examples:
  # Generate schema from proto files (auto-includes options.proto)
  protoc-gen-graphql generate -o ./graphql ./protos/*.proto

  # Generate with options
  protoc-gen-graphql generate --target=3 --combine_output -o ./schema ./api.proto

  # Initialize options.proto in default location (./protobuf/options/)
  protoc-gen-graphql init

  # Initialize options.proto in custom proto directory
  protoc-gen-graphql init ./protos

Plugin Mode (for direct protoc usage):
  protoc --plugin=protoc-gen-graphql --graphql_out=. \
    -I./protos -I/path/to/options hello.proto`)
}
