package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVaultSyncJob_Process_Success tests successful processing of a vault sync job.
func TestVaultSyncJob_Process_Success(t *testing.T) {
	// Setup mock API server
	var getConfigCalled, gitPullCalled, updateStatusCalled int32
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/internal/v1/vault/config/config-123" && r.Method == http.MethodGet:
			atomic.AddInt32(&getConfigCalled, 1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           "config-123",
				"workspace_id": "ws-456",
				"org_id":       "org-789",
				"repo_url":     "https://github.com/test/repo.git",
				"local_path":   "/tmp/vaults/test-vault",
				"branch":       "main",
			})

		case r.URL.Path == "/internal/v1/vault/config/config-123/sync-status" && r.Method == http.MethodPut:
			atomic.AddInt32(&updateStatusCalled, 1)
			body, _ := io.ReadAll(r.Body)
			var status map[string]interface{}
			json.Unmarshal(body, &status)
			assert.Equal(t, "success", status["status"])
			w.WriteHeader(http.StatusOK)

		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	// Setup mock git executor
	mockGit := &mockGitExecutor{
		pullFunc: func(ctx context.Context, path string) error {
			atomic.AddInt32(&gitPullCalled, 1)
			assert.Equal(t, "/tmp/vaults/test-vault", path)
			return nil
		},
	}

	// Setup Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	// Create job
	job := NewVaultSyncJob(
		NewAPIClient(mockAPI.URL, "test-service-token"),
		rdb,
		mockGit,
		nil, // logger
	)

	// Create job payload
	payload := VaultSyncPayload{
		OrgID:       "org-789",
		WorkspaceID: "ws-456",
		ConfigID:    "config-123",
	}
	payloadBytes, _ := json.Marshal(payload)

	queueJob := &QueueJob{
		ID:      "job-001",
		Payload: string(payloadBytes),
	}

	// Process the job
	err = job.Process(context.Background(), queueJob)
	require.NoError(t, err)

	// Verify all API calls were made
	assert.Equal(t, int32(1), atomic.LoadInt32(&getConfigCalled), "GetConfig should be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&gitPullCalled), "Git pull should be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&updateStatusCalled), "UpdateStatus should be called once")
}

// TestVaultSyncJob_Process_GitError tests handling of git errors during sync.
func TestVaultSyncJob_Process_GitError(t *testing.T) {
	var updateStatusCalled int32
	var reportedStatus string

	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/internal/v1/vault/config/config-123" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           "config-123",
				"workspace_id": "ws-456",
				"org_id":       "org-789",
				"repo_url":     "https://github.com/test/repo.git",
				"local_path":   "/tmp/vaults/test-vault",
				"branch":       "main",
			})

		case r.URL.Path == "/internal/v1/vault/config/config-123/sync-status" && r.Method == http.MethodPut:
			atomic.AddInt32(&updateStatusCalled, 1)
			body, _ := io.ReadAll(r.Body)
			var status map[string]interface{}
			json.Unmarshal(body, &status)
			reportedStatus = status["status"].(string)
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	// Mock git executor that returns an error
	mockGit := &mockGitExecutor{
		pullFunc: func(ctx context.Context, path string) error {
			return fmt.Errorf("git pull failed: authentication required")
		},
	}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJob(
		NewAPIClient(mockAPI.URL, "test-token"),
		rdb,
		mockGit,
		nil,
	)

	payload := VaultSyncPayload{
		OrgID:       "org-789",
		WorkspaceID: "ws-456",
		ConfigID:    "config-123",
	}
	payloadBytes, _ := json.Marshal(payload)

	queueJob := &QueueJob{
		ID:      "job-002",
		Payload: string(payloadBytes),
	}

	err = job.Process(context.Background(), queueJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git pull failed")

	// Status should be reported as failed
	assert.Equal(t, int32(1), atomic.LoadInt32(&updateStatusCalled))
	assert.Equal(t, "failed", reportedStatus)
}

