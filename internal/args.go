package internal

import (
	"strings"

	"github.com/fverse/protoc-graphql/pkg/utils"
)

type Args struct {
	// Sets the code gen target
	Target string
	// If true, keep the casing for type fields. Else fields will be converted to camel case
	KeepCase bool
	// If true, keeps the prefix in type names
	KeepPrefix bool
	// If true, combines the output file to one single file
	CombineOutput bool
	// Sets custom output file names
	OutputFileNames []string
	// Wether to suffix or prefix input names. Suffixing the word 'Input' is the default behavior
	InputNaming string
	// What to prefix or suffix with the input type names.
	// Word 'Input' is default for suffix and letter 'I' is default for prefix
	Affix string
	// If true, generate schema against all the files explicitly listed in the command line
	// and everything they import. Default to false
	All bool
}

func ParseArgs(params string, logger *Logger) *Args {
	args := Args{}
	argSlice := strings.Split(params, ",")

	for _, p := range argSlice {
		var k string
		var v string

		pair := strings.Split(p, "=")
		k = pair[0]
		if len(pair) >= 2 {
			v = pair[1]
		}

		switch k {
		case "target":
			args.Target = v
		case "keep_case":
			args.KeepCase = true
		case "keep_prefix":
			args.KeepPrefix = utils.ParseTrue(v)
		case "combine_output":
			args.CombineOutput = true
		case "output_filenames":
			args.OutputFileNames = append(args.OutputFileNames, v)
		case "input_naming":
			args.InputNaming = v
		case "affix":
			args.Affix = v
		case "all":
			args.All = utils.ParseTrue(v)
		}
	}
	return &args
}
