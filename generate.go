package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fverse/protoc-graphql/internal/embedded"
)

type generateConfig struct {
	protoFiles []string
	outputDir  string
	protoPaths []string
	pluginOpts []string
}

func runGenerate() {
	config := parseGenerateArgs()

	if len(config.protoFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no proto files specified")
		fmt.Fprintln(os.Stderr, "Usage: protoc-gen-graphql generate [options] <proto_files...>")
		os.Exit(1)
	}

	// Check if protoc is available
	if _, err := exec.LookPath("protoc"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: protoc not found in PATH")
		fmt.Fprintln(os.Stderr, "Please install protoc: https://grpc.io/docs/protoc-installation/")
		os.Exit(1)
	}

	// Get the path to this executable (the plugin)
	pluginPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding plugin path: %v\n", err)
		os.Exit(1)
	}

	// Extract embedded protos to temp directory
	tempDir, err := embedded.ExtractProtos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting proto files: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Build protoc command
	args := []string{
		fmt.Sprintf("--plugin=protoc-gen-graphql=%s", pluginPath),
		fmt.Sprintf("--graphql_out=%s", config.outputDir),
	}

	// Add embedded proto path first (for options.proto)
	args = append(args, fmt.Sprintf("-I%s", tempDir))

	// Add user-specified proto paths
	for _, p := range config.protoPaths {
		args = append(args, fmt.Sprintf("-I%s", p))
	}

	// Add current directory as proto path if not already included
	cwd, _ := os.Getwd()
	hasCwd := false
	for _, p := range config.protoPaths {
		absP, _ := filepath.Abs(p)
		if absP == cwd {
			hasCwd = true
			break
		}
	}
	if !hasCwd {
		args = append(args, fmt.Sprintf("-I%s", cwd))
	}

	// Add plugin options if any
	if len(config.pluginOpts) > 0 {
		args[1] = fmt.Sprintf("--graphql_out=%s:%s", strings.Join(config.pluginOpts, ","), config.outputDir)
	}

	// Add proto files
	args = append(args, config.protoFiles...)

	// Run protoc
	cmd := exec.Command("protoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running protoc: %v\n", err)
		os.Exit(1)
	}
}

func parseGenerateArgs() *generateConfig {
	config := &generateConfig{
		outputDir: ".",
	}

	args := os.Args[2:] // Skip "protoc-gen-graphql" and "generate"

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-o" || arg == "--out":
			if i+1 < len(args) {
				i++
				config.outputDir = args[i]
			}
		case strings.HasPrefix(arg, "-o="):
			config.outputDir = strings.TrimPrefix(arg, "-o=")
		case strings.HasPrefix(arg, "--out="):
			config.outputDir = strings.TrimPrefix(arg, "--out=")

		case arg == "-I" || arg == "--proto_path":
			if i+1 < len(args) {
				i++
				config.protoPaths = append(config.protoPaths, args[i])
			}
		case strings.HasPrefix(arg, "-I"):
			config.protoPaths = append(config.protoPaths, strings.TrimPrefix(arg, "-I"))
		case strings.HasPrefix(arg, "--proto_path="):
			config.protoPaths = append(config.protoPaths, strings.TrimPrefix(arg, "--proto_path="))

		case arg == "--target":
			if i+1 < len(args) {
				i++
				config.pluginOpts = append(config.pluginOpts, "target="+args[i])
			}
		case strings.HasPrefix(arg, "--target="):
			config.pluginOpts = append(config.pluginOpts, "target="+strings.TrimPrefix(arg, "--target="))

		case arg == "--keep_case":
			config.pluginOpts = append(config.pluginOpts, "keep_case")

		case arg == "--keep_prefix":
			config.pluginOpts = append(config.pluginOpts, "keep_prefix=true")

		case arg == "--combine_output":
			config.pluginOpts = append(config.pluginOpts, "combine_output")

		case arg == "--output_filename":
			if i+1 < len(args) {
				i++
				config.pluginOpts = append(config.pluginOpts, "output_filenames="+args[i])
			}
		case strings.HasPrefix(arg, "--output_filename="):
			config.pluginOpts = append(config.pluginOpts, "output_filenames="+strings.TrimPrefix(arg, "--output_filename="))

		case arg == "--input_naming":
			if i+1 < len(args) {
				i++
				config.pluginOpts = append(config.pluginOpts, "input_naming="+args[i])
			}
		case strings.HasPrefix(arg, "--input_naming="):
			config.pluginOpts = append(config.pluginOpts, "input_naming="+strings.TrimPrefix(arg, "--input_naming="))

		case arg == "--affix":
			if i+1 < len(args) {
				i++
				config.pluginOpts = append(config.pluginOpts, "affix="+args[i])
			}
		case strings.HasPrefix(arg, "--affix="):
			config.pluginOpts = append(config.pluginOpts, "affix="+strings.TrimPrefix(arg, "--affix="))

		case arg == "--all":
			config.pluginOpts = append(config.pluginOpts, "all=true")

		case !strings.HasPrefix(arg, "-"):
			// Assume it's a proto file
			config.protoFiles = append(config.protoFiles, arg)
		}
	}

	return config
}
