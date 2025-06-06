package internal

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/fverse/protoc-graphql/internal/analyzer"
	"github.com/fverse/protoc-graphql/internal/descriptor"
	"github.com/fverse/protoc-graphql/internal/syntax"
	"github.com/fverse/protoc-graphql/options"
	"github.com/fverse/protoc-graphql/pkg/utils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Schema struct {
	*strings.Builder

	// Plugin's parsed command line arguments
	args   *Args
	Logger *Logger

	protoFile   *descriptorpb.FileDescriptorProto
	packageName *string
	fileName    *string

	// Type analyzer for dependency-based filtering
	typeAnalyzer *analyzer.TypeAnalyzer

	objectTypes []*descriptor.ObjectType
	enums       []*descriptor.Enumeration
	inputTypes  []*descriptor.InputType
	mutations   []*descriptor.Mutation
	queries     []*descriptor.Query
}

// Checks the keepCase option for the fields
func keepCase(fieldOptions *descriptorpb.FieldOptions) bool {
	if proto.HasExtension(fieldOptions, options.E_KeepCase) {
		ext := proto.GetExtension(fieldOptions, options.E_KeepCase)
		return ext.(bool)
	}
	return false
}

// Constructs the Object types from message types and fills the schema.objectTypes
func (schema *Schema) makeObjectTypes(messages []*descriptorpb.DescriptorProto) {
	schema.makeObjectTypesWithPrefix(messages, "")
}

// makeObjectTypesWithPrefix constructs object types with a name prefix for nested types
func (schema *Schema) makeObjectTypesWithPrefix(messages []*descriptorpb.DescriptorProto, prefix string) {
	for _, message := range messages {
		// Build the fully qualified name for reachability check
		var fullName string
		if prefix == "" {
			if schema.packageName != nil && *schema.packageName != "" {
				fullName = "." + *schema.packageName + "." + message.GetName()
			} else {
				fullName = "." + message.GetName()
			}
		} else {
			fullName = prefix + "." + message.GetName()
		}

		// Check if this type is reachable before processing
		if !schema.typeAnalyzer.IsTypeReachable(fullName) {
			continue
		}

		if len(message.Field) > 0 {
			objectType := new(descriptor.ObjectType)
			objectType.Name = message.Name

			// Generate type fields
			objectType.Fields = generateFields(message.Field)

			// Construct embedded object types (with updated prefix)
			for _, nested := range message.NestedType {
				schema.makeObjectTypesWithPrefix([]*descriptorpb.DescriptorProto{nested}, fullName)
			}

			// Construct embedded enums (only if reachable)
			for _, enumType := range message.EnumType {
				enumFullName := fullName + "." + enumType.GetName()
				if schema.typeAnalyzer.IsEnumReachable(enumFullName) {
					enum := new(descriptor.Enumeration)
					enum.Name = enumType.Name
					for _, value := range enumType.Value {
						enum.Values = append(enum.Values, enumValues(value))
					}
					schema.enums = append(schema.enums, enum)
				}
			}
			schema.objectTypes = append(schema.objectTypes, objectType)
		}
	}
}

// Return the string value of the provided enum value
func enumValues(value *descriptorpb.EnumValueDescriptorProto) *string {
	return value.Name
}

// Constructs the fields of an object type
func generateFields(fields []*descriptorpb.FieldDescriptorProto) []*descriptor.Field {
	result := make([]*descriptor.Field, 0, len(fields))

	for _, field := range fields {
		f := &descriptor.Field{
			Name: field.Name,
		}
		// Obtain the type of field
		f.GetType(field)

		// Sets wether the field is optional or not
		f.IsRequired(field)

		// Sets wether the field is required or not
		f.IsRepeated(field)

		if !keepCase(field.GetOptions()) {
			f.Name = utils.String(utils.CamelCase(*field.Name))
		}
		result = append(result, f)
	}
	return result
}