// TestVaultSyncJob_Process_APIError tests handling of API errors when fetching config.
func TestVaultSyncJob_Process_APIError(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/internal/v1/vault/config/config-123" && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "config not found",
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	mockGit := &mockGitExecutor{}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJob(
		NewAPIClient(mockAPI.URL, "test-token"),
		rdb,
		mockGit,
		nil,
	)

	payload := VaultSyncPayload{
		OrgID:       "org-789",
		WorkspaceID: "ws-456",
		ConfigID:    "config-123",
	}
	payloadBytes, _ := json.Marshal(payload)

	queueJob := &QueueJob{
		ID:      "job-003",
		Payload: string(payloadBytes),
	}

	err = job.Process(context.Background(), queueJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get vault config")
}

// TestVaultSyncJob_Process_InvalidPayload tests handling of invalid job payloads.
func TestVaultSyncJob_Process_InvalidPayload(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJob(
		NewAPIClient("http://localhost:8080", "test-token"),
		rdb,
		&mockGitExecutor{},
		nil,
	)

	queueJob := &QueueJob{
		ID:      "job-004",
		Payload: "invalid json{",
	}

	err = job.Process(context.Background(), queueJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal payload")
}

// TestVaultSyncJob_Process_MissingConfigID tests validation of required fields.
func TestVaultSyncJob_Process_MissingConfigID(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJob(
		NewAPIClient("http://localhost:8080", "test-token"),
		rdb,
		&mockGitExecutor{},
		nil,
	)

	payload := VaultSyncPayload{
		OrgID:       "org-789",
		WorkspaceID: "ws-456",
		ConfigID:    "", // Missing required field
	}
	payloadBytes, _ := json.Marshal(payload)

	queueJob := &QueueJob{
		ID:      "job-005",
		Payload: string(payloadBytes),
	}

	err = job.Process(context.Background(), queueJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config_id is required")
}

// TestVaultSyncJob_Start tests the Redis Stream consumer startup.
func TestVaultSyncJob_Start(t *testing.T) {
	var processedJobs int32

	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           "config-123",
				"workspace_id": "ws-456",
				"org_id":       "org-789",
				"repo_url":     "https://github.com/test/repo.git",
				"local_path":   "/tmp/vaults/test-vault",
				"branch":       "main",
			})

		case r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	mockGit := &mockGitExecutor{
		pullFunc: func(ctx context.Context, path string) error {
			atomic.AddInt32(&processedJobs, 1)
			return nil
		},
	}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJob(
		NewAPIClient(mockAPI.URL, "test-token"),
		rdb,
		mockGit,
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the consumer in background
	go func() {
		err := job.Start(ctx)
		if err != nil && ctx.Err() == nil {
			t.Errorf("Start() error = %v", err)
		}
	}()

	// Give the consumer time to start
	time.Sleep(100 * time.Millisecond)

	// Enqueue a job to the stream
	payload := VaultSyncPayload{
		OrgID:       "org-789",
		WorkspaceID: "ws-456",
		ConfigID:    "config-123",
	}
	payloadBytes, _ := json.Marshal(payload)

	err = rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamVaultSyncJobs,
		Values: map[string]interface{}{
			"payload": string(payloadBytes),
		},
	}).Err()
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Verify job was processed
	assert.GreaterOrEqual(t, atomic.LoadInt32(&processedJobs), int32(1), "At least one job should be processed")

	cancel()
}

// TestVaultSyncJob_Retry tests that failed jobs are retried.
func TestVaultSyncJob_Retry(t *testing.T) {
	var attempts int32

	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           "config-123",
				"workspace_id": "ws-456",
				"org_id":       "org-789",
				"repo_url":     "https://github.com/test/repo.git",
				"local_path":   "/tmp/vaults/test-vault",
				"branch":       "main",
			})

		case r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	// Git executor fails the first 2 attempts, succeeds on the 3rd
	mockGit := &mockGitExecutor{
		pullFunc: func(ctx context.Context, path string) error {
			attempt := atomic.AddInt32(&attempts, 1)
			if attempt < 3 {
				return fmt.Errorf("temporary error on attempt %d", attempt)
			}
			return nil
		},
	}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJobWithOptions(
		NewAPIClient(mockAPI.URL, "test-token"),
		rdb,
		mockGit,
		nil,
		VaultSyncJobOptions{
			MaxRetries:    3,
			RetryInterval: 10 * time.Millisecond,
		},
	)

	payload := VaultSyncPayload{
		OrgID:       "org-789",
		WorkspaceID: "ws-456",
		ConfigID:    "config-123",
	}
	payloadBytes, _ := json.Marshal(payload)

	queueJob := &QueueJob{
		ID:      "job-retry",
		Payload: string(payloadBytes),
	}

	err = job.ProcessWithRetry(context.Background(), queueJob)
	require.NoError(t, err)

	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts), "Should have attempted 3 times")
}

