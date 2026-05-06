package tools

import "context"

type ToolHandler func(ctx context.Context, rawArgs []byte) (any, error)

type Tool struct {
	Name		string
	Description	string
	handler		ToolHandler
	InputSchema	any
}
