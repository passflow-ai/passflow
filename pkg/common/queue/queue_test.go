package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryQueue(t *testing.T) {
	q := NewMemoryQueue()
	assert.NotNil(t, q)
}

func TestQueue_Enqueue_SingleItem(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Payload:   map[string]interface{}{"taskId": "123"},
		Priority:  PriorityNormal,
		QueueName: "default",
	}

	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Pending)
}

func TestQueue_Enqueue_WithPriority(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	// Enqueue items with different priorities
	items := []*QueueItem{
		{ID: "low-1", Type: "task", Priority: PriorityLow, QueueName: "default"},
		{ID: "high-1", Type: "task", Priority: PriorityHigh, QueueName: "default"},
		{ID: "normal-1", Type: "task", Priority: PriorityNormal, QueueName: "default"},
		{ID: "critical-1", Type: "task", Priority: PriorityCritical, QueueName: "default"},
	}

	for _, item := range items {
		err := q.Enqueue(ctx, item)
		require.NoError(t, err)
	}

	// Dequeue should return in priority order (highest first)
	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "critical-1", dequeued.ID)

	dequeued, err = q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "high-1", dequeued.ID)

	dequeued, err = q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "normal-1", dequeued.ID)

	dequeued, err = q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "low-1", dequeued.ID)
}

func TestQueue_Dequeue_EmptyQueue(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item, err := q.Dequeue(ctx, "default", 100*time.Millisecond)
	assert.Nil(t, item)
	assert.Equal(t, ErrQueueEmpty, err)
}

func TestQueue_Dequeue_WithLease(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Priority:  PriorityNormal,
		QueueName: "default",
	}
	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	// Dequeue with lease
	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "task-1", dequeued.ID)
	assert.NotEmpty(t, dequeued.LeaseID)
	assert.False(t, dequeued.LeaseExpiry.IsZero())

	// Queue should show item as processing
	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(1), stats.Processing)
}

func TestQueue_Ack_Success(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Priority:  PriorityNormal,
		QueueName: "default",
	}
	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)

	// Acknowledge completion
	err = q.Ack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID)
	require.NoError(t, err)

	// Queue should be empty
	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(0), stats.Processing)
}

func TestQueue_Ack_InvalidLease(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Priority:  PriorityNormal,
		QueueName: "default",
	}
	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)

	// Try to ack with wrong lease ID
	err = q.Ack(ctx, dequeued.QueueName, dequeued.ID, "wrong-lease-id")
	assert.Equal(t, ErrInvalidLease, err)
}

func TestQueue_Nack_RequeuesItem(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Priority:  PriorityNormal,
		QueueName: "default",
	}
	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)

	// Negative acknowledge - should requeue
	err = q.Nack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID, true)
	require.NoError(t, err)

	// Item should be back in pending
	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Pending)
	assert.Equal(t, int64(0), stats.Processing)

	// Should be able to dequeue again
	dequeued2, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "task-1", dequeued2.ID)
	assert.Equal(t, 2, dequeued2.Attempts) // First dequeue=1, second dequeue=2
}

func TestQueue_Nack_NoRequeue(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Priority:  PriorityNormal,
		QueueName: "default",
	}
	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)

	// Negative acknowledge without requeue
	err = q.Nack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID, false)
	require.NoError(t, err)

	// Queue should be empty
	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(0), stats.Processing)
}

func TestQueue_ExtendLease(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Priority:  PriorityNormal,
		QueueName: "default",
	}
	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	dequeued, err := q.Dequeue(ctx, "default", 1*time.Second)
	require.NoError(t, err)

	originalExpiry := dequeued.LeaseExpiry

	// Extend lease
	newExpiry, err := q.ExtendLease(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID, 30*time.Second)
	require.NoError(t, err)
	assert.True(t, newExpiry.After(originalExpiry))
}

func TestQueue_MultipleQueues(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	// Enqueue to different queues
	item1 := &QueueItem{ID: "task-1", Type: "task", Priority: PriorityNormal, QueueName: "queue-a"}
	item2 := &QueueItem{ID: "task-2", Type: "task", Priority: PriorityNormal, QueueName: "queue-b"}

	err := q.Enqueue(ctx, item1)
	require.NoError(t, err)
	err = q.Enqueue(ctx, item2)
	require.NoError(t, err)

	// Stats should be separate
	statsA, err := q.Stats(ctx, "queue-a")
	require.NoError(t, err)
	assert.Equal(t, int64(1), statsA.Pending)

	statsB, err := q.Stats(ctx, "queue-b")
	require.NoError(t, err)
	assert.Equal(t, int64(1), statsB.Pending)

	// Dequeue from specific queue
	dequeued, err := q.Dequeue(ctx, "queue-a", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "task-1", dequeued.ID)
}

func TestQueue_ScheduledItem(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	// Schedule item for future
	item := &QueueItem{
		ID:          "task-1",
		Type:        "task_execution",
		Priority:    PriorityNormal,
		QueueName:   "default",
		ScheduledAt: time.Now().Add(100 * time.Millisecond),
	}

	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	// Should not be available immediately
	dequeued, err := q.Dequeue(ctx, "default", 10*time.Millisecond)
	assert.Nil(t, dequeued)
	assert.Equal(t, ErrQueueEmpty, err)

	// Wait for scheduled time
	time.Sleep(150 * time.Millisecond)

	// Now should be available
	dequeued, err = q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "task-1", dequeued.ID)
}

func TestQueue_MaxAttempts(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:          "task-1",
		Type:        "task_execution",
		Priority:    PriorityNormal,
		QueueName:   "default",
		MaxAttempts: 2,
	}

	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	// First attempt
	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	err = q.Nack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID, true)
	require.NoError(t, err)

	// Second attempt
	dequeued, err = q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	err = q.Nack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID, true)
	require.NoError(t, err)

	// Should not requeue (max attempts reached)
	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(1), stats.DeadLetter)
}

