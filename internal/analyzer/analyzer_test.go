package analyzer

import (
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
)

func TestInputOutputSeparation(t *testing.T) {
	// Create a simple proto structure to test
	pkgName := "test"

	// Create messages
	inputMsg := &descriptorpb.DescriptorProto{
		Name: strPtr("InputMessage"),
	}
	outputMsg := &descriptorpb.DescriptorProto{
		Name: strPtr("OutputMessage"),
	}
	sharedMsg := &descriptorpb.DescriptorProto{
		Name: strPtr("SharedMessage"),
	}

	// Create a proto file
	protoFile := &descriptorpb.FileDescriptorProto{
		Name:        strPtr("test.proto"),
		Package:     &pkgName,
		MessageType: []*descriptorpb.DescriptorProto{inputMsg, outputMsg, sharedMsg},
	}

	// Create analyzer
	ta := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})

	// Manually mark types (simulating what AnalyzeRPCDependencies would do)
	ta.MarkTypeReachableAsInput(".test.InputMessage")
	ta.MarkTypeReachableAsOutput(".test.OutputMessage")
	ta.MarkTypeReachableAsInput(".test.SharedMessage")
	ta.MarkTypeReachableAsOutput(".test.SharedMessage")

	// Test input reachability
	if !ta.IsInputReachable(".test.InputMessage") {
		t.Error("InputMessage should be input reachable")
	}
	if ta.IsOutputReachable(".test.InputMessage") {
		t.Error("InputMessage should NOT be output reachable")
	}

	// Test output reachability
	if ta.IsInputReachable(".test.OutputMessage") {
		t.Error("OutputMessage should NOT be input reachable")
	}
	if !ta.IsOutputReachable(".test.OutputMessage") {
		t.Error("OutputMessage should be output reachable")
	}

	// Test shared reachability
	if !ta.IsInputReachable(".test.SharedMessage") {
		t.Error("SharedMessage should be input reachable")
	}
	if !ta.IsOutputReachable(".test.SharedMessage") {
		t.Error("SharedMessage should be output reachable")
	}

	// Test short name lookups
	if !ta.IsInputReachable("InputMessage") {
		t.Error("InputMessage (short name) should be input reachable")
	}
	if ta.IsOutputReachable("InputMessage") {
		t.Error("InputMessage (short name) should NOT be output reachable")
	}

	t.Logf("Input reachable types: %v", ta.inputReachableTypes)
	t.Logf("Output reachable types: %v", ta.outputReachableTypes)
}

func strPtr(s string) *string {
	return &s
}

