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

	// Invokes the codegen
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