func TestQueue_DeadLetterQueue(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:          "task-1",
		Type:        "task_execution",
		Priority:    PriorityNormal,
		QueueName:   "default",
		MaxAttempts: 1,
	}

	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	err = q.Nack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID, true)
	require.NoError(t, err)

	// Get dead letter items
	deadItems, err := q.GetDeadLetterItems(ctx, "default", 10)
	require.NoError(t, err)
	require.Len(t, deadItems, 1)
	assert.Equal(t, "task-1", deadItems[0].ID)
}

func TestQueue_RedriveDeadLetter(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	item := &QueueItem{
		ID:          "task-1",
		Type:        "task_execution",
		Priority:    PriorityNormal,
		QueueName:   "default",
		MaxAttempts: 1,
	}

	err := q.Enqueue(ctx, item)
	require.NoError(t, err)

	dequeued, err := q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	err = q.Nack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID, true)
	require.NoError(t, err)

	// Redrive dead letter item
	err = q.RedriveDeadLetter(ctx, "default", "task-1")
	require.NoError(t, err)

	// Should be back in pending queue
	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Pending)
	assert.Equal(t, int64(0), stats.DeadLetter)

	// Should be able to dequeue
	dequeued, err = q.Dequeue(ctx, "default", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "task-1", dequeued.ID)
	assert.Equal(t, 1, dequeued.Attempts) // Redrive resets to 0, then dequeue increments to 1
}

func TestQueue_ConcurrentAccess(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	// Enqueue many items
	for i := 0; i < 100; i++ {
		item := &QueueItem{
			ID:        fmt.Sprintf("task-%d", i),
			Type:      "task_execution",
			Priority:  PriorityNormal,
			QueueName: "default",
		}
		err := q.Enqueue(ctx, item)
		require.NoError(t, err)
	}

	// Concurrent dequeue
	var wg sync.WaitGroup
	dequeuedIDs := make(chan string, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				item, err := q.Dequeue(ctx, "default", 100*time.Millisecond)
				if err == ErrQueueEmpty {
					return
				}
				if err == nil && item != nil {
					dequeuedIDs <- item.ID
					// Ack the item
					q.Ack(ctx, item.QueueName, item.ID, item.LeaseID)
				}
			}
		}()
	}

	wg.Wait()
	close(dequeuedIDs)

	// Collect dequeued IDs
	ids := make(map[string]bool)
	for id := range dequeuedIDs {
		ids[id] = true
	}

	// All items should have been dequeued exactly once
	assert.Equal(t, 100, len(ids))
}

func TestQueueItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    *QueueItem
		wantErr bool
	}{
		{
			name: "valid item",
			item: &QueueItem{
				ID:        "task-1",
				Type:      "task_execution",
				QueueName: "default",
				Priority:  PriorityNormal,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			item: &QueueItem{
				Type:      "task_execution",
				QueueName: "default",
				Priority:  PriorityNormal,
			},
			wantErr: true,
		},
		{
			name: "missing type",
			item: &QueueItem{
				ID:        "task-1",
				QueueName: "default",
				Priority:  PriorityNormal,
			},
			wantErr: true,
		},
		{
			name: "missing queue name",
			item: &QueueItem{
				ID:       "task-1",
				Type:     "task_execution",
				Priority: PriorityNormal,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStats_UnknownQueueReturnsZero verifies that Stats for a queue that has
// never received an item returns zero-value stats without creating the queue
// entry. The race detector catches the former bug (RLock + write via
// getOrCreateQueue).
func TestStats_UnknownQueueReturnsZero(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	stats, err := q.Stats(ctx, "never-enqueued-queue")
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(0), stats.Processing)
	assert.Equal(t, int64(0), stats.DeadLetter)
	assert.Equal(t, int64(0), stats.Scheduled)
	assert.Equal(t, int64(0), stats.TotalEnqueued)
	assert.Equal(t, "never-enqueued-queue", stats.QueueName)
}

// TestStats_ConcurrentReadWrite verifies that Stats is safe to call concurrently
// with Enqueue (the race detector will catch any unsynchronised accesses).
func TestStats_ConcurrentReadWrite(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			_ = q.Enqueue(ctx, &QueueItem{
				ID:        fmt.Sprintf("item-%d", n),
				Type:      "task",
				QueueName: "race-queue",
				Priority:  PriorityNormal,
			})
		}(i)
		go func() {
			defer wg.Done()
			_, _ = q.Stats(ctx, "race-queue")
		}()
	}
	wg.Wait()
}

func TestQueueStats(t *testing.T) {
	q := NewMemoryQueue()
	ctx := context.Background()

	// Initial stats
	stats, err := q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(0), stats.Processing)
	assert.Equal(t, int64(0), stats.DeadLetter)

	// Enqueue
	item := &QueueItem{
		ID:        "task-1",
		Type:      "task_execution",
		Priority:  PriorityNormal,
		QueueName: "default",
	}
	err = q.Enqueue(ctx, item)
	require.NoError(t, err)

	stats, err = q.Stats(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Pending)

	// Dequeue
	dequeued, _ := q.Dequeue(ctx, "default", 30*time.Second)
	stats, _ = q.Stats(ctx, "default")
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(1), stats.Processing)

	// Ack
	q.Ack(ctx, dequeued.QueueName, dequeued.ID, dequeued.LeaseID)
	stats, _ = q.Stats(ctx, "default")
	assert.Equal(t, int64(0), stats.Processing)
}
