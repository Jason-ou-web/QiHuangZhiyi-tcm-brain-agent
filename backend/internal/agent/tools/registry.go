package tools

import (
	"context"
	"fmt"

	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
)

// Registry holds all registered TCM tools
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) List() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// Execute runs a tool by name and returns the result
func (r *Registry) Execute(ctx context.Context, name string, args map[string]any, ragPipe *rag.Pipeline) (*model.ToolResult, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(ctx, args, ragPipe)
}
