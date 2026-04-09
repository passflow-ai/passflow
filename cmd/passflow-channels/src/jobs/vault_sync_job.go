package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
)

// Stream and consumer group constants for vault sync jobs.
const (
	StreamVaultSyncJobs     = "vault:sync:jobs"
	ConsumerGroupVaultSync  = "vault-sync-workers"
	DefaultMaxRetries       = 3
	DefaultRetryInterval    = 5 * time.Second
	DefaultBlockTimeout     = 5 * time.Second
	DefaultConsumerName     = "vault-sync-consumer"
)

// VaultSyncPayload represents the job payload enqueued by passflow-api.
type VaultSyncPayload struct {
	OrgID       string `json:"org_id"`
	WorkspaceID string `json:"workspace_id"`
	ConfigID    string `json:"config_id"`
}

// QueueJob represents a job from the Redis Stream queue.
type QueueJob struct {
	ID      string
	Payload string
}

// VaultConfig represents the vault configuration returned by passflow-api.
type VaultConfig struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	OrgID       string `json:"org_id"`
	RepoURL     string `json:"repo_url"`
	LocalPath   string `json:"local_path"`
	Branch      string `json:"branch"`
}

// SyncStatus represents the status update sent to passflow-api.
type SyncStatus struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	SyncedAt  string `json:"synced_at,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

// GitExecutor abstracts git operations for testing.
type GitExecutor interface {
	Pull(ctx context.Context, path string) error
}

// VaultSyncJobOptions configures the vault sync job behavior.
type VaultSyncJobOptions struct {
	MaxRetries    int
	RetryInterval time.Duration
	BlockTimeout  time.Duration
	ConsumerName  string
}

// VaultSyncJob processes vault synchronization jobs from the Redis Stream.
type VaultSyncJob struct {
	apiClient   *APIClient
	redisClient *redis.Client
	gitExecutor GitExecutor
	logger      *slog.Logger
	options     VaultSyncJobOptions
}

// NewVaultSyncJob creates a new VaultSyncJob with default options.
func NewVaultSyncJob(
	apiClient *APIClient,
	redisClient *redis.Client,
	gitExecutor GitExecutor,
	logger *slog.Logger,
) *VaultSyncJob {
	return NewVaultSyncJobWithOptions(apiClient, redisClient, gitExecutor, logger, VaultSyncJobOptions{
		MaxRetries:    DefaultMaxRetries,
		RetryInterval: DefaultRetryInterval,
		BlockTimeout:  DefaultBlockTimeout,
		ConsumerName:  DefaultConsumerName,
	})
}

// NewVaultSyncJobWithOptions creates a new VaultSyncJob with custom options.
func NewVaultSyncJobWithOptions(
	apiClient *APIClient,
	redisClient *redis.Client,
	gitExecutor GitExecutor,
	logger *slog.Logger,
	options VaultSyncJobOptions,
) *VaultSyncJob {
	if logger == nil {
		logger = slog.Default()
	}
	if options.MaxRetries == 0 {
		options.MaxRetries = DefaultMaxRetries
	}
	if options.RetryInterval == 0 {
		options.RetryInterval = DefaultRetryInterval
	}
	if options.BlockTimeout == 0 {
		options.BlockTimeout = DefaultBlockTimeout
	}
	if options.ConsumerName == "" {
		options.ConsumerName = DefaultConsumerName
	}

	return &VaultSyncJob{
		apiClient:   apiClient,
		redisClient: redisClient,
		gitExecutor: gitExecutor,
		logger:      logger,
		options:     options,
	}
}

// Start begins consuming jobs from the Redis Stream.
// It blocks until the context is cancelled.
func (j *VaultSyncJob) Start(ctx context.Context) error {
	j.logger.Info("starting vault sync job consumer",
		slog.String("stream", StreamVaultSyncJobs),
		slog.String("consumer_group", ConsumerGroupVaultSync),
		slog.String("consumer_name", j.options.ConsumerName),
	)

	// Create consumer group if it doesn't exist
	err := j.ensureConsumerGroup(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure consumer group: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("vault sync job consumer stopping: context cancelled")
			return ctx.Err()
		default:
			err := j.consumeMessages(ctx)
			if err != nil && ctx.Err() == nil {
				j.logger.Error("error consuming messages", slog.String("error", err.Error()))
				time.Sleep(time.Second) // Back off on errors
			}
		}
	}
}

// ensureConsumerGroup creates the consumer group if it doesn't exist.
func (j *VaultSyncJob) ensureConsumerGroup(ctx context.Context) error {
	err := j.redisClient.XGroupCreateMkStream(ctx, StreamVaultSyncJobs, ConsumerGroupVaultSync, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

// consumeMessages reads and processes messages from the stream.
func (j *VaultSyncJob) consumeMessages(ctx context.Context) error {
	streams, err := j.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroupVaultSync,
		Consumer: j.options.ConsumerName,
		Streams:  []string{StreamVaultSyncJobs, ">"},
		Count:    1,
		Block:    j.options.BlockTimeout,
	}).Result()

	if err == redis.Nil {
		return nil // No messages available
	}
	if err != nil {
		return fmt.Errorf("XReadGroup error: %w", err)
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			payload, ok := msg.Values["payload"].(string)
			if !ok {
				j.logger.Warn("invalid message payload", slog.String("message_id", msg.ID))
				j.ackMessage(ctx, msg.ID)
				continue
			}

			queueJob := &QueueJob{
				ID:      msg.ID,
				Payload: payload,
			}

			err := j.ProcessWithRetry(ctx, queueJob)
			if err != nil {
				j.logger.Error("failed to process job",
					slog.String("job_id", queueJob.ID),
					slog.String("error", err.Error()),
				)
			}

			// ACK the message regardless of success/failure to avoid reprocessing
			j.ackMessage(ctx, msg.ID)
		}
	}

	return nil
}

// ackMessage acknowledges a message in the consumer group.
func (j *VaultSyncJob) ackMessage(ctx context.Context, messageID string) {
	err := j.redisClient.XAck(ctx, StreamVaultSyncJobs, ConsumerGroupVaultSync, messageID).Err()
	if err != nil {
		j.logger.Error("failed to ack message",
			slog.String("message_id", messageID),
			slog.String("error", err.Error()),
		)
	}
}

// Process processes a single vault sync job.
func (j *VaultSyncJob) Process(ctx context.Context, job *QueueJob) error {
	j.logger.Info("processing vault sync job", slog.String("job_id", job.ID))

	// 1. Deserialize payload
	var payload VaultSyncPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 2. Validate required fields
	if payload.ConfigID == "" {
		return fmt.Errorf("config_id is required")
	}

	// 3. Get vault config from passflow-api
	config, err := j.apiClient.GetVaultConfig(ctx, payload.ConfigID)
	if err != nil {
		return fmt.Errorf("failed to get vault config: %w", err)
	}

	// 4. Execute git pull
	err = j.gitExecutor.Pull(ctx, config.LocalPath)
	if err != nil {
		// Report failure to API
		j.reportSyncStatus(ctx, payload.ConfigID, SyncStatus{
			Status:    "failed",
			Message:   err.Error(),
			ErrorCode: "GIT_PULL_FAILED",
		})
		return fmt.Errorf("git pull failed: %w", err)
	}

	// 5. Report success to passflow-api
	j.reportSyncStatus(ctx, payload.ConfigID, SyncStatus{
		Status:   "success",
		SyncedAt: time.Now().UTC().Format(time.RFC3339),
	})

	j.logger.Info("vault sync job completed successfully",
		slog.String("job_id", job.ID),
		slog.String("config_id", payload.ConfigID),
	)

	return nil
}

// ProcessWithRetry processes a job with retry logic.
func (j *VaultSyncJob) ProcessWithRetry(ctx context.Context, job *QueueJob) error {
	var lastErr error

	for attempt := 1; attempt <= j.options.MaxRetries; attempt++ {
		err := j.Process(ctx, job)
		if err == nil {
			return nil
		}

		lastErr = err
		j.logger.Warn("vault sync job failed, will retry",
			slog.String("job_id", job.ID),
			slog.Int("attempt", attempt),
			slog.Int("max_retries", j.options.MaxRetries),
			slog.String("error", err.Error()),
		)

		if attempt < j.options.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(j.options.RetryInterval):
				// Continue to next attempt
			}
		}
	}

	return lastErr
}

// reportSyncStatus sends the sync status to passflow-api.
func (j *VaultSyncJob) reportSyncStatus(ctx context.Context, configID string, status SyncStatus) {
	err := j.apiClient.UpdateSyncStatus(ctx, configID, status)
	if err != nil {
		j.logger.Error("failed to report sync status",
			slog.String("config_id", configID),
			slog.String("status", status.Status),
			slog.String("error", err.Error()),
		)
	}
}
