package internal

import (
	"strings"

	"github.com/fverse/protoc-graphql/pkg/utils"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// Checks if the proto files is explicitly passed in the command line
func (plugin *Plugin) isFileExplicit(protoFile *descriptorpb.FileDescriptorProto) bool {
	for _, file := range plugin.Request.FileToGenerate {
		if file == *protoFile.Name {
			return true
		}
	}
	return false
}

// Generates the protoc response
func (plugin *Plugin) Execute() {
	plugin.processProtoFiles()
	plugin.generateOutput()
}

func (plugin *Plugin) processProtoFiles() {
	for _, protoFile := range plugin.Request.ProtoFile {
		if !plugin.isFileExplicit(protoFile) {
			continue
		}
		schema := CreateSchema(plugin, protoFile)
		plugin.schema = append(plugin.schema, schema)
	}
}

func (plugin *Plugin) generateOutput() {
	if plugin.args.CombineOutput {
		plugin.generateCombinedOutput()
		return
	}
	plugin.generateSeparateOutputs()
}

func (plugin *Plugin) generateCombinedOutput() {
	var combinedSchema = new(Schema)
	combinedSchema.Builder = new(strings.Builder)
	combinedSchema.args = plugin.args

	// Track already-generated type names for deduplication
	seenObjectTypes := make(map[string]bool)
	seenEnums := make(map[string]bool)
	seenInputTypes := make(map[string]bool)
	seenMutations := make(map[string]bool)
	seenQueries := make(map[string]bool)

	for _, schema := range plugin.schema {
		// Deduplicate object types
		for _, objType := range schema.objectTypes {
			if objType.Name != nil && !seenObjectTypes[*objType.Name] {
				seenObjectTypes[*objType.Name] = true
				combinedSchema.objectTypes = append(combinedSchema.objectTypes, objType)
			}
		}

		// Deduplicate enums
		for _, enum := range schema.enums {
			if enum.Name != nil && !seenEnums[*enum.Name] {
				seenEnums[*enum.Name] = true
				combinedSchema.enums = append(combinedSchema.enums, enum)
			}
		}

		// Deduplicate input types
		for _, inputType := range schema.inputTypes {
			if inputType.Name != nil && !seenInputTypes[*inputType.Name] {
				seenInputTypes[*inputType.Name] = true
				combinedSchema.inputTypes = append(combinedSchema.inputTypes, inputType)
			}
		}

		// Deduplicate mutations
		for _, mutation := range schema.mutations {
			if mutation.Name != nil && !seenMutations[*mutation.Name] {
				seenMutations[*mutation.Name] = true
				combinedSchema.mutations = append(combinedSchema.mutations, mutation)
			}
		}

		// Deduplicate queries
		for _, query := range schema.queries {
			if query.Name != nil && !seenQueries[*query.Name] {
				seenQueries[*query.Name] = true
				combinedSchema.queries = append(combinedSchema.queries, query)
			}
		}
	}
	combinedSchema.generate()

	// Use custom output filename if provided, otherwise default to "schema.graphql"
	outputFileName := "schema.graphql"
	if len(plugin.args.OutputFileNames) > 0 {
		outputFileName = plugin.args.OutputFileNames[0]
	}

	plugin.Response.File = append(plugin.Response.File, &pluginpb.CodeGeneratorResponse_File{
		Name:    utils.String(outputFileName),
		Content: utils.String(combinedSchema.String()),
	})
}

func (plugin *Plugin) generateSeparateOutputs() {
	for _, schema := range plugin.schema {
		schema.generate()
		plugin.Response.File = append(plugin.Response.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    schema.fileName,
			Content: utils.String(schema.String()),
		})
	}
}
