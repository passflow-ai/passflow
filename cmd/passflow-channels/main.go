package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jaak-ai/passflow-channels/config"
	"github.com/jaak-ai/passflow-channels/domain"
	"github.com/jaak-ai/passflow-channels/input"
	"github.com/jaak-ai/passflow-channels/middleware"
	"github.com/jaak-ai/passflow-channels/output"
	"github.com/jaak-ai/passflow-channels/store"
	"github.com/jaak-ai/passflow-channels/trigger"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🔌 Passflow Channels Service starting...")

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Shutdown signal received")
		cancel()
	}()

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("✅ Connected to Redis: %s", cfg.RedisURL)

	// Rule store
	ruleStore := store.New(ctx, rdb)

	// Webhook event logger
	webhookEventLogger := input.NewWebhookEventLogger(rdb)

	// Dispatcher
	disp := trigger.NewDispatcher(rdb, cfg.StreamKey, cfg.APIURL, cfg.ServiceToken, ruleStore)

	// Cron runner
	cronRunner := input.NewCronRunner(disp, ruleStore)
	cronRunner.Start(ctx)

	// Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		// Enforce a global body-size cap. Webhook requests exceeding
		// MaxWebhookBodyBytes are rejected with 413 before parsing.
		BodyLimit: input.MaxWebhookBodyBytes,
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "passflow-channels"})
	})

	// Input handlers
	webhookHandler := input.NewWebhookHandlerWithLogger(input.DispatcherFunc(func(event domain.Event) {
		disp.Dispatch(ctx, event)
	}), ruleStore, webhookEventLogger)
	webhookHandler.Register(app)

	slackHandler := input.NewSlackHandler(cfg.SlackSigningSecret, disp)
	slackHandler.Register(app)

	// Email (IMAP) poller
	emailPoller := input.NewEmailPoller(cfg, disp)
	emailPoller.Start(ctx)

	// Internal output delivery API
	replyHandler := output.NewReplyHandler(output.New(cfg), ruleStore, rdb)
	internal := app.Group("/internal/v1", middleware.AdminAuth(cfg.ServiceToken))
	replyHandler.RegisterInternal(internal)

	// Admin API — CRUD for trigger rules
	registerAdminAPI(app, ruleStore, webhookEventLogger, cfg)

	// Start server in background
	go func() {
		log.Printf("🚀 Listening on %s", cfg.Addr)
		if err := app.Listen(cfg.Addr); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down channels service...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}

