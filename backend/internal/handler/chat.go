package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	"agri-qa-system/internal/model"
	"agri-qa-system/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ChatHandler struct {
	svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) Chat(c *fiber.Ctx) error {
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

	ctx := c.Context()
	answer, citations, err := h.svc.Chat(ctx, req.SessionID, req.Query)
	if err != nil {
		log.Printf("[Chat] error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "内部服务错误"})
	}

	return c.JSON(model.ChatResponse{
		SessionID: req.SessionID,
		Answer:    answer,
		Citations: citations,
	})
}

func (h *ChatHandler) ChatStream(c *fiber.Ctx) error {
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
	eventCh, citations, err := h.svc.ChatStream(ctx, req.SessionID, req.Query)
	if err != nil {
		log.Printf("[ChatStream] error: %v", err)
		sendSSEError(c, "知识检索服务暂时不可用，请稍后重试")
		return nil
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// send session meta
		writeSSE(w, model.SSEEvent{Type: "meta", Data: map[string]string{"session_id": req.SessionID}})

		// send citations
		writeSSE(w, model.SSEEvent{Type: "citations", Data: citations})

		// stream tokens
		var fullAnswer string
		for event := range eventCh {
			writeSSE(w, event)
			if event.Type == "token" {
				fullAnswer += event.Content
			}
			w.Flush()
		}

		// save final answer
		h.svc.SaveAnswer(ctx, req.SessionID, fullAnswer)
	})

	return nil
}

func (h *ChatHandler) GetSessions(c *fiber.Ctx) error {
	ctx := c.Context()
	sessions, err := h.svc.ListSessions(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "内部服务错误"})
	}
	return c.JSON(fiber.Map{"sessions": sessions})
}

func (h *ChatHandler) GetSession(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "session id required"})
	}
	ctx := c.Context()
	session, err := h.svc.GetSession(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "内部服务错误"})
	}
	if session == nil {
		return c.Status(404).JSON(fiber.Map{"error": "session not found"})
	}
	return c.JSON(fiber.Map{"session": session})
}

func (h *ChatHandler) DeleteSession(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "session id required"})
	}
	ctx := c.Context()
	if err := h.svc.DeleteSession(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "内部服务错误"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func writeSSE(w *bufio.Writer, event model.SSEEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
}

func sendSSEError(c *fiber.Ctx, msg string) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		writeSSE(w, model.SSEEvent{Type: "error", Content: msg})
		w.Flush()
	})
}
