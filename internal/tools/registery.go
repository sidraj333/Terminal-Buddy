package tools

import (
	"context"
	"fmt"
)

type Registery struct {
	tools map[string]Tool
}

func NewRegistery() *Registery {
	return &Registery{
		tools: make(map[string]Tool),
	}
}

func (r *Registery) validateTool(tool Tool) error {
	//validates that a tool has required properties for a tool struct

	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}

	if tool.Description == "" {
		return fmt.Errorf("tool description is required")
	}

	if tool.Handler == nil {
		return fmt.Errorf("tool handler is required")
	}

	for _, parameter := range tool.InputSchema.Parameters {
		if parameter.Name == "" {
			return fmt.Errorf("tool parameter name is required")
		}
		if parameter.Type == "" {
			return fmt.Errorf("tool parameter type is required")
		}
		if parameter.Description == "" {
			return fmt.Errorf("tool parameter description is required")
		}
	}

	return nil
}

func (r *Registery) RegisterTool(tool_name string, tool Tool) error {
	if tool_name == "" {
		return fmt.Errorf("tool lookup name is required")
	}

	if err := r.validateTool(tool); err != nil {
		return err
	}

	_, ok := r.tools[tool_name]
	if ok {
		return fmt.Errorf("tool already exists")
	}
	r.tools[tool_name] = tool
	return nil
}

func (r *Registery) Get(tool_name string) (Tool, error) {
	tool, ok := r.tools[tool_name]
	if !ok {
		return Tool{}, fmt.Errorf("tool does not exist in registery")
	}

	return tool, nil
}

func (r *Registery) Call(tool_name string, ctx context.Context, rawArgs []byte, authClient HTTPClientProvider) (any, error) {
	tool, err := r.Get(tool_name)
	if err != nil {
		return nil, err
	}

	tool_response, err := tool.Handler(ctx, rawArgs, authClient)
	if err != nil {
		return nil, err
	}

	return tool_response, nil

}

func (r *Registery) ListTools() []Tool {
	tools := []Tool {}
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}
