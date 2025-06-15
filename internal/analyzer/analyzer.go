package analyzer

import (
	"github.com/fverse/protoc-graphql/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// TypeAnalyzer performs dependency analysis on protobuf descriptors
// to determine which types are reachable from target-matched RPC methods.
// It maintains separate tracking for input and output type contexts.
type TypeAnalyzer struct {
	// Map of fully qualified type name to message descriptor
	typeRegistry map[string]*descriptorpb.DescriptorProto

	enumRegistry map[string]*descriptorpb.EnumDescriptorProto

	inputReachableTypes map[string]bool

	// Set of types reachable via RPC OUTPUT paths
	outputReachableTypes map[string]bool

	reachableEnums map[string]bool

	inProgressInput map[string]bool

	inProgressOutput map[string]bool

	// Package names for cross-file resolution
	packageName  string
	packageNames map[string]bool
}

func NewTypeAnalyzer(protoFiles []*descriptorpb.FileDescriptorProto) *TypeAnalyzer {
	ta := &TypeAnalyzer{
		typeRegistry:         make(map[string]*descriptorpb.DescriptorProto),
		enumRegistry:         make(map[string]*descriptorpb.EnumDescriptorProto),
		inputReachableTypes:  make(map[string]bool),
		outputReachableTypes: make(map[string]bool),
		reachableEnums:       make(map[string]bool),
		inProgressInput:      make(map[string]bool),
		inProgressOutput:     make(map[string]bool),
		packageNames:         make(map[string]bool),
	}

	if len(protoFiles) > 0 {
		ta.packageName = protoFiles[0].GetPackage()
	}

	for _, protoFile := range protoFiles {
		pkgName := protoFile.GetPackage()
		ta.packageNames[pkgName] = true
		ta.RegisterTypesFromFile(protoFile.MessageType, "", pkgName)
		ta.RegisterEnumsFromFile(protoFile.EnumType, "", pkgName)
	}

	return ta
}

func NewTypeAnalyzerSingle(protoFile *descriptorpb.FileDescriptorProto) *TypeAnalyzer {
	return NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})
}

func (ta *TypeAnalyzer) RegisterTypes(messages []*descriptorpb.DescriptorProto, prefix string) {
	ta.RegisterTypesFromFile(messages, prefix, ta.packageName)
}

func (ta *TypeAnalyzer) RegisterTypesFromFile(messages []*descriptorpb.DescriptorProto, prefix string, pkgName string) {
	for _, message := range messages {
		var fullName string
		if prefix == "" {
			if pkgName != "" {
				fullName = "." + pkgName + "." + message.GetName()
			} else {
				fullName = "." + message.GetName()
			}
		} else {
			fullName = prefix + "." + message.GetName()
		}

		ta.typeRegistry[fullName] = message

		if len(message.NestedType) > 0 {
			ta.RegisterTypesFromFile(message.NestedType, fullName, pkgName)
		}

		if len(message.EnumType) > 0 {
			ta.registerNestedEnums(message.EnumType, fullName)
		}
	}
}

func (ta *TypeAnalyzer) registerNestedEnums(enums []*descriptorpb.EnumDescriptorProto, prefix string) {
	for _, enum := range enums {
		fullName := prefix + "." + enum.GetName()
		ta.enumRegistry[fullName] = enum
	}
}

func (ta *TypeAnalyzer) RegisterEnums(enums []*descriptorpb.EnumDescriptorProto, prefix string) {
	ta.RegisterEnumsFromFile(enums, prefix, ta.packageName)
}

func (ta *TypeAnalyzer) RegisterEnumsFromFile(enums []*descriptorpb.EnumDescriptorProto, prefix string, pkgName string) {
	for _, enum := range enums {
		var fullName string
		if prefix == "" {
			if pkgName != "" {
				fullName = "." + pkgName + "." + enum.GetName()
			} else {
				fullName = "." + enum.GetName()
			}
		} else {
			fullName = prefix + "." + enum.GetName()
		}

		ta.enumRegistry[fullName] = enum
	}
}