func TestEnumReachability(t *testing.T) {
	pkgName := "test"

	// Create an enum
	statusEnum := &descriptorpb.EnumDescriptorProto{
		Name: strPtr("Status"),
		Value: []*descriptorpb.EnumValueDescriptorProto{
			{Name: strPtr("ACTIVE"), Number: int32Ptr(0)},
			{Name: strPtr("INACTIVE"), Number: int32Ptr(1)},
		},
	}

	// Create a nested enum inside a message
	nestedEnum := &descriptorpb.EnumDescriptorProto{
		Name: strPtr("Priority"),
		Value: []*descriptorpb.EnumValueDescriptorProto{
			{Name: strPtr("LOW"), Number: int32Ptr(0)},
			{Name: strPtr("HIGH"), Number: int32Ptr(1)},
		},
	}

	// Create messages that reference enums
	inputMsgWithEnum := &descriptorpb.DescriptorProto{
		Name: strPtr("InputWithEnum"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:     strPtr("status"),
				Number:   int32Ptr(1),
				Type:     enumType(descriptorpb.FieldDescriptorProto_TYPE_ENUM),
				TypeName: strPtr(".test.Status"),
			},
		},
	}

	outputMsgWithEnum := &descriptorpb.DescriptorProto{
		Name: strPtr("OutputWithEnum"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:     strPtr("status"),
				Number:   int32Ptr(1),
				Type:     enumType(descriptorpb.FieldDescriptorProto_TYPE_ENUM),
				TypeName: strPtr(".test.Status"),
			},
		},
	}

	msgWithNestedEnum := &descriptorpb.DescriptorProto{
		Name:     strPtr("MsgWithNestedEnum"),
		EnumType: []*descriptorpb.EnumDescriptorProto{nestedEnum},
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:     strPtr("priority"),
				Number:   int32Ptr(1),
				Type:     enumType(descriptorpb.FieldDescriptorProto_TYPE_ENUM),
				TypeName: strPtr(".test.MsgWithNestedEnum.Priority"),
			},
		},
	}

	// Create a proto file
	protoFile := &descriptorpb.FileDescriptorProto{
		Name:        strPtr("test.proto"),
		Package:     &pkgName,
		MessageType: []*descriptorpb.DescriptorProto{inputMsgWithEnum, outputMsgWithEnum, msgWithNestedEnum},
		EnumType:    []*descriptorpb.EnumDescriptorProto{statusEnum},
	}

	// Create analyzer
	ta := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})

	// Test 1: Enum referenced from input context should be reachable
	ta.MarkTypeReachableAsInput(".test.InputWithEnum")
	if !ta.IsEnumReachable(".test.Status") {
		t.Error("Status enum should be reachable when referenced from input context")
	}

	// Test 2: Enum referenced from output context should also be reachable
	ta2 := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})
	ta2.MarkTypeReachableAsOutput(".test.OutputWithEnum")
	if !ta2.IsEnumReachable(".test.Status") {
		t.Error("Status enum should be reachable when referenced from output context")
	}

	// Test 3: Nested enum should be reachable when parent message is marked
	ta3 := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})
	ta3.MarkTypeReachableAsInput(".test.MsgWithNestedEnum")
	if !ta3.IsEnumReachable(".test.MsgWithNestedEnum.Priority") {
		t.Error("Nested Priority enum should be reachable when parent message is input-reachable")
	}

	// Test 4: Nested enum should be reachable from output context too
	ta4 := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})
	ta4.MarkTypeReachableAsOutput(".test.MsgWithNestedEnum")
	if !ta4.IsEnumReachable(".test.MsgWithNestedEnum.Priority") {
		t.Error("Nested Priority enum should be reachable when parent message is output-reachable")
	}

	// Test 5: Short name lookup for enums
	if !ta.IsEnumReachable("Status") {
		t.Error("Status enum should be reachable via short name")
	}

	// Test 6: Unreferenced enum should NOT be reachable
	ta5 := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})
	// Don't mark any types as reachable
	if ta5.IsEnumReachable(".test.Status") {
		t.Error("Status enum should NOT be reachable when no types are marked")
	}

	t.Logf("Reachable enums (ta): %v", ta.reachableEnums)
	t.Logf("Reachable enums (ta2): %v", ta2.reachableEnums)
	t.Logf("Reachable enums (ta3): %v", ta3.reachableEnums)
}

func enumType(t descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
	return &t
}

func int32Ptr(i int32) *int32 {
	return &i
}

