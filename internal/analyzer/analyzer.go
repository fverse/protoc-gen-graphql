package analyzer

import (
	"github.com/fverse/protoc-graphql/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// TypeAnalyzer performs dependency analysis on protobuf descriptors
// to determine which types are reachable from target-matched RPC methods.
type TypeAnalyzer struct {
	typeRegistry   map[string]*descriptorpb.DescriptorProto
	enumRegistry   map[string]*descriptorpb.EnumDescriptorProto
	reachableTypes map[string]bool
	reachableEnums map[string]bool
	inProgress     map[string]bool
	packageName    string
	packageNames   map[string]bool
}

func NewTypeAnalyzer(protoFiles []*descriptorpb.FileDescriptorProto) *TypeAnalyzer {
	ta := &TypeAnalyzer{
		typeRegistry:   make(map[string]*descriptorpb.DescriptorProto),
		enumRegistry:   make(map[string]*descriptorpb.EnumDescriptorProto),
		reachableTypes: make(map[string]bool),
		reachableEnums: make(map[string]bool),
		inProgress:     make(map[string]bool),
		packageNames:   make(map[string]bool),
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

func (ta *TypeAnalyzer) MarkTypeReachable(typeName string) {
	resolvedName := ta.ResolveTypeName(typeName)

	if ta.reachableTypes[resolvedName] || ta.inProgress[resolvedName] {
		return
	}

	descriptor, exists := ta.typeRegistry[resolvedName]
	if !exists {
		return
	}

	ta.inProgress[resolvedName] = true
	ta.reachableTypes[resolvedName] = true

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

	delete(ta.inProgress, resolvedName)
}

func (ta *TypeAnalyzer) AnalyzeRPCDependencies(services []*descriptorpb.ServiceDescriptorProto, target string) {
	for _, service := range services {
		for _, method := range service.Method {
			methodOptions := getMethodOptions(method)

			if !shouldIncludeMethod(target, methodOptions) {
				continue
			}

			if inputType := method.GetInputType(); inputType != "" {
				ta.MarkTypeReachable(inputType)
			}

			if outputType := method.GetOutputType(); outputType != "" {
				ta.MarkTypeReachable(outputType)
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

func (ta *TypeAnalyzer) IsTypeReachable(typeName string) bool {
	if ta.reachableTypes[typeName] {
		return true
	}

	if len(typeName) > 0 && typeName[0] != '.' {
		if ta.packageName != "" {
			if ta.reachableTypes["."+ta.packageName+"."+typeName] {
				return true
			}
		}

		for pkgName := range ta.packageNames {
			if pkgName != "" && pkgName != ta.packageName {
				if ta.reachableTypes["."+pkgName+"."+typeName] {
					return true
				}
			}
		}

		if ta.reachableTypes["."+typeName] {
			return true
		}

		suffix := "." + typeName
		for reachableType := range ta.reachableTypes {
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