// MarkTypeReachable marks a type as reachable in both input and output contexts.
// This is kept for backward compatibility. For context-specific marking,
// use MarkTypeReachableAsInput or MarkTypeReachableAsOutput.
func (ta *TypeAnalyzer) MarkTypeReachable(typeName string) {
	resolvedName := ta.ResolveTypeName(typeName)

	// Check if already reachable in both contexts
	if (ta.inputReachableTypes[resolvedName] && ta.outputReachableTypes[resolvedName]) ||
		(ta.inProgressInput[resolvedName] && ta.inProgressOutput[resolvedName]) {
		return
	}

	descriptor, exists := ta.typeRegistry[resolvedName]
	if !exists {
		return
	}

	ta.inProgressInput[resolvedName] = true
	ta.inProgressOutput[resolvedName] = true
	ta.inputReachableTypes[resolvedName] = true
	ta.outputReachableTypes[resolvedName] = true

	for _, nested := range descriptor.NestedType {
		nestedName := resolvedName + "." + nested.GetName()
		ta.MarkTypeReachable(nestedName)
	}

	for _, enum := range descriptor.EnumType {
		enumName := resolvedName + "." + enum.GetName()
		ta.reachableEnums[enumName] = true
	}

	for _, field := range descriptor.Field {
		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
			ta.MarkTypeReachable(field.GetTypeName())
		}

		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
			resolvedEnumName := ta.ResolveEnumName(field.GetTypeName())
			ta.reachableEnums[resolvedEnumName] = true
		}
	}

	delete(ta.inProgressInput, resolvedName)
	delete(ta.inProgressOutput, resolvedName)
}

// MarkTypeReachableAsInput recursively marks a type and its dependencies as input-reachable.
// This is used for RPC input types that need GraphQL input generation.
func (ta *TypeAnalyzer) MarkTypeReachableAsInput(typeName string) {
	resolvedName := ta.ResolveTypeName(typeName)

	// Skip if already reachable or currently being processed in input context
	if ta.inputReachableTypes[resolvedName] || ta.inProgressInput[resolvedName] {
		return
	}

	descriptor, exists := ta.typeRegistry[resolvedName]
	if !exists {
		return
	}

	// Mark as in-progress for cycle detection
	ta.inProgressInput[resolvedName] = true
	// Mark as input-reachable
	ta.inputReachableTypes[resolvedName] = true

	// Process nested types in input context
	for _, nested := range descriptor.NestedType {
		nestedName := resolvedName + "." + nested.GetName()
		ta.MarkTypeReachableAsInput(nestedName)
	}

	// Mark nested enums as reachable
	for _, enum := range descriptor.EnumType {
		enumName := resolvedName + "." + enum.GetName()
		ta.reachableEnums[enumName] = true
	}

	// Traverse field dependencies in input context
	for _, field := range descriptor.Field {
		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
			ta.MarkTypeReachableAsInput(field.GetTypeName())
		}

		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
			resolvedEnumName := ta.ResolveEnumName(field.GetTypeName())
			ta.reachableEnums[resolvedEnumName] = true
		}
	}

	// Clear in-progress flag
	delete(ta.inProgressInput, resolvedName)
}

// MarkTypeReachableAsOutput recursively marks a type and its dependencies as output-reachable.
// This is used for RPC output types that need GraphQL type generation.
func (ta *TypeAnalyzer) MarkTypeReachableAsOutput(typeName string) {
	resolvedName := ta.ResolveTypeName(typeName)

	// Skip if already reachable or currently being processed in output context
	if ta.outputReachableTypes[resolvedName] || ta.inProgressOutput[resolvedName] {
		return
	}

	descriptor, exists := ta.typeRegistry[resolvedName]
	if !exists {
		return
	}

	// Mark as in-progress for cycle detection
	ta.inProgressOutput[resolvedName] = true
	// Mark as output-reachable
	ta.outputReachableTypes[resolvedName] = true

	// Process nested types in output context
	for _, nested := range descriptor.NestedType {
		nestedName := resolvedName + "." + nested.GetName()
		ta.MarkTypeReachableAsOutput(nestedName)
	}

	// Mark nested enums as reachable
	for _, enum := range descriptor.EnumType {
		enumName := resolvedName + "." + enum.GetName()
		ta.reachableEnums[enumName] = true
	}

	// Traverse field dependencies in output context
	for _, field := range descriptor.Field {
		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
			ta.MarkTypeReachableAsOutput(field.GetTypeName())
		}

		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
			resolvedEnumName := ta.ResolveEnumName(field.GetTypeName())
			ta.reachableEnums[resolvedEnumName] = true
		}
	}

	// Clear in-progress flag
	delete(ta.inProgressOutput, resolvedName)
}

