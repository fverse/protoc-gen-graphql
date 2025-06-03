package internal

import (
	"log"
	"os"
	"strings"

	"google.golang.org/protobuf/types/pluginpb"
)

const (
	NAME    = "protoc-gen-graphql"
	VERSION = "v0.1"
)

type Plugin struct {
	Request  *pluginpb.CodeGeneratorRequest
	Response *pluginpb.CodeGeneratorResponse

	args   *Args
	Logger *Logger

	schema []*Schema
}

// Sets the support optional field option
func (plugin *Plugin) SetSupportOptionalField() {
	opt := uint64(pluginpb.CodeGeneratorResponse_Feature_value["FEATURE_PROTO3_OPTIONAL"])
	plugin.Response.SupportedFeatures = &opt
}

// New creates a new Plugin
func New(request *pluginpb.CodeGeneratorRequest) *Plugin {
	logger := NewLogger()
	args := ParseArgs(request.GetParameter(), logger)
	logger.Log("args2: %+v", args)

	return &Plugin{
		Request:  request,
		Response: new(pluginpb.CodeGeneratorResponse),
		args:     args,
		Logger:   logger,
	}
}

func (p *Plugin) Version() string {
	return NAME + " " + VERSION
}

// Prints an error, and exits.
func (p *Plugin) Error(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ": " + err.Error()
	log.Print(NAME+": error: ", s)
	os.Exit(1)
}

// Prints a message
func (p *Plugin) Info(msg ...string) {
	s := strings.Join(msg, " ")
	log.Print(NAME+": ", s)
}