// registerAdminAPI mounts the trigger rule management endpoints.
// All routes require X-Service-Token header. When serviceToken is empty the
// admin API is disabled entirely (fail closed via middleware.AdminAuth).
func registerAdminAPI(app *fiber.App, ruleStore *store.RuleStore, eventLogger *input.WebhookEventLogger, cfg *config.Config) {
	admin := app.Group("/admin/v1", middleware.AdminAuth(cfg.ServiceToken))

	// Create admin handlers
	adminHandlers := middleware.NewAdminHandlers(ruleStore, eventLogger)

	// Get rule details with recent events
	admin.Get("/rules/:ruleId", adminHandlers.GetRule)

	// List all rules
	admin.Get("/rules", func(c *fiber.Ctx) error {
		rules := ruleStore.GetAllRules()
		return c.JSON(middleware.SanitizeRulesForResponse(rules))
	})

	// List rules for a workspace
	admin.Get("/workspaces/:workspaceId/rules", func(c *fiber.Ctx) error {
		rules := ruleStore.GetRulesForWorkspace(c.Params("workspaceId"))
		return c.JSON(middleware.SanitizeRulesForResponse(rules))
	})

	// List triggers (simplified view for dropdowns/assignment)
	triggersHandler := input.NewTriggersListHandler(ruleStore)
	triggersHandler.RegisterOn(admin)

	// Create rule
	admin.Post("/workspaces/:workspaceId/rules", func(c *fiber.Ctx) error {
		var rule domain.TriggerRule
		if err := json.Unmarshal(c.Body(), &rule); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if rule.ID == "" {
			rule.ID = uuid.New().String()
		}
		rule.WorkspaceID = c.Params("workspaceId")
		rule.CreatedAt = time.Now()
		if err := output.ValidateReplyOutputChannel(rule.ChannelType, rule.Action.OutputChannel); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := ruleStore.Upsert(c.Context(), rule); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(middleware.SanitizeRuleForResponse(rule, nil))
	})

	// Partial update rule (toggle enabled, etc.)
	admin.Patch("/rules/:ruleId", func(c *fiber.Ctx) error {
		ruleID := c.Params("ruleId")
		existing := ruleStore.GetByID(ruleID)
		if existing == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rule not found"})
		}

		var patch map[string]interface{}
		if err := json.Unmarshal(c.Body(), &patch); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}

		updated := *existing
		if v, ok := patch["enabled"]; ok {
			if enabled, isBool := v.(bool); isBool {
				updated.Enabled = enabled
			}
		}
		if v, ok := patch["name"]; ok {
			if name, isStr := v.(string); isStr && name != "" {
				updated.Name = name
			}
		}

		if err := ruleStore.Upsert(c.Context(), updated); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(middleware.SanitizeRuleForResponse(updated, nil))
	})

	// Update rule
	admin.Put("/workspaces/:workspaceId/rules/:ruleId", func(c *fiber.Ctx) error {
		var rule domain.TriggerRule
		if err := json.Unmarshal(c.Body(), &rule); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		rule.ID = c.Params("ruleId")
		rule.WorkspaceID = c.Params("workspaceId")
		rule = mergeRuleForUpdate(rule, ruleStore.GetByID(rule.ID))
		if err := output.ValidateReplyOutputChannel(rule.ChannelType, rule.Action.OutputChannel); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := ruleStore.Upsert(c.Context(), rule); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(middleware.SanitizeRuleForResponse(rule, nil))
	})

	// Delete rule
	admin.Delete("/rules/:ruleId", func(c *fiber.Ctx) error {
		if err := ruleStore.Delete(c.Context(), c.Params("ruleId")); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"ok": true})
	})

	// Update auth strategy for a rule
	admin.Patch("/rules/:ruleId/auth", func(c *fiber.Ctx) error {
		ruleID := c.Params("ruleId")
		existing := ruleStore.GetByID(ruleID)
		if existing == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rule not found"})
		}

		var auth domain.AuthStrategy
		if err := json.Unmarshal(c.Body(), &auth); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}

		updated := *existing
		// Generate a new secret if switching to an auth type that requires one
		if auth.Type != domain.AuthNone && auth.Secret == "" && (updated.Auth == nil || updated.Auth.Secret == "") {
			auth.Secret = generateSecret()
		} else if auth.Secret == "" && updated.Auth != nil {
			auth.Secret = updated.Auth.Secret
		}
		updated.Auth = &auth

		if err := ruleStore.Upsert(c.Context(), updated); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Return with masked secret
		resp := updated
		if resp.Auth != nil && resp.Auth.Secret != "" {
			masked := resp.Auth.Secret[:4] + "****"
			resp.Auth = &domain.AuthStrategy{
				Type:   resp.Auth.Type,
				Secret: masked,
				Header: resp.Auth.Header,
			}
		}
		return c.JSON(resp)
	})

	// Regenerate secret for a rule
	admin.Post("/rules/:ruleId/secret/regenerate", func(c *fiber.Ctx) error {
		ruleID := c.Params("ruleId")
		existing := ruleStore.GetByID(ruleID)
		if existing == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rule not found"})
		}

		newSecret := generateSecret()
		updated := *existing
		if updated.Auth == nil {
			updated.Auth = &domain.AuthStrategy{Type: domain.AuthCustomHeader}
		}
		updated.Auth.Secret = newSecret

		if err := ruleStore.Upsert(c.Context(), updated); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"secret":        newSecret,
			"secret_masked": newSecret[:4] + "****",
		})
	})

	// Test webhook — send a real test event through the webhook handler
	admin.Post("/rules/:ruleId/test", func(c *fiber.Ctx) error {
		ruleID := c.Params("ruleId")
		existing := ruleStore.GetByID(ruleID)
		if existing == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rule not found"})
		}

		if !existing.Enabled {
			return c.JSON(fiber.Map{
				"success":          false,
				"error_message":    "Rule is disabled",
				"status_code":      0,
				"response_time_ms": 0,
			})
		}

		// Build a local request to the webhook handler
		testPayload := `{"test":true,"source":"channels-admin","rule_id":"` + ruleID + `"}`
		webhookURL := fmt.Sprintf("http://127.0.0.1%s/webhook/%s", cfg.Addr, existing.WorkspaceID)

		req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, webhookURL,
			strings.NewReader(testPayload))
		if err != nil {
			return c.JSON(fiber.Map{
				"success":          false,
				"error_message":    err.Error(),
				"status_code":      0,
				"response_time_ms": 0,
			})
		}
		req.Header.Set("Content-Type", "application/json")
		// Set auth header so the webhook handler can authenticate
		auth := existing.Auth
		if auth != nil && auth.Secret != "" {
			switch auth.Type {
			case domain.AuthCustomHeader:
				header := auth.Header
				if header == "" {
					header = "X-Webhook-Secret"
				}
				req.Header.Set(header, auth.Secret)
			default:
				req.Header.Set("X-Webhook-Secret", auth.Secret)
			}
		}

		client := &http.Client{Timeout: 10 * time.Second}
		start := time.Now()
		resp, err := client.Do(req)
		elapsed := time.Since(start).Milliseconds()

		if err != nil {
			return c.JSON(fiber.Map{
				"success":          false,
				"error_message":    err.Error(),
				"status_code":      0,
				"response_time_ms": elapsed,
			})
		}
		defer resp.Body.Close()

		return c.JSON(fiber.Map{
			"success":          resp.StatusCode >= 200 && resp.StatusCode < 300,
			"status_code":      resp.StatusCode,
			"response_time_ms": elapsed,
		})
	})
}

// generateSecret creates a 32-byte hex-encoded random secret.
func generateSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}
