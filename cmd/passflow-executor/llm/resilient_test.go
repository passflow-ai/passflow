package llm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockClient is a test double for llm.Client.
type mockClient struct {
	responses []*Response
	errors    []error
	callCount int
}

func (m *mockClient) Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error) {
	idx := m.callCount
	m.callCount++
	if idx < len(m.errors) && m.errors[idx] != nil {
		return nil, m.errors[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return &Response{Content: "default"}, nil
}

// Test that ResilientClient implements the Client interface
func TestResilientClient_ImplementsInterface(t *testing.T) {
	var _ Client = (*ResilientClient)(nil)
}

func TestNewResilientClient_DefaultConfig(t *testing.T) {
	primary := &mockClient{}
	client := NewResilientClient(primary, nil, ResilientConfig{})

	if client.primary != primary {
		t.Error("primary client not set correctly")
	}
	if client.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", client.maxRetries)
	}
	if client.baseDelay != time.Second {
		t.Errorf("baseDelay = %v, want %v", client.baseDelay, time.Second)
	}
	if client.maxDelay != 30*time.Second {
		t.Errorf("maxDelay = %v, want %v", client.maxDelay, 30*time.Second)
	}
}

func TestNewResilientClient_CustomConfig(t *testing.T) {
	primary := &mockClient{}
	config := ResilientConfig{
		MaxRetries: 5,
		BaseDelay:  2 * time.Second,
		MaxDelay:   60 * time.Second,
	}
	client := NewResilientClient(primary, nil, config)

	if client.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", client.maxRetries)
	}
	if client.baseDelay != 2*time.Second {
		t.Errorf("baseDelay = %v, want %v", client.baseDelay, 2*time.Second)
	}
	if client.maxDelay != 60*time.Second {
		t.Errorf("maxDelay = %v, want %v", client.maxDelay, 60*time.Second)
	}
}

func TestResilientClient_Complete_Success(t *testing.T) {
	expected := &Response{Content: "Hello", Usage: Usage{TotalTokens: 10}}
	primary := &mockClient{
		responses: []*Response{expected},
	}
	client := NewResilientClient(primary, nil, ResilientConfig{})

	ctx := context.Background()
	got, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err != nil {
		t.Fatalf("Complete() error = %v, want nil", err)
	}
	if got.Content != expected.Content {
		t.Errorf("Content = %q, want %q", got.Content, expected.Content)
	}
	if primary.callCount != 1 {
		t.Errorf("callCount = %d, want 1", primary.callCount)
	}
}

func TestResilientClient_Complete_RetryOnTransientError(t *testing.T) {
	transientErr := &RetryableError{Err: errors.New("connection timeout"), Retryable: true}
	expected := &Response{Content: "Success after retry"}
	primary := &mockClient{
		errors:    []error{transientErr, transientErr, nil},
		responses: []*Response{nil, nil, expected},
	}
	config := ResilientConfig{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond, // fast for tests
		MaxDelay:   10 * time.Millisecond,
	}
	client := NewResilientClient(primary, nil, config)

	ctx := context.Background()
	got, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err != nil {
		t.Fatalf("Complete() error = %v, want nil", err)
	}
	if got.Content != expected.Content {
		t.Errorf("Content = %q, want %q", got.Content, expected.Content)
	}
	if primary.callCount != 3 {
		t.Errorf("callCount = %d, want 3 (2 retries + 1 success)", primary.callCount)
	}
}

func TestResilientClient_Complete_NoRetryOnNonRetryableError(t *testing.T) {
	nonRetryableErr := &RetryableError{Err: errors.New("invalid API key"), Retryable: false}
	primary := &mockClient{
		errors: []error{nonRetryableErr},
	}
	config := ResilientConfig{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
	}
	client := NewResilientClient(primary, nil, config)

	ctx := context.Background()
	_, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err == nil {
		t.Fatal("Complete() should return an error")
	}
	if primary.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no retries for non-retryable)", primary.callCount)
	}
}

func TestResilientClient_Complete_FallbackOnExhaustedRetries(t *testing.T) {
	transientErr := &RetryableError{Err: errors.New("rate limited"), Retryable: true}
	primary := &mockClient{
		errors: []error{transientErr, transientErr, transientErr, transientErr},
	}
	fallbackExpected := &Response{Content: "Fallback success"}
	fallback := &mockClient{
		responses: []*Response{fallbackExpected},
	}
	config := ResilientConfig{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	}
	client := NewResilientClient(primary, []Client{fallback}, config)

	ctx := context.Background()
	got, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err != nil {
		t.Fatalf("Complete() error = %v, want nil", err)
	}
	if got.Content != fallbackExpected.Content {
		t.Errorf("Content = %q, want %q", got.Content, fallbackExpected.Content)
	}
	// Primary should be called maxRetries+1 times (initial + retries)
	if primary.callCount != 4 {
		t.Errorf("primary.callCount = %d, want 4", primary.callCount)
	}
	if fallback.callCount != 1 {
		t.Errorf("fallback.callCount = %d, want 1", fallback.callCount)
	}
}

func TestResilientClient_Complete_MultipleFallbacks(t *testing.T) {
	transientErr := &RetryableError{Err: errors.New("service unavailable"), Retryable: true}
	primary := &mockClient{
		errors: []error{transientErr, transientErr, transientErr, transientErr},
	}
	fallback1 := &mockClient{
		errors: []error{transientErr, transientErr, transientErr, transientErr},
	}
	fallbackExpected := &Response{Content: "Second fallback success"}
	fallback2 := &mockClient{
		responses: []*Response{fallbackExpected},
	}
	config := ResilientConfig{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	}
	client := NewResilientClient(primary, []Client{fallback1, fallback2}, config)

	ctx := context.Background()
	got, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err != nil {
		t.Fatalf("Complete() error = %v, want nil", err)
	}
	if got.Content != fallbackExpected.Content {
		t.Errorf("Content = %q, want %q", got.Content, fallbackExpected.Content)
	}
}

