package tools

import (
	"context"
	"testing"
)

var fakeInputSchema = InputSchema{
	Parameters: []ParameterSchema{
		{
			Name:        "document_url",
			Type:        "string",
			Description: "test parameter",
			Required:    true,
		},
	},
}

func TestRegister_HappyPath(t *testing.T) {
	registery := NewRegistery()

	handler := func(ctx context.Context, rawArgs []byte, authClient HTTPClientProvider) (any, error) {
		return "ok", nil
	}

	tool := Tool{
		Name:        "read_google_doc",
		Description: "tool to read a google doc given a url",
		Handler:     handler,
		InputSchema: fakeInputSchema,
	}

	err := registery.register_tool("read_doc", tool)

	if err != nil {
		t.Fatalf("expected no error registering tool, got %v", err)
	}

	retrieved_tool, err := registery.Get("read_doc")
	if err != nil {
		t.Fatalf("expected registered tool to be retrievable, got %v", err)
	}

	if retrieved_tool.Name != "read_google_doc" {
		t.Fatalf("expected tool name %q, got %q", "read_google_doc", retrieved_tool.Name)
	}
}

func TestRegister_DuplicateToolReturnsError(t *testing.T) {
	registery := NewRegistery()

	handler := func(ctx context.Context, rawArgs []byte, authClient HTTPClientProvider) (any, error) {
		return "ok", nil
	}

	tool := Tool{
		Name:        "read_google_doc",
		Description: "tool to read a google doc given a url",
		Handler:     handler,
		InputSchema: fakeInputSchema,
	}

	err := registery.register_tool("read_doc", tool)
	if err != nil {
		t.Fatalf("expected first registration to succeed, got %v", err)
	}

	err = registery.register_tool("read_doc", tool)
	if err == nil {
		t.Fatal("expected duplicate registration to return an error")
	}
}

func TestGetting_Unregistered_Tool(t *testing.T) {
	registery := NewRegistery()

	handler := func(ctx context.Context, rawArgs []byte, authClient HTTPClientProvider) (any, error) {
		return "ok", nil
	}

	tool := Tool{
		Name:        "read_google_doc",
		Description: "tool to read a google doc given a url",
		Handler:     handler,
		InputSchema: fakeInputSchema,
	}

	err := registery.register_tool("read_doc", tool)
	if err != nil {
		t.Fatalf("expected first registration to succeed, got %v", err)
	}

	retrieved_tool, err := registery.Get("unregistered_tool")
	if err == nil {
		t.Fatalf("expected unregistered tool to throw an error but got. %q", retrieved_tool.Name)
	}

}

func TestRegister_EmptyToolNameReturnsError(t *testing.T) {
	registery := NewRegistery()

	handler := func(ctx context.Context, rawArgs []byte, authClient HTTPClientProvider) (any, error) {
		return "ok", nil
	}

	tool := Tool{
		Name:        "",
		Description: "tool with missing name",
		Handler:     handler,
		InputSchema: fakeInputSchema,
	}

	err := registery.register_tool("", tool)
	if err == nil {
		t.Fatal("expected empty tool name to return an error")
	}
}

func TestRegister_NilHandlerReturnsError(t *testing.T) {
	registery := NewRegistery()

	tool := Tool{
		Name:        "read_google_doc",
		Description: "tool with nil handler",
		Handler:     nil,
		InputSchema: fakeInputSchema,
	}

	err := registery.register_tool("read_google_doc", tool)
	if err == nil {
		t.Fatal("expected nil handler to return an error")
	}
}