func (ta *TypeAnalyzer) AnalyzeRPCDependencies(services []*descriptorpb.ServiceDescriptorProto, target string) {
	for _, service := range services {
		for _, method := range service.Method {
			methodOptions := getMethodOptions(method)

			if !shouldIncludeMethod(target, methodOptions) {
				continue
			}

			// Mark input type in input context
			if inputType := method.GetInputType(); inputType != "" {
				ta.MarkTypeReachableAsInput(inputType)
			}

			// Mark output type in output context
			if outputType := method.GetOutputType(); outputType != "" {
				ta.MarkTypeReachableAsOutput(outputType)
			}
		}
	}
}

func getMethodOptions(method *descriptorpb.MethodDescriptorProto) *options.MethodOptions {
	opts := method.GetOptions()
	if proto.HasExtension(opts, options.E_Method) {
		ext := proto.GetExtension(opts, options.E_Method)
		return ext.(*options.MethodOptions)
	}
	return &options.MethodOptions{}
}

func shouldIncludeMethod(cliTarget string, methodOptions *options.MethodOptions) bool {
	if methodOptions.Skip {
		return false
	}

	// "all" or "*" acts as wildcard (matches everything)
	if cliTarget == "all" || cliTarget == "*" || methodOptions.Target == "all" || methodOptions.Target == "*" {
		return true
	}

	return cliTarget == methodOptions.Target
}

// IsInputReachable checks if a type needs GraphQL input generation.
// It handles both fully qualified names
func (ta *TypeAnalyzer) IsInputReachable(typeName string) bool {
	// Check direct match first
	if ta.inputReachableTypes[typeName] {
		return true
	}

	// Handle short names (not starting with '.')
	if len(typeName) > 0 && typeName[0] != '.' {
		// Try with primary package prefix
		if ta.packageName != "" {
			fullyQualified := "." + ta.packageName + "." + typeName
			if ta.inputReachableTypes[fullyQualified] {
				return true
			}
		}

		// Try with other known package prefixes
		for pkgName := range ta.packageNames {
			if pkgName != "" && pkgName != ta.packageName {
				fullyQualified := "." + pkgName + "." + typeName
				if ta.inputReachableTypes[fullyQualified] {
					return true
				}
			}
		}

		// Try with just a leading dot (no package)
		if ta.inputReachableTypes["."+typeName] {
			return true
		}

		// Try suffix matching for nested types
		suffix := "." + typeName
		for reachableType := range ta.inputReachableTypes {
			if len(reachableType) >= len(suffix) && reachableType[len(reachableType)-len(suffix):] == suffix {
				return true
			}
		}
	}

	return false
}

// IsOutputReachable checks if a type needs GraphQL type generation.
// It handles both fully qualified names
func (ta *TypeAnalyzer) IsOutputReachable(typeName string) bool {
	// Check direct match first
	if ta.outputReachableTypes[typeName] {
		return true
	}

	// Handle short names (not starting with '.')
	if len(typeName) > 0 && typeName[0] != '.' {
		// Try with primary package prefix
		if ta.packageName != "" {
			fullyQualified := "." + ta.packageName + "." + typeName
			if ta.outputReachableTypes[fullyQualified] {
				return true
			}
		}

		// Try with other known package prefixes
		for pkgName := range ta.packageNames {
			if pkgName != "" && pkgName != ta.packageName {
				fullyQualified := "." + pkgName + "." + typeName
				if ta.outputReachableTypes[fullyQualified] {
					return true
				}
			}
		}

		// Try with just a leading dot (no package)
		if ta.outputReachableTypes["."+typeName] {
			return true
		}

		// Try suffix matching for nested types
		suffix := "." + typeName
		for reachableType := range ta.outputReachableTypes {
			if len(reachableType) >= len(suffix) && reachableType[len(reachableType)-len(suffix):] == suffix {
				return true
			}
		}
	}

	return false
}

