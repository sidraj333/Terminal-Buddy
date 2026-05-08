package tools

import (
	"context"
	"net/http"
)
type ParameterSchema struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

type InputSchema struct {
	Parameters 	[]ParameterSchema
}
type HTTPClientProvider interface {
	HTTPClient(ctx context.Context) (*http.Client, error)
}

type ToolHandler func(ctx context.Context, rawArgs []byte, authClient HTTPClientProvider) (any, error)

type Tool struct {
	Name		string
	Description	string
	Handler		ToolHandler
	InputSchema	InputSchema
}