func getMethodOptions(method *descriptorpb.MethodDescriptorProto) *options.MethodOptions {
	opts := method.GetOptions()
	if proto.HasExtension(opts, options.E_Method) {
		ext := proto.GetExtension(opts, options.E_Method)
		return ext.(*options.MethodOptions)
	}
	return &options.MethodOptions{}
}

func getGqlOutputType(outputType string, mo *string, packageName *string) *string {
	if outputType != "" {
		outputType = utils.UppercaseFirst(outputType)
		return &outputType
	}
	outputType = strings.TrimPrefix(*mo, "."+*packageName+".")
	return &outputType
}

func isBoolean(t *string) bool {
	return strings.Contains(*t, "Bool")
}

func isEmpty(t *string) bool {
	// query.Input.Type == empty || query.Input.Type == "Empty" || query.Input.Type == "empty"
	return *t == "Empty"
}

func isArray(t *options.GqlInput, length int) bool {
	f := t.Type[:1]
	l := t.Type[length-1:]
	return f == "[" && l == "]"
}

func parseType(input *options.GqlInput) {
	if input.Type == "" {
		return
	}

	if isArray(input, len(input.Type)) {
		input.Array = true
		input.Type = utils.UppercaseFirst(input.Type[1 : len(input.Type)-1])
	} else {
		input.Type = utils.UppercaseFirst(input.Type)
	}

	if isPrimitive(&input.Type) {
		input.Primitive = true
		if input.Type == "Bool" {
			input.Type = "Boolean"
		}
	} else if isEmpty(&input.Type) {
		input.Empty = true
	}
}

func isPrimitive(t *string) bool {
	switch *t {
	case "String", "Boolean", "Bool", "Int", "Float":
		return true
	default:
		return false
	}
}

func getGqlInputParam(input *options.GqlInput) string {
	if param := input.GetParam(); param != "" {
		return param
	}
	return string(syntax.Input)
}

func getGqlInputType(input *options.GqlInput, mi *string, packageName *string) *options.GqlInput {
	if input == nil {
		input = &options.GqlInput{
			Type: "I" + strings.TrimPrefix(*mi, "."+*packageName+"."),
		}
	} else if input.Type != "" {
		parseType(input)
		if !input.Primitive && !input.Empty {
			input.Type = "I" + input.Type
		} else if input.Array {
			input.Type = "[" + input.Type + "]"
		} else {
			input.Type = "I" + strings.TrimPrefix(*mi, "."+*packageName+".")
		}
	} else {
		input.Type = "I" + strings.TrimPrefix(*mi, "."+*packageName+".")
	}

	input.Param = getGqlInputParam(input)
	return input
}

// checkCompilerTarget checks if the CLI target matches the method's target.
func checkCompilerTarget(compilerTarget *string, options *options.MethodOptions) bool {
	// "all" or "*" acts as wildcard
	if *compilerTarget == "all" || *compilerTarget == "*" {
		return true
	}
	return *compilerTarget == options.Target
}

// skipMethod determines if a method should be skipped based on target matching.
func skipMethod(compilerTarget *string, options *options.MethodOptions) bool {
	// Skip if method is explicitly marked to skip
	if options.Skip {
		return true
	}
	// Don't skip if CLI target matches method target
	if checkCompilerTarget(compilerTarget, options) {
		return false
	}
	// Method target "all" or "*" matches any CLI target
	if options.Target == "all" || options.Target == "*" {
		return false
	}
	// Skip: no match
	return true
}

