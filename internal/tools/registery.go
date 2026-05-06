package tools

import "fmt"

type Registery struct {
	tools map[string]Tool
}

func NewRegistery() *Registery {
	return &Registery{
		tools: make(map[string]Tool),
	}
}

func (r *Registery) register_tool(tool_name string, tool Tool) error {
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
