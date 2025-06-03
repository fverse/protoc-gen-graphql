package internal

import (
	"fmt"

	"github.com/fverse/protoc-graphql/internal/descriptor"
	"github.com/fverse/protoc-graphql/internal/syntax"
	"github.com/fverse/protoc-graphql/pkg/utils"
)

func (schema *Schema) generateType(object *descriptor.ObjectType) {
	schema.WriteTypeName(syntax.ObjectType, object.Name)

	for _, field := range object.Fields {
		schema.Space(2)
		schema.Write(*field.Name + string(syntax.Colon))
		schema.Space()

		if field.IsList {
			schema.Write(string(syntax.LBracket))
			schema.Write(field.Type.String())

			if !field.Optional {
				schema.Write(string(syntax.Bang))
			}

			schema.Write(string(syntax.RBracket))
		} else {
			schema.Write(field.Type.String())

			if !field.Optional {
				schema.Write(string(syntax.Bang))
			}
		}

		schema.NewLine()
	}
	schema.Write(string(syntax.RBrace))
	schema.NewLine(2)

	// Generate input type
	schema.WriteString(fmt.Sprintf("input I%s {\n", *object.Name))

	for _, field := range object.Fields {
		schema.Space(2)
		schema.Write(*field.Name + string(syntax.Colon))
		schema.Space()

		if field.IsList {
			schema.Write(string(syntax.LBracket))
		}

		if field.NonPrimitive {
			schema.Write("I" + field.Type.String())
		} else {
			schema.Write(field.Type.String())
		}

		if !field.Optional {
			schema.Write(string(syntax.Bang))
		}

		if field.IsList {
			schema.Write(string(syntax.RBracket))
		}

		schema.NewLine()
	}

	schema.Write(string(syntax.RBrace))
	schema.NewLine(2)
}

func (schema *Schema) generateTypes() {
	for _, object := range schema.objectTypes {
		schema.generateType(object)
	}
}

// Generate enums
func (schema *Schema) generateEnums() {
	for _, enum := range schema.enums {
		schema.WriteTypeName(syntax.Enum, enum.Name)

		for _, value := range enum.Values {
			schema.Space(3)
			schema.Write(*value)
			schema.NewLine()
		}
		schema.Write(string(syntax.RBrace))
		schema.NewLine(2)
	}
}

// Generate queries
func (schema *Schema) generateQueries() {
	schema.Write("type Query {\n")

	for _, query := range schema.queries {
		if query.Input.Empty {
			schema.Write(fmt.Sprintf("  %s: %s\n", utils.LowercaseFirst(*query.Name), *query.Payload))
		} else {
			if query.Input.Optional {
				schema.Write(fmt.Sprintf("  %s(%s: %s): %s!\n", utils.LowercaseFirst(*query.Name),
					query.Input.Param, query.Input.Type, *query.Payload))
			} else {
				schema.Write(fmt.Sprintf("  %s(%s: %s!): %s!\n", utils.LowercaseFirst(*query.Name),
					query.Input.Param, query.Input.Type, *query.Payload))
			}
		}
		// q(input: InputType): ObjectType
	}
	schema.Write("}")
	schema.NewLine()
	schema.NewLine()
}

func (schema *Schema) generateMutations() {
	schema.Write("type Mutation {\n")

	for _, mutation := range schema.mutations {
		if mutation.Input.Empty {
			schema.Write(fmt.Sprintf("  %s: %s\n", utils.LowercaseFirst(*mutation.Name), *mutation.Payload))
		} else {
			if mutation.Input.Optional {
				schema.Write(fmt.Sprintf("  %s(%s: %s): %s!\n", utils.LowercaseFirst(*mutation.Name),
					mutation.Input.Param, mutation.Input.Type, *mutation.Payload))
			} else {
				schema.Write(fmt.Sprintf("  %s(%s: %s!): %s!\n", utils.LowercaseFirst(*mutation.Name),
					mutation.Input.Param, mutation.Input.Type, *mutation.Payload))
			}
		}
		// q(input: InputType): ObjectType
	}
	schema.Write("}")
	schema.NewLine()
}

func (schema *Schema) generate() {
	// Write the header content to the string builder
	schema.WriteHeader()

	// Generate types
	schema.generateTypes()

	// Generate enums
	schema.generateEnums()

	// // Generate queries
	schema.generateQueries()

	// Generate mutations
	schema.generateMutations()
}

// Writes the type's name
func (schema *Schema) WriteTypeName(keyWord syntax.Keyword, name *string) {
	schema.Write(string(keyWord))
	schema.Space()
	schema.Write(*name)
	schema.Space()
	schema.Write(string(syntax.LBrace))
	schema.NewLine()
}

func (schema *Schema) WriteMethod(kind, name *string) {
	schema.Write("")
	schema.Space()
	schema.Write(*name)
	schema.Space()
	schema.Write(string(syntax.LBrace))
	schema.NewLine()
}