// Constructs the Object types from message types and fills the schema.objectTypes
func (schema *Schema) AddQueriesAndMutations() {
	for _, service := range schema.protoFile.Service {
		for _, method := range service.Method {

			// NewLogger().Log("target: %v", schema.args.Target)
			methodOptions := getMethodOptions(method)

			schema.Logger.Log("methodOptions: %v", methodOptions)

			schema.Logger.Log("target: %s", schema.args.Target)

			if skipMethod(&schema.args.Target, methodOptions) {
				continue
			}

			if methodOptions.Kind == "mutation" || methodOptions.Kind == "Mutation" {
				mutation := new(descriptor.Mutation)
				mutation.Name = method.Name
				mutation.Input = getGqlInputType(methodOptions.GqlInput, method.InputType, schema.packageName)
				mutation.Payload = getGqlOutputType(methodOptions.GqlOutput, method.OutputType, schema.packageName)
				schema.mutations = append(schema.mutations, mutation)
			} else {
				query := new(descriptor.Query)
				query.Name = method.Name
				query.Input = getGqlInputType(methodOptions.GqlInput, method.InputType, schema.packageName)
				query.Payload = getGqlOutputType(methodOptions.GqlOutput, method.OutputType, schema.packageName)
				schema.queries = append(schema.queries, query)
			}
		}
	}
}

// Construct enums (only reachable ones)
func (schema *Schema) Enums() {
	for _, enumType := range schema.protoFile.EnumType {
		// Build fully qualified name for reachability check
		var fullName string
		if schema.packageName != nil && *schema.packageName != "" {
			fullName = "." + *schema.packageName + "." + enumType.GetName()
		} else {
			fullName = "." + enumType.GetName()
		}

		// Check if this enum is reachable before processing
		if !schema.typeAnalyzer.IsEnumReachable(fullName) {
			continue
		}

		enum := new(descriptor.Enumeration)
		enum.Name = enumType.Name
		for _, value := range enumType.Value {
			enum.Values = append(enum.Values, enumValues(value))
		}
		schema.enums = append(schema.enums, enum)
	}
}

// Creates new Schema
func CreateSchema(plugin *Plugin, protoFile *descriptorpb.FileDescriptorProto) *Schema {
	schema := new(Schema)
	schema.Builder = new(strings.Builder)
	schema.protoFile = protoFile
	schema.args = plugin.args
	schema.Logger = plugin.Logger

	// get package name
	schema.packageName = protoFile.Package

	schema.FileName(protoFile.Name)

	// Create type analyzer for dependency-based filtering
	// Pass all proto files for cross-file type resolution
	schema.typeAnalyzer = analyzer.NewTypeAnalyzer(plugin.Request.ProtoFile)

	// Analyze RPC dependencies based on target
	schema.typeAnalyzer.AnalyzeRPCDependencies(protoFile.Service, schema.args.Target)

	// Construct Object types
	schema.makeObjectTypes(protoFile.MessageType)

	schema.Enums()

	schema.AddQueriesAndMutations()
	return schema
}

// Puts a new line in the generated content
func (schema *Schema) NewLine(length ...int) {
	if len(length) == 0 {
		schema.Write("\n")
		return
	}
	for i := 0; i < length[0]; i++ {
		schema.Write("\n")
	}
}

// Adds a space to the generated content
func (schema *Schema) Space(length ...int) {
	if len(length) == 0 {
		schema.Write(" ")
		return
	}
	for i := 0; i < length[0]; i++ {
		schema.Write(" ")
	}
}

// Puts a graphql comment in the generated content
func (schema *Schema) Comment(s string) {
	schema.Write("#")
	schema.Space()
	schema.Write(s)
}

// Write writes a string to the string builder
func (schema *Schema) Write(s string) {
	if len(s) == 0 {
		return
	}
	schema.WriteString(s)
}

// Creates a file name based on the given proto file name
func (schema *Schema) FileName(filename *string) {
	ext := filepath.Ext(*filename)
	schema.fileName = utils.String(strings.TrimSuffix(*filename, ext) + ".graphql")
}

// Prints a message
func (schema *Schema) Print(msg ...string) {
	s := strings.Join(msg, " ")
	log.Print(s)
}

// Write the header content
func (schema *Schema) WriteHeader() {
	schema.Comment("Auto-generated by protoc-gen-graphql. DO NOT EDIT\n")
	schema.Comment(NAME + " " + VERSION)
	schema.NewLine()
	schema.NewLine()
}
