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

func TestDieturePackageReachability(t *testing.T) {
	// Simulate the actual dieture package structure
	pkgName := "dieture"

	// Create messages similar to hello.proto
	orderRequest := &descriptorpb.DescriptorProto{
		Name: strPtr("OrderRequest"),
	}
	orderResponse := &descriptorpb.DescriptorProto{
		Name: strPtr("OrderResponse"),
	}
	polygonPayload := &descriptorpb.DescriptorProto{
		Name: strPtr("PolygonPayload"),
	}
	empty := &descriptorpb.DescriptorProto{
		Name: strPtr("Empty"),
	}

	// Create a proto file
	protoFile := &descriptorpb.FileDescriptorProto{
		Name:        strPtr("hello.proto"),
		Package:     &pkgName,
		MessageType: []*descriptorpb.DescriptorProto{orderRequest, orderResponse, polygonPayload, empty},
	}

	// Create analyzer
	ta := NewTypeAnalyzer([]*descriptorpb.FileDescriptorProto{protoFile})

	// Simulate RPC analysis:
	// GetHello(PolygonPayload) returns (Empty) - target: "*"
	// GetOrder(OrderRequest) returns (OrderResponse) - target: "admin"

	// Mark input types
	ta.MarkTypeReachableAsInput(".dieture.PolygonPayload")
	ta.MarkTypeReachableAsInput(".dieture.OrderRequest")

	// Mark output types
	ta.MarkTypeReachableAsOutput(".dieture.Empty")
	ta.MarkTypeReachableAsOutput(".dieture.OrderResponse")

	t.Logf("Input reachable types: %v", ta.inputReachableTypes)
	t.Logf("Output reachable types: %v", ta.outputReachableTypes)

	// Test that OrderRequest is ONLY input reachable
	if !ta.IsInputReachable("OrderRequest") {
		t.Error("OrderRequest should be input reachable")
	}
	if ta.IsOutputReachable("OrderRequest") {
		t.Errorf("OrderRequest should NOT be output reachable, but IsOutputReachable returned true")
	}

	// Test that OrderResponse is ONLY output reachable
	if ta.IsInputReachable("OrderResponse") {
		t.Error("OrderResponse should NOT be input reachable")
	}
	if !ta.IsOutputReachable("OrderResponse") {
		t.Error("OrderResponse should be output reachable")
	}

	// Test with fully qualified names
	if ta.IsOutputReachable(".dieture.OrderRequest") {
		t.Error(".dieture.OrderRequest should NOT be output reachable")
	}
	if ta.IsInputReachable(".dieture.OrderResponse") {
		t.Error(".dieture.OrderResponse should NOT be input reachable")
	}
}
