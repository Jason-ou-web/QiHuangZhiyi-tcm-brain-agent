package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agri-qa-system/config"
	"agri-qa-system/internal/handler"
	"agri-qa-system/internal/llm"
	"agri-qa-system/internal/middleware"
	"agri-qa-system/internal/rag"
	"agri-qa-system/internal/service"
	agentSvc "agri-qa-system/internal/service/agent"
	"agri-qa-system/internal/store"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env from backend/ directory
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// initialize stores with graceful degradation for MVP testing
	qdrantStore, err := store.NewQdrantStore(cfg)
	if err != nil {
		log.Printf("WARNING: Qdrant not available: %v (server starts degraded)", err)
		qdrantStore = nil
	}

	redisStore, err := store.NewRedisStore(cfg)
	if err != nil {
		log.Printf("WARNING: Redis not available: %v (server starts degraded)", err)
		redisStore = nil
	}

	embedClient := store.NewEmbeddingClient(cfg.EmbeddingURL)
	rerankerClient := store.NewRerankerClient(cfg.RerankerURL)
	llmClient := llm.NewClient(cfg)

	var ragPipeline *rag.Pipeline
	var chatSvc *service.ChatService
	var agentCoordinator *agentSvc.Coordinator
	var chatHandler *handler.ChatHandler
	var agentHandler *handler.AgentHandler

	if qdrantStore != nil && redisStore != nil {
		ragPipeline = rag.NewPipeline(qdrantStore, embedClient, rerankerClient)
		chatSvc = service.NewChatService(ragPipeline, llmClient, redisStore)
		agentCoordinator = agentSvc.NewCoordinator(ragPipeline, llmClient, redisStore)
		chatHandler = handler.NewChatHandler(chatSvc)
		agentHandler = handler.NewAgentHandler(agentCoordinator)
	}

	// setup fiber app
	app := fiber.New(fiber.Config{
		AppName:      "tcm-agent-system",
		BodyLimit:    1 * 1024 * 1024, // 1MB
		ErrorHandler: defaultErrorHandler,
	})

	// global middleware
	app.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// routes
	app.Get("/health", func(c *fiber.Ctx) error {
		status := "ok"
		details := map[string]string{
			"qdrant": "ok",
			"redis":  "ok",
		}
		if qdrantStore == nil {
			details["qdrant"] = "unavailable"
			status = "degraded"
		}
		if redisStore == nil {
			details["redis"] = "unavailable"
			status = "degraded"
		}
		return c.JSON(fiber.Map{"status": status, "details": details})
	})

	api := app.Group("/api/v1")

	// chat endpoints: moderate rate limit
	chatGroup := api.Group("", middleware.RateLimitMiddleware(20, 1*time.Minute))
	if chatHandler != nil {
		chatGroup.Post("/chat", chatHandler.Chat)
		chatGroup.Post("/chat/stream", chatHandler.ChatStream)
	}
	// agent endpoints
	if agentHandler != nil {
		chatGroup.Post("/agent/analyze", agentHandler.Analyze)
		chatGroup.Post("/agent/chat/stream", agentHandler.ChatStream)
	}
	// session endpoints: higher rate limit
	sessionGroup := api.Group("", middleware.RateLimitMiddleware(60, 1*time.Minute))
	if chatHandler != nil {
		sessionGroup.Get("/sessions", chatHandler.GetSessions)
		sessionGroup.Get("/sessions/:id", chatHandler.GetSession)
		sessionGroup.Delete("/sessions/:id", chatHandler.DeleteSession)
	}

	log.Printf("Server starting on :%s", cfg.ServerPort)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server gracefully...")

		shutdownTimeout := 10 * time.Second
		if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}

		if qdrantStore != nil {
			if err := qdrantStore.Close(); err != nil {
				log.Printf("Qdrant close error: %v", err)
			}
		}
		if redisStore != nil {
			if err := redisStore.Close(); err != nil {
				log.Printf("Redis close error: %v", err)
			}
		}
		log.Println("Server stopped")
	}()

	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "内部服务错误"
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}
	return c.Status(code).JSON(fiber.Map{"error": msg})
}