// TestVaultSyncJob_ProcessWithRetry_ExhaustsRetries tests that all retries are exhausted.
func TestVaultSyncJob_ProcessWithRetry_ExhaustsRetries(t *testing.T) {
	var attempts int32

	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           "config-123",
				"workspace_id": "ws-456",
				"org_id":       "org-789",
				"local_path":   "/tmp/vaults/test-vault",
			})

		case r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	// Git executor always fails
	mockGit := &mockGitExecutor{
		pullFunc: func(ctx context.Context, path string) error {
			atomic.AddInt32(&attempts, 1)
			return fmt.Errorf("persistent error")
		},
	}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJobWithOptions(
		NewAPIClient(mockAPI.URL, "test-token"),
		rdb,
		mockGit,
		nil,
		VaultSyncJobOptions{
			MaxRetries:    3,
			RetryInterval: 10 * time.Millisecond,
		},
	)

	payload := VaultSyncPayload{ConfigID: "config-123"}
	payloadBytes, _ := json.Marshal(payload)

	queueJob := &QueueJob{
		ID:      "job-exhaust",
		Payload: string(payloadBytes),
	}

	err = job.ProcessWithRetry(context.Background(), queueJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "persistent error")
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts), "Should have attempted exactly 3 times")
}

// TestVaultSyncJob_ProcessWithRetry_ContextCancelled tests context cancellation during retry.
func TestVaultSyncJob_ProcessWithRetry_ContextCancelled(t *testing.T) {
	var attempts int32

	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         "config-123",
				"local_path": "/tmp/vaults/test-vault",
			})

		case r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	mockGit := &mockGitExecutor{
		pullFunc: func(ctx context.Context, path string) error {
			atomic.AddInt32(&attempts, 1)
			return fmt.Errorf("error")
		},
	}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJobWithOptions(
		NewAPIClient(mockAPI.URL, "test-token"),
		rdb,
		mockGit,
		nil,
		VaultSyncJobOptions{
			MaxRetries:    5,
			RetryInterval: 100 * time.Millisecond,
		},
	)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	payload := VaultSyncPayload{ConfigID: "config-123"}
	payloadBytes, _ := json.Marshal(payload)

	queueJob := &QueueJob{
		ID:      "job-cancel",
		Payload: string(payloadBytes),
	}

	err = job.ProcessWithRetry(ctx, queueJob)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	// Should have attempted at most 2 times before cancellation
	assert.LessOrEqual(t, atomic.LoadInt32(&attempts), int32(2))
}

// TestAPIClient_GetVaultConfig tests the API client GetVaultConfig method.
func TestAPIClient_GetVaultConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/internal/v1/vault/config/test-config", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "test-token", r.Header.Get("X-Service-Token"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           "test-config",
			"workspace_id": "ws-1",
			"org_id":       "org-1",
			"repo_url":     "https://github.com/test/repo.git",
			"local_path":   "/tmp/test",
			"branch":       "main",
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-token")
	config, err := client.GetVaultConfig(context.Background(), "test-config")

	require.NoError(t, err)
	assert.Equal(t, "test-config", config.ID)
	assert.Equal(t, "ws-1", config.WorkspaceID)
	assert.Equal(t, "org-1", config.OrgID)
	assert.Equal(t, "/tmp/test", config.LocalPath)
}

// TestAPIClient_UpdateSyncStatus tests the API client UpdateSyncStatus method.
func TestAPIClient_UpdateSyncStatus(t *testing.T) {
	var receivedStatus SyncStatus
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/internal/v1/vault/config/test-config/sync-status", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedStatus)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-token")
	status := SyncStatus{
		Status:   "success",
		SyncedAt: "2026-03-17T12:00:00Z",
	}

	err := client.UpdateSyncStatus(context.Background(), "test-config", status)

	require.NoError(t, err)
	assert.Equal(t, "success", receivedStatus.Status)
	assert.Equal(t, "2026-03-17T12:00:00Z", receivedStatus.SyncedAt)
}

// TestAPIClient_UpdateSyncStatus_Error tests error handling in UpdateSyncStatus.
func TestAPIClient_UpdateSyncStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-token")
	status := SyncStatus{Status: "success"}

	err := client.UpdateSyncStatus(context.Background(), "test-config", status)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 500")
}

// TestVaultSyncJob_DefaultOptions tests that default options are applied correctly.
func TestVaultSyncJob_DefaultOptions(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	job := NewVaultSyncJob(
		NewAPIClient("http://localhost:8080", "test-token"),
		rdb,
		&mockGitExecutor{},
		nil,
	)

	assert.Equal(t, DefaultMaxRetries, job.options.MaxRetries)
	assert.Equal(t, DefaultRetryInterval, job.options.RetryInterval)
	assert.Equal(t, DefaultBlockTimeout, job.options.BlockTimeout)
	assert.Equal(t, DefaultConsumerName, job.options.ConsumerName)
}

// mockGitExecutor is a test double for GitExecutor.
type mockGitExecutor struct {
	pullFunc func(ctx context.Context, path string) error
}

func (m *mockGitExecutor) Pull(ctx context.Context, path string) error {
	if m.pullFunc != nil {
		return m.pullFunc(ctx, path)
	}
	return nil
}