func TestResilientClient_Complete_AllProvidersFail(t *testing.T) {
	transientErr := &RetryableError{Err: errors.New("service unavailable"), Retryable: true}
	primary := &mockClient{
		errors: []error{transientErr, transientErr, transientErr, transientErr},
	}
	fallback := &mockClient{
		errors: []error{transientErr, transientErr, transientErr, transientErr},
	}
	config := ResilientConfig{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	}
	client := NewResilientClient(primary, []Client{fallback}, config)

	ctx := context.Background()
	_, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err == nil {
		t.Fatal("Complete() should return an error when all providers fail")
	}
}

func TestResilientClient_Complete_ContextCancellation(t *testing.T) {
	transientErr := &RetryableError{Err: errors.New("timeout"), Retryable: true}
	primary := &mockClient{
		errors: []error{transientErr, transientErr, transientErr, transientErr},
	}
	config := ResilientConfig{
		MaxRetries: 10,          // many retries
		BaseDelay:  time.Second, // long delay
		MaxDelay:   30 * time.Second,
	}
	client := NewResilientClient(primary, nil, config)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after a short time
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	_, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err == nil {
		t.Fatal("Complete() should return an error on context cancellation")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error should be context.Canceled, got %v", err)
	}
}

func TestResilientClient_Complete_ContextTimeout(t *testing.T) {
	transientErr := &RetryableError{Err: errors.New("timeout"), Retryable: true}
	primary := &mockClient{
		errors: []error{transientErr, transientErr, transientErr, transientErr},
	}
	config := ResilientConfig{
		MaxRetries: 10,
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
	}
	client := NewResilientClient(primary, nil, config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := client.Complete(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)

	if err == nil {
		t.Fatal("Complete() should return an error on context timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error should be context.DeadlineExceeded, got %v", err)
	}
}

func TestCalculateBackoff(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	maxDelay := 10 * time.Second

	tests := []struct {
		attempt int
		wantMin time.Duration
		wantMax time.Duration
	}{
		{0, 100 * time.Millisecond, 130 * time.Millisecond},             // 100ms * 2^0 = 100ms + up to 20% jitter
		{1, 200 * time.Millisecond, 250 * time.Millisecond},             // 100ms * 2^1 = 200ms + up to 20% jitter
		{2, 400 * time.Millisecond, 490 * time.Millisecond},             // 100ms * 2^2 = 400ms + up to 20% jitter
		{3, 800 * time.Millisecond, 970 * time.Millisecond},             // 100ms * 2^3 = 800ms + up to 20% jitter
		{10, maxDelay, maxDelay + 2*time.Second + 100*time.Millisecond}, // capped at maxDelay + up to 20% jitter
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calculateBackoff(tt.attempt, baseDelay, maxDelay)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateBackoff(%d) = %v, want between %v and %v", tt.attempt, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "RetryableError with Retryable=true",
			err:  &RetryableError{Err: errors.New("timeout"), Retryable: true},
			want: true,
		},
		{
			name: "RetryableError with Retryable=false",
			err:  &RetryableError{Err: errors.New("invalid key"), Retryable: false},
			want: false,
		},
		{
			name: "context.Canceled",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "context.DeadlineExceeded",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "wrapped context.Canceled",
			err:  errors.New("wrapped: " + context.Canceled.Error()),
			want: true, // plain error is retryable by default
		},
		{
			name: "plain error is retryable by default",
			err:  errors.New("unknown error"),
			want: true,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryableError_Error(t *testing.T) {
	err := &RetryableError{Err: errors.New("connection failed"), Retryable: true}
	if err.Error() != "connection failed" {
		t.Errorf("Error() = %q, want %q", err.Error(), "connection failed")
	}
}

func TestRetryableError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &RetryableError{Err: inner, Retryable: true}
	if !errors.Is(err, inner) {
		t.Error("Unwrap() should return the inner error")
	}
}

func TestResilientClient_Complete_PassesMessagesAndTools(t *testing.T) {
	var capturedMessages []Message
	var capturedTools []ToolDefinition

	primary := &mockClient{
		responses: []*Response{{Content: "OK"}},
	}

	// Wrap primary to capture arguments
	wrapper := &captureClient{
		inner:            primary,
		capturedMessages: &capturedMessages,
		capturedTools:    &capturedTools,
	}

	client := NewResilientClient(wrapper, nil, ResilientConfig{})

	messages := []Message{
		{Role: "system", Content: "You are helpful"},
		{Role: "user", Content: "Hello"},
	}
	tools := []ToolDefinition{
		{Name: "get_time", Description: "Get current time"},
	}

	ctx := context.Background()
	_, err := client.Complete(ctx, messages, tools)

	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if len(capturedMessages) != 2 {
		t.Errorf("messages length = %d, want 2", len(capturedMessages))
	}
	if len(capturedTools) != 1 {
		t.Errorf("tools length = %d, want 1", len(capturedTools))
	}
}

// captureClient wraps a Client to capture arguments passed to Complete.
type captureClient struct {
	inner            Client
	capturedMessages *[]Message
	capturedTools    *[]ToolDefinition
}

func (c *captureClient) Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error) {
	*c.capturedMessages = messages
	*c.capturedTools = tools
	return c.inner.Complete(ctx, messages, tools)
}
