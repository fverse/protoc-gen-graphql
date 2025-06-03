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

	for _, schema := range plugin.schema {
		combinedSchema.objectTypes = append(combinedSchema.objectTypes, schema.objectTypes...)
		combinedSchema.enums = append(combinedSchema.enums, schema.enums...)
		combinedSchema.inputTypes = append(combinedSchema.inputTypes, schema.inputTypes...)
		combinedSchema.mutations = append(combinedSchema.mutations, schema.mutations...)
		combinedSchema.queries = append(combinedSchema.queries, schema.queries...)
	}
	combinedSchema.generate()

	plugin.Response.File = append(plugin.Response.File, &pluginpb.CodeGeneratorResponse_File{
		Name:    utils.String("schema.graphql"),
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
