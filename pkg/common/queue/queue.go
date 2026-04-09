package queue

import (
	"context"
	"errors"
	"time"
)

// Errors returned by the queue.
var (
	ErrQueueEmpty     = errors.New("queue is empty")
	ErrInvalidLease   = errors.New("invalid or expired lease")
	ErrItemNotFound   = errors.New("item not found")
	ErrInvalidItem    = errors.New("invalid queue item")
	ErrQueueNotFound  = errors.New("queue not found")
	ErrItemExists     = errors.New("item already exists")
	ErrMaxAttempts    = errors.New("maximum attempts exceeded")
)

// Priority levels for queue items.
type Priority int

const (
	PriorityLow      Priority = 0
	PriorityNormal   Priority = 1
	PriorityHigh     Priority = 2
	PriorityCritical Priority = 3
)

// String returns the string representation of the priority.
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// QueueItem represents an item in the queue.
type QueueItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	QueueName   string                 `json:"queueName"`
	Priority    Priority               `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]string      `json:"metadata"`
	ScheduledAt time.Time              `json:"scheduledAt,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"maxAttempts"`
	LastError   string                 `json:"lastError,omitempty"`
	LeaseID     string                 `json:"leaseId,omitempty"`
	LeaseExpiry time.Time              `json:"leaseExpiry,omitempty"`
	WorkspaceID string                 `json:"workspaceId,omitempty"`
}

// Validate checks if the queue item is valid.
func (q *QueueItem) Validate() error {
	if q.ID == "" {
		return errors.New("queue item ID is required")
	}
	if q.Type == "" {
		return errors.New("queue item type is required")
	}
	if q.QueueName == "" {
		return errors.New("queue name is required")
	}
	return nil
}

// IsScheduled returns true if the item is scheduled for a future time.
func (q *QueueItem) IsScheduled() bool {
	return !q.ScheduledAt.IsZero() && q.ScheduledAt.After(time.Now())
}

// IsLeaseValid returns true if the lease is still valid.
func (q *QueueItem) IsLeaseValid() bool {
	return q.LeaseID != "" && q.LeaseExpiry.After(time.Now())
}

// CanRetry returns true if the item can be retried.
func (q *QueueItem) CanRetry() bool {
	if q.MaxAttempts == 0 {
		return true // No limit
	}
	return q.Attempts < q.MaxAttempts
}

// QueueStats contains statistics for a queue.
type QueueStats struct {
	QueueName   string    `json:"queueName"`
	Pending     int64     `json:"pending"`
	Processing  int64     `json:"processing"`
	DeadLetter  int64     `json:"deadLetter"`
	Scheduled   int64     `json:"scheduled"`
	TotalEnqueued int64   `json:"totalEnqueued"`
	TotalProcessed int64  `json:"totalProcessed"`
	TotalFailed   int64   `json:"totalFailed"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// QueueConfig holds configuration for a queue.
type QueueConfig struct {
	Name             string        `json:"name"`
	MaxRetries       int           `json:"maxRetries"`
	RetryDelay       time.Duration `json:"retryDelay"`
	DefaultLease     time.Duration `json:"defaultLease"`
	MaxLease         time.Duration `json:"maxLease"`
	VisibilityTimeout time.Duration `json:"visibilityTimeout"`
}

// DefaultQueueConfig returns a default queue configuration.
func DefaultQueueConfig(name string) QueueConfig {
	return QueueConfig{
		Name:             name,
		MaxRetries:       3,
		RetryDelay:       5 * time.Second,
		DefaultLease:     30 * time.Second,
		MaxLease:         5 * time.Minute,
		VisibilityTimeout: 30 * time.Second,
	}
}

// Queue defines the interface for a queue manager.
type Queue interface {
	// Enqueue adds an item to the queue.
	Enqueue(ctx context.Context, item *QueueItem) error

	// Dequeue retrieves and locks the next available item from the queue.
	// Returns ErrQueueEmpty if no items are available.
	Dequeue(ctx context.Context, queueName string, leaseDuration time.Duration) (*QueueItem, error)

	// Ack acknowledges successful processing of an item.
	Ack(ctx context.Context, queueName, itemID, leaseID string) error

	// Nack negatively acknowledges an item, optionally requeuing it.
	Nack(ctx context.Context, queueName, itemID, leaseID string, requeue bool) error

	// ExtendLease extends the lease on a currently processing item.
	ExtendLease(ctx context.Context, queueName, itemID, leaseID string, extension time.Duration) (time.Time, error)

	// Stats returns statistics for a queue.
	Stats(ctx context.Context, queueName string) (*QueueStats, error)

	// GetDeadLetterItems returns items in the dead letter queue.
	GetDeadLetterItems(ctx context.Context, queueName string, limit int) ([]*QueueItem, error)

	// RedriveDeadLetter moves an item from dead letter back to the main queue.
	RedriveDeadLetter(ctx context.Context, queueName, itemID string) error

	// Purge removes all items from a queue.
	Purge(ctx context.Context, queueName string) error

	// Close closes the queue connection.
	Close() error
}

// BatchQueue extends Queue with batch operations.
type BatchQueue interface {
	Queue

	// EnqueueBatch adds multiple items to the queue.
	EnqueueBatch(ctx context.Context, items []*QueueItem) error

	// DequeueBatch retrieves multiple items from the queue.
	DequeueBatch(ctx context.Context, queueName string, count int, leaseDuration time.Duration) ([]*QueueItem, error)

	// AckBatch acknowledges multiple items.
	AckBatch(ctx context.Context, queueName string, items []struct{ ID, LeaseID string }) error
}

// DelayedQueue extends Queue with delayed message support.
type DelayedQueue interface {
	Queue

	// EnqueueDelayed adds an item to be processed after a delay.
	EnqueueDelayed(ctx context.Context, item *QueueItem, delay time.Duration) error

	// ScheduleAt schedules an item to be processed at a specific time.
	ScheduleAt(ctx context.Context, item *QueueItem, processAt time.Time) error
}
