package agent

import (
	"context"
	"fmt"
	"log"

	"agri-qa-system/internal/agent/executor"
	"agri-qa-system/internal/agent/memory"
	"agri-qa-system/internal/agent/planner"
	"agri-qa-system/internal/agent/tools"
	"agri-qa-system/internal/agent/tools/tcm"
	"agri-qa-system/internal/llm"
	"agri-qa-system/internal/model"
	"agri-qa-system/internal/rag"
	"agri-qa-system/internal/store"
)

// Coordinator orchestrates the full Agent workflow:
// Analyze query → Plan tasks → Execute ReAct loop → Synthesize answer
type Coordinator struct {
	analyzer  *planner.QueryAnalyzer
	planner   *planner.TaskPlanner
	executor  *executor.ReactExecutor
	memory    *memory.ShortTermMemory
	registry  *tools.Registry
}

func NewCoordinator(ragPipe *rag.Pipeline, llmClient *llm.Client, redisStore *store.RedisStore) *Coordinator {
	registry := tools.NewRegistry()
	registry.Register(tcm.NewDiagnoseTool())
	registry.Register(tcm.NewHerbTool())
	registry.Register(tcm.NewPrescriptionTool())
	registry.Register(tcm.NewWellnessTool())
	registry.Register(tcm.NewAcupointTool())

	return &Coordinator{
		analyzer: planner.NewQueryAnalyzer(),
		planner:  planner.NewTaskPlanner(registry),
		executor: executor.NewReactExecutor(registry, ragPipe, llmClient),
		memory:   memory.NewShortTermMemory(redisStore),
		registry: registry,
	}
}

// Analyze checks if the query needs Agent mode
func (c *Coordinator) Analyze(query string) (bool, string) {
	return c.analyzer.Analyze(query)
}

// Execute runs the full Agent workflow and streams events
func (c *Coordinator) Execute(ctx context.Context, sessionID, query string, eventCh chan<- model.AgentEvent) error {
	// Get conversation history (non-fatal on error, just proceed without history)
	history, err := c.memory.GetHistory(ctx, sessionID)
	if err != nil {
		log.Printf("[Coordinator] get history error: %v", err)
		history = nil
	}

	// Save user message
	if err := c.memory.SaveMessage(ctx, sessionID, model.Message{Role: "user", Content: query}); err != nil {
		return fmt.Errorf("save user message: %w", err)
	}

	// Plan tasks
	tasks := c.planner.Plan(ctx, query)

	// Emit plan
	emitAgentEvent(eventCh, model.AgentEvent{
		Type: "task_progress",
		Data: map[string]any{
			"tasks":  tasks,
			"status": "planned",
		},
	})

	// Execute ReAct loop
	if err := c.executor.Run(ctx, tasks, query, history, eventCh); err != nil {
		emitAgentEvent(eventCh, model.AgentEvent{
			Type:    "error",
			Content: err.Error(),
		})
		return err
	}

	return nil
}

func (c *Coordinator) SaveAnswer(ctx context.Context, sessionID, answer string) {
	c.memory.SaveMessage(ctx, sessionID, model.Message{Role: "assistant", Content: answer})
}

func emitAgentEvent(ch chan<- model.AgentEvent, event model.AgentEvent) {
	select {
	case ch <- event:
	default:
	}
}