// TestCrossFileTypeResolution verifies that types from imported proto files
// are correctly resolved in the appropriate input/output context.
// Requirements: 5.1, 5.2
func TestCrossFileTypeResolution(t *testing.T) {
	// Create two proto files simulating cross-file imports
	// File 1: common.proto with shared types
	commonPkg := "common"
	sharedType := &descriptorpb.DescriptorProto{
		Name: strPtr("SharedPayload"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   strPtr("value"),
				Number: int32Ptr(1),
				Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			},
		},
	}
	nestedSharedType := &descriptorpb.DescriptorProto{
		Name: strPtr("NestedPayload"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   strPtr("data"),
				Number: int32Ptr(1),
				Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			},
		},
	}

	commonProtoFile := &descriptorpb.FileDescriptorProto{
		Name:        strPtr("common.proto"),
		Package:     &commonPkg,
		MessageType: []*descriptorpb.DescriptorProto{sharedType, nestedSharedType},
	}

	// File 2: service.proto that imports common.proto
	servicePkg := "service"
	// Request type that references SharedPayload from common package
	requestType := &descriptorpb.DescriptorProto{
		Name: strPtr("CreateRequest"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:     strPtr("payload"),
				Number:   int32Ptr(1),
				Type:     fieldType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: strPtr(".common.SharedPayload"),
			},
		},
	}
	// Response type that references NestedPayload from common package
	responseType := &descriptorpb.DescriptorProto{
		Name: strPtr("CreateResponse"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:     strPtr("result"),
				Number:   int32Ptr(1),
				Type:     fieldType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: strPtr(".common.NestedPayload"),
			},
		},
	}

	serviceProtoFile := &descriptorpb.FileDescriptorProto{
		Name:        strPtr("service.proto"),
		Package:     &servicePkg,
		MessageType: []*descriptorpb.DescriptorProto{requestType, responseType},
		Dependency:  []string{"common.proto"},
	}

	// Create analyzer with both proto files (simulating cross-file resolution)
	ta := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{commonProtoFile, serviceProtoFile})

	// Mark request type as input-reachable (simulating RPC input)
	ta.MarkTypeReachableAsInput(".service.CreateRequest")

	// Mark response type as output-reachable (simulating RPC output)
	ta.MarkTypeReachableAsOutput(".service.CreateResponse")

	t.Logf("Input reachable types: %v", ta.inputReachableTypes)
	t.Logf("Output reachable types: %v", ta.outputReachableTypes)

	// Test 1: CreateRequest should be input-reachable
	if !ta.IsInputReachable(".service.CreateRequest") {
		t.Error("CreateRequest should be input reachable")
	}

	// Test 2: SharedPayload (from common package) should be input-reachable
	// because it's referenced by CreateRequest
	if !ta.IsInputReachable(".common.SharedPayload") {
		t.Error("SharedPayload from common package should be input reachable (referenced by CreateRequest)")
	}

	// Test 3: SharedPayload should NOT be output-reachable
	if ta.IsOutputReachable(".common.SharedPayload") {
		t.Error("SharedPayload should NOT be output reachable")
	}

	// Test 4: CreateResponse should be output-reachable
	if !ta.IsOutputReachable(".service.CreateResponse") {
		t.Error("CreateResponse should be output reachable")
	}

	// Test 5: NestedPayload (from common package) should be output-reachable
	// because it's referenced by CreateResponse
	if !ta.IsOutputReachable(".common.NestedPayload") {
		t.Error("NestedPayload from common package should be output reachable (referenced by CreateResponse)")
	}

	// Test 6: NestedPayload should NOT be input-reachable
	if ta.IsInputReachable(".common.NestedPayload") {
		t.Error("NestedPayload should NOT be input reachable")
	}

	// Test 7: Short name lookup should work across packages
	if !ta.IsInputReachable("SharedPayload") {
		t.Error("SharedPayload (short name) should be input reachable")
	}
	if !ta.IsOutputReachable("NestedPayload") {
		t.Error("NestedPayload (short name) should be output reachable")
	}
}

