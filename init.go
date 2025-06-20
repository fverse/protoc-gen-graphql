package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fverse/protoc-graphql/internal/embedded"
)

func runInit() {
	args := os.Args[2:]

	// Default proto directory
	protoDir := "./protobuf"

	// Parse arguments - first non-flag argument is the proto directory
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			protoDir = arg
			break
		}
	}

	// Create the options directory inside the proto directory
	optionsDir := filepath.Join(protoDir, "options")
	if err := os.MkdirAll(optionsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Write options.proto
	optionsPath := filepath.Join(optionsDir, "options.proto")

	// Check if file already exists
	if _, err := os.Stat(optionsPath); err == nil {
		fmt.Printf("options.proto already exists at %s\n", optionsPath)
		fmt.Println("Use --force to overwrite")

		// Check for --force flag
		force := false
		for _, arg := range args {
			if arg == "--force" {
				force = true
				break
			}
		}
		if !force {
			os.Exit(0)
		}
	}

	if err := os.WriteFile(optionsPath, []byte(embedded.OptionsProto), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing options.proto: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created %s\n", optionsPath)
	fmt.Println()
	fmt.Println("Add this import to your proto files:")
	fmt.Println()
	fmt.Printf("  import \"%s/options/options.proto\";\n", filepath.Base(protoDir))
	fmt.Println()
	fmt.Println("Then you can use options like:")
	fmt.Println(`  - [(required) = true] on fields`)
	fmt.Println(`  - [(keep_case) = true] on fields`)
	fmt.Println(`  - option (method) = { kind: "query" ... } on RPC methods`)
	fmt.Println(`  - option (skip) = true on messages`)
}
