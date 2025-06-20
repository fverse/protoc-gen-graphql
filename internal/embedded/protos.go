package embedded

import (
	"os"
	"path/filepath"
)

// OptionsProto contains the embedded options.proto content
const OptionsProto = `syntax = "proto3";

option go_package = "/options";

import "google/protobuf/descriptor.proto";

extend google.protobuf.MethodOptions {
    MethodOptions method = 50000;
}

message GqlInput {
  string param = 50031;
  string type = 50032;
  bool optional = 50033;
  bool primitive = 50034;
  bool array = 50035;
  bool empty = 50036;
}

message MethodOptions {
  string kind = 50001;
  string target = 50002;
  GqlInput gql_input = 50003;
  string gql_output = 50004;
  bool skip = 50005;
}

extend google.protobuf.MessageOptions {
  bool skip = 50011;
}

extend google.protobuf.FieldOptions {
  optional bool required = 50021;
  optional bool keep_case = 50022;
}
`

// ExtractProtos extracts the embedded proto files to a temporary directory
// and returns the path to that directory. The caller is responsible for
// cleaning up the directory when done.
func ExtractProtos() (string, error) {
	tempDir, err := os.MkdirTemp("", "protoc-gen-graphql-protos-*")
	if err != nil {
		return "", err
	}

	// Create the protobuf/options directory structure
	optionsDir := filepath.Join(tempDir, "protobuf", "options")
	if err := os.MkdirAll(optionsDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Write options.proto
	optionsPath := filepath.Join(optionsDir, "options.proto")
	if err := os.WriteFile(optionsPath, []byte(OptionsProto), 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}