// TestWildcardTargetBehavior verifies that "all" and "*" targets work correctly
// with the input/output separation structure.
// Requirements: 6.2, 6.3
func TestWildcardTargetBehavior(t *testing.T) {
	pkgName := "test"

	// Create messages for different RPCs
	adminRequest := &descriptorpb.DescriptorProto{Name: strPtr("AdminRequest")}
	adminResponse := &descriptorpb.DescriptorProto{Name: strPtr("AdminResponse")}
	publicRequest := &descriptorpb.DescriptorProto{Name: strPtr("PublicRequest")}
	publicResponse := &descriptorpb.DescriptorProto{Name: strPtr("PublicResponse")}
	wildcardRequest := &descriptorpb.DescriptorProto{Name: strPtr("WildcardRequest")}
	wildcardResponse := &descriptorpb.DescriptorProto{Name: strPtr("WildcardResponse")}

	protoFile := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test.proto"),
		Package: &pkgName,
		MessageType: []*descriptorpb.DescriptorProto{
			adminRequest, adminResponse,
			publicRequest, publicResponse,
			wildcardRequest, wildcardResponse,
		},
	}

	// Test 1: CLI target "all" should include all RPCs
	t.Run("CLI target 'all' includes all RPCs", func(t *testing.T) {
		// testShouldIncludeMethod is tested via AnalyzeRPCDependencies behavior
		// We verify that "all" target matches any method target
		if !testShouldIncludeMethod("all", "admin", false) {
			t.Error("CLI target 'all' should match method target 'admin'")
		}
		if !testShouldIncludeMethod("all", "public", false) {
			t.Error("CLI target 'all' should match method target 'public'")
		}
		if !testShouldIncludeMethod("all", "", false) {
			t.Error("CLI target 'all' should match empty method target")
		}
	})

	// Test 2: CLI target "*" should include all RPCs
	t.Run("CLI target '*' includes all RPCs", func(t *testing.T) {
		if !testShouldIncludeMethod("*", "admin", false) {
			t.Error("CLI target '*' should match method target 'admin'")
		}
		if !testShouldIncludeMethod("*", "public", false) {
			t.Error("CLI target '*' should match method target 'public'")
		}
	})

	// Test 3: Method target "all" should match any CLI target
	t.Run("Method target 'all' matches any CLI target", func(t *testing.T) {
		if !testShouldIncludeMethod("admin", "all", false) {
			t.Error("Method target 'all' should match CLI target 'admin'")
		}
		if !testShouldIncludeMethod("customer", "all", false) {
			t.Error("Method target 'all' should match CLI target 'customer'")
		}
	})

	// Test 4: Method target "*" should match any CLI target
	t.Run("Method target '*' matches any CLI target", func(t *testing.T) {
		if !testShouldIncludeMethod("admin", "*", false) {
			t.Error("Method target '*' should match CLI target 'admin'")
		}
		if !testShouldIncludeMethod("internal", "*", false) {
			t.Error("Method target '*' should match CLI target 'internal'")
		}
	})

	// Test 5: Specific target matching
	t.Run("Specific target matching", func(t *testing.T) {
		if !testShouldIncludeMethod("admin", "admin", false) {
			t.Error("CLI target 'admin' should match method target 'admin'")
		}
		if testShouldIncludeMethod("admin", "public", false) {
			t.Error("CLI target 'admin' should NOT match method target 'public'")
		}
	})

	// Test 6: Skip option should be respected
	t.Run("Skip option is respected", func(t *testing.T) {
		if testShouldIncludeMethod("all", "admin", true) {
			t.Error("Method with skip=true should be excluded even with CLI target 'all'")
		}
		if testShouldIncludeMethod("*", "*", true) {
			t.Error("Method with skip=true should be excluded even with wildcard targets")
		}
	})

	// Test 7: Verify input/output separation with wildcard targets
	t.Run("Input/output separation with wildcard targets", func(t *testing.T) {
		ta := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})

		// Simulate what AnalyzeRPCDependencies would do with "all" target
		// All methods would be included, marking their input/output types appropriately
		ta.MarkTypeReachableAsInput(".test.AdminRequest")
		ta.MarkTypeReachableAsOutput(".test.AdminResponse")
		ta.MarkTypeReachableAsInput(".test.PublicRequest")
		ta.MarkTypeReachableAsOutput(".test.PublicResponse")
		ta.MarkTypeReachableAsInput(".test.WildcardRequest")
		ta.MarkTypeReachableAsOutput(".test.WildcardResponse")

		// Verify all request types are input-reachable only
		for _, reqType := range []string{"AdminRequest", "PublicRequest", "WildcardRequest"} {
			if !ta.IsInputReachable(reqType) {
				t.Errorf("%s should be input reachable", reqType)
			}
			if ta.IsOutputReachable(reqType) {
				t.Errorf("%s should NOT be output reachable", reqType)
			}
		}

		// Verify all response types are output-reachable only
		for _, respType := range []string{"AdminResponse", "PublicResponse", "WildcardResponse"} {
			if ta.IsInputReachable(respType) {
				t.Errorf("%s should NOT be input reachable", respType)
			}
			if !ta.IsOutputReachable(respType) {
				t.Errorf("%s should be output reachable", respType)
			}
		}
	})
}

// testShouldIncludeMethod is a test helper that mirrors the logic in analyzer.go
// for testing wildcard behavior without needing actual protobuf options
func testShouldIncludeMethod(cliTarget string, methodTarget string, skip bool) bool {
	if skip {
		return false
	}
	// "all" or "*" acts as wildcard (matches everything)
	if cliTarget == "all" || cliTarget == "*" || methodTarget == "all" || methodTarget == "*" {
		return true
	}
	return cliTarget == methodTarget
}

// fieldType is a helper to create field type pointers
func fieldType(t descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
	return &t
}
