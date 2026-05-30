package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	"agri-qa-system/internal/model"
	agentSvc "agri-qa-system/internal/service/agent"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AgentHandler struct {
	coordinator *agentSvc.Coordinator
}

func NewAgentHandler(coordinator *agentSvc.Coordinator) *AgentHandler {
	return &AgentHandler{coordinator: coordinator}
}

// Analyze checks if a query needs Agent mode
func (h *AgentHandler) Analyze(c *fiber.Ctx) error {
	var req model.ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "query is required"})
	}

	needsAgent, reason := h.coordinator.Analyze(req.Query)
	return c.JSON(fiber.Map{
		"needs_agent": needsAgent,
		"reason":      reason,
	})
}

// ChatStream handles Agent mode SSE streaming chat
func (h *AgentHandler) ChatStream(c *fiber.Ctx) error {
	var req model.ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "query is required"})
	}
	if len([]rune(req.Query)) > 2000 {
		return c.Status(400).JSON(fiber.Map{"error": "query too long, max 2000 characters"})
	}
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ctx := c.Context()
	eventCh := make(chan model.AgentEvent, 100)

	go func() {
		defer close(eventCh)
		if err := h.coordinator.Execute(ctx, req.SessionID, req.Query, eventCh); err != nil {
			log.Printf("[AgentStream] error: %v", err)
			emitAgentEvent(eventCh, model.AgentEvent{
				Type:    "error",
				Content: "Agent 推理服务暂时不可用，请稍后重试",
			})
		}
	}()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// send session meta
		writeAgentSSE(w, model.AgentEvent{
			Type: "meta",
			Data: map[string]string{"session_id": req.SessionID, "mode": "agent"},
		})

		var fullAnswer string
		for event := range eventCh {
			writeAgentSSE(w, event)
			if event.Type == "final_answer" {
				fullAnswer = event.Content
			}
			w.Flush()
		}

		// save final answer
		if fullAnswer != "" {
			h.coordinator.SaveAnswer(ctx, req.SessionID, fullAnswer)
		}
	})

	return nil
}

func writeAgentSSE(w *bufio.Writer, event model.AgentEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
}

func emitAgentEvent(ch chan<- model.AgentEvent, event model.AgentEvent) {
	select {
	case ch <- event:
	default:
	}
}