// IsTypeReachable checks if a type is reachable in either input or output context.
// For context-specific checks, use IsInputReachable or IsOutputReachable.
func (ta *TypeAnalyzer) IsTypeReachable(typeName string) bool {
	// Check both input and output reachable sets
	if ta.inputReachableTypes[typeName] || ta.outputReachableTypes[typeName] {
		return true
	}

	if len(typeName) > 0 && typeName[0] != '.' {
		if ta.packageName != "" {
			fullyQualified := "." + ta.packageName + "." + typeName
			if ta.inputReachableTypes[fullyQualified] || ta.outputReachableTypes[fullyQualified] {
				return true
			}
		}

		for pkgName := range ta.packageNames {
			if pkgName != "" && pkgName != ta.packageName {
				fullyQualified := "." + pkgName + "." + typeName
				if ta.inputReachableTypes[fullyQualified] || ta.outputReachableTypes[fullyQualified] {
					return true
				}
			}
		}

		if ta.inputReachableTypes["."+typeName] || ta.outputReachableTypes["."+typeName] {
			return true
		}

		suffix := "." + typeName
		for reachableType := range ta.inputReachableTypes {
			if len(reachableType) >= len(suffix) && reachableType[len(reachableType)-len(suffix):] == suffix {
				return true
			}
		}
		for reachableType := range ta.outputReachableTypes {
			if len(reachableType) >= len(suffix) && reachableType[len(reachableType)-len(suffix):] == suffix {
				return true
			}
		}
	}

	return false
}

func (ta *TypeAnalyzer) IsEnumReachable(enumName string) bool {
	if ta.reachableEnums[enumName] {
		return true
	}

	if len(enumName) > 0 && enumName[0] != '.' {
		if ta.packageName != "" {
			if ta.reachableEnums["."+ta.packageName+"."+enumName] {
				return true
			}
		}

		for pkgName := range ta.packageNames {
			if pkgName != "" && pkgName != ta.packageName {
				if ta.reachableEnums["."+pkgName+"."+enumName] {
					return true
				}
			}
		}

		if ta.reachableEnums["."+enumName] {
			return true
		}

		suffix := "." + enumName
		for reachableEnum := range ta.reachableEnums {
			if len(reachableEnum) >= len(suffix) && reachableEnum[len(reachableEnum)-len(suffix):] == suffix {
				return true
			}
		}
	}

	return false
}

func (ta *TypeAnalyzer) ResolveTypeName(typeName string) string {
	if len(typeName) > 0 && typeName[0] == '.' {
		if _, exists := ta.typeRegistry[typeName]; exists {
			return typeName
		}
	}

	if ta.packageName != "" {
		fullyQualified := "." + ta.packageName + "." + typeName
		if _, exists := ta.typeRegistry[fullyQualified]; exists {
			return fullyQualified
		}
	}

	for pkgName := range ta.packageNames {
		if pkgName != "" {
			fullyQualified := "." + pkgName + "." + typeName
			if _, exists := ta.typeRegistry[fullyQualified]; exists {
				return fullyQualified
			}
		}
	}

	if _, exists := ta.typeRegistry["."+typeName]; exists {
		return "." + typeName
	}

	return typeName
}

func (ta *TypeAnalyzer) ResolveEnumName(enumName string) string {
	if len(enumName) > 0 && enumName[0] == '.' {
		if _, exists := ta.enumRegistry[enumName]; exists {
			return enumName
		}
	}

	if ta.packageName != "" {
		fullyQualified := "." + ta.packageName + "." + enumName
		if _, exists := ta.enumRegistry[fullyQualified]; exists {
			return fullyQualified
		}
	}

	for pkgName := range ta.packageNames {
		if pkgName != "" {
			fullyQualified := "." + pkgName + "." + enumName
			if _, exists := ta.enumRegistry[fullyQualified]; exists {
				return fullyQualified
			}
		}
	}

	if _, exists := ta.enumRegistry["."+enumName]; exists {
		return "." + enumName
	}

	return enumName
}
