package descriptor

import (
	"log"
	"strings"

	"github.com/fverse/protoc-graphql/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GraphQLType represents a GraphQL type
type GraphQLType string

const (
	Int     GraphQLType = "Int"
	Float   GraphQLType = "Float"
	Boolean GraphQLType = "Boolean"
	String  GraphQLType = "String"
	Object  GraphQLType = "type"
	Input   GraphQLType = "input"
	Enum    GraphQLType = "enum"
	Unknown GraphQLType = "Unknown"
)

// Represents GraphQL Mutation type
type Mutation struct {
	Name    *string
	Target  uint32
	Input   *options.GqlInput
	Payload *string
	Skip    bool
}

// Represents GraphQL Query type
type Query struct {
	Name    *string
	Target  uint32
	Input   *options.GqlInput
	Payload *string
	Skip    bool
}

type ObjectType struct {
	Fields []*Field
	Name   *string
	Nested []*ObjectType
	Enums  []*Enumeration
}

type Enumeration struct {
	Name   *string
	Values []*string
}

type InputType struct {
	Fields []*Field
	Name   *string
}

// Field represents a field inside a an object type
type Field struct {
	Name         *string
	Type         *GraphQLType
	NonPrimitive bool
	Optional     bool
	IsList       bool
}

type GqlOutput struct {
	Type      string
	Array     bool
	Primitive bool
	Empty     bool
}

// GetType obtains the type of field
func (f *Field) GetType(field *descriptorpb.FieldDescriptorProto) {
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		f.Type = scalar(Int)
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT, descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		f.Type = scalar(Float)
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		f.Type = scalar(Boolean)
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		f.Type = scalar(String)
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if isWellKnownType(field) {
			// TODO: This needs to mapped to a custom Gql scalar type instead of string
			f.Type = scalar(String)
		} else {
			f.Type = (*GraphQLType)(getTypeName(field))
			f.NonPrimitive = true
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		f.Type = (*GraphQLType)(getTypeName(field))
	default:
		f.Type = scalar(Unknown) // TODO: handle this
	}
}

// String returns the actual string value of the GraphQLType type
func (s *GraphQLType) String() string {
	if s == nil {
		return ""
	}
	return string(*s)
}

// Copies the given scalar value and returns a pointer to it.
func scalar(v GraphQLType) *GraphQLType {
	return &v
}

// Check if the field is optional
func (f *Field) IsOptional(field *descriptorpb.FieldDescriptorProto) {
	f.Optional = isOptional(field)
}

// Checks if the field is required
func (f *Field) IsRequired(field *descriptorpb.FieldDescriptorProto) {
	f.Optional = !fieldRequired(field.GetOptions()) && !isRequired(field)
}

// Check if the field is repeated
func (f *Field) IsRepeated(field *descriptorpb.FieldDescriptorProto) {
	f.IsList = isRepeated(field)
}

// Prints a message
func (p *Field) Print(msg ...string) {
	s := strings.Join(msg, " ")
	log.Print(s)
}

func fieldRequired(fieldOptions *descriptorpb.FieldOptions) bool {
	// options := field.GetOptions()
	if proto.HasExtension(fieldOptions, options.E_Required) {
		ext := proto.GetExtension(fieldOptions, options.E_Required)
		return ext.(bool)
	}
	return false
}
