package queue

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryQueue is an in-memory implementation of Queue for testing.
type MemoryQueue struct {
	mu         sync.RWMutex
	queues     map[string]*memoryQueueData
	deadLetter map[string][]*QueueItem
}

type memoryQueueData struct {
	pending    []*QueueItem
	processing map[string]*QueueItem
	scheduled  []*QueueItem
	stats      *QueueStats
}

// NewMemoryQueue creates a new in-memory queue.
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		queues:     make(map[string]*memoryQueueData),
		deadLetter: make(map[string][]*QueueItem),
	}
}

func (m *MemoryQueue) getOrCreateQueue(name string) *memoryQueueData {
	if q, exists := m.queues[name]; exists {
		return q
	}
	q := &memoryQueueData{
		pending:    make([]*QueueItem, 0),
		processing: make(map[string]*QueueItem),
		scheduled:  make([]*QueueItem, 0),
		stats: &QueueStats{
			QueueName:   name,
			LastUpdated: time.Now(),
		},
	}
	m.queues[name] = q
	return q
}

// getQueue performs a read-only lookup; returns nil if the queue does not exist.
// Must be called while holding at least a read lock.
func (m *MemoryQueue) getQueue(name string) *memoryQueueData {
	return m.queues[name]
}

// Enqueue adds an item to the queue.
func (m *MemoryQueue) Enqueue(ctx context.Context, item *QueueItem) error {
	if err := item.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	q := m.getOrCreateQueue(item.QueueName)

	// Set timestamps
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	// Set default max attempts if not specified
	if item.MaxAttempts == 0 {
		item.MaxAttempts = 3
	}

	// Check if scheduled for future
	if item.IsScheduled() {
		q.scheduled = append(q.scheduled, item)
		q.stats.Scheduled++
	} else {
		q.pending = append(q.pending, item)
		q.stats.Pending++
	}

	q.stats.TotalEnqueued++
	q.stats.LastUpdated = now

	// Sort pending by priority (highest first)
	m.sortPending(q)

	return nil
}

func (m *MemoryQueue) sortPending(q *memoryQueueData) {
	sort.SliceStable(q.pending, func(i, j int) bool {
		return q.pending[i].Priority > q.pending[j].Priority
	})
}

// Dequeue retrieves and locks the next available item from the queue.
func (m *MemoryQueue) Dequeue(ctx context.Context, queueName string, leaseDuration time.Duration) (*QueueItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	q := m.getOrCreateQueue(queueName)

	// Move scheduled items that are due to pending
	m.moveScheduledToPending(q)

	if len(q.pending) == 0 {
		return nil, ErrQueueEmpty
	}

	// Get the highest priority item
	item := q.pending[0]
	q.pending = q.pending[1:]
	q.stats.Pending--

	// Create lease
	item.LeaseID = uuid.New().String()
	item.LeaseExpiry = time.Now().Add(leaseDuration)
	item.Attempts++
	item.UpdatedAt = time.Now()

	// Move to processing
	q.processing[item.ID] = item
	q.stats.Processing++

	return item, nil
}

func (m *MemoryQueue) moveScheduledToPending(q *memoryQueueData) {
	now := time.Now()
	stillScheduled := make([]*QueueItem, 0)

	for _, item := range q.scheduled {
		if !item.ScheduledAt.After(now) {
			q.pending = append(q.pending, item)
			q.stats.Pending++
			q.stats.Scheduled--
		} else {
			stillScheduled = append(stillScheduled, item)
		}
	}

	q.scheduled = stillScheduled
	m.sortPending(q)
}

// Ack acknowledges successful processing of an item.
func (m *MemoryQueue) Ack(ctx context.Context, queueName, itemID, leaseID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	q, exists := m.queues[queueName]
	if !exists {
		return ErrQueueNotFound
	}

	item, exists := q.processing[itemID]
	if !exists {
		return ErrItemNotFound
	}

	if item.LeaseID != leaseID {
		return ErrInvalidLease
	}

	delete(q.processing, itemID)
	q.stats.Processing--
	q.stats.TotalProcessed++
	q.stats.LastUpdated = time.Now()

	return nil
}

// Nack negatively acknowledges an item, optionally requeuing it.
func (m *MemoryQueue) Nack(ctx context.Context, queueName, itemID, leaseID string, requeue bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	q, exists := m.queues[queueName]
	if !exists {
		return ErrQueueNotFound
	}

	item, exists := q.processing[itemID]
	if !exists {
		return ErrItemNotFound
	}

	if item.LeaseID != leaseID {
		return ErrInvalidLease
	}

	delete(q.processing, itemID)
	q.stats.Processing--

	if requeue && item.CanRetry() {
		// Reset lease and requeue
		item.LeaseID = ""
		item.LeaseExpiry = time.Time{}
		item.UpdatedAt = time.Now()
		q.pending = append(q.pending, item)
		q.stats.Pending++
		m.sortPending(q)
	} else if requeue {
		// Max attempts reached - move to dead letter
		m.moveToDeadLetter(queueName, item)
		q.stats.TotalFailed++
	} else {
		// Don't requeue - just remove
		q.stats.TotalFailed++
	}

	q.stats.LastUpdated = time.Now()
	return nil
}

func (m *MemoryQueue) moveToDeadLetter(queueName string, item *QueueItem) {
	if _, exists := m.deadLetter[queueName]; !exists {
		m.deadLetter[queueName] = make([]*QueueItem, 0)
	}
	m.deadLetter[queueName] = append(m.deadLetter[queueName], item)

	if q, exists := m.queues[queueName]; exists {
		q.stats.DeadLetter++
	}
}

// ExtendLease extends the lease on a currently processing item.
func (m *MemoryQueue) ExtendLease(ctx context.Context, queueName, itemID, leaseID string, extension time.Duration) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	q, exists := m.queues[queueName]
	if !exists {
		return time.Time{}, ErrQueueNotFound
	}

	item, exists := q.processing[itemID]
	if !exists {
		return time.Time{}, ErrItemNotFound
	}

	if item.LeaseID != leaseID {
		return time.Time{}, ErrInvalidLease
	}

	item.LeaseExpiry = time.Now().Add(extension)
	item.UpdatedAt = time.Now()

	return item.LeaseExpiry, nil
}

// Stats returns statistics for a queue.
// If the queue has never been created, zero-value stats are returned without
// side effects (no write to m.queues under a read lock).
func (m *MemoryQueue) Stats(ctx context.Context, queueName string) (*QueueStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	q := m.getQueue(queueName)
	if q == nil {
		// Queue does not exist yet — return zero stats without creating it.
		return &QueueStats{QueueName: queueName}, nil
	}

	// Return a copy to avoid external mutation of internal state.
	stats := &QueueStats{
		QueueName:      q.stats.QueueName,
		Pending:        q.stats.Pending,
		Processing:     q.stats.Processing,
		DeadLetter:     q.stats.DeadLetter,
		Scheduled:      q.stats.Scheduled,
		TotalEnqueued:  q.stats.TotalEnqueued,
		TotalProcessed: q.stats.TotalProcessed,
		TotalFailed:    q.stats.TotalFailed,
		LastUpdated:    q.stats.LastUpdated,
	}

	return stats, nil
}

// GetDeadLetterItems returns items in the dead letter queue.
func (m *MemoryQueue) GetDeadLetterItems(ctx context.Context, queueName string, limit int) ([]*QueueItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items, exists := m.deadLetter[queueName]
	if !exists {
		return []*QueueItem{}, nil
	}

	if limit > len(items) {
		limit = len(items)
	}

	result := make([]*QueueItem, limit)
	copy(result, items[:limit])

	return result, nil
}

// RedriveDeadLetter moves an item from dead letter back to the main queue.
func (m *MemoryQueue) RedriveDeadLetter(ctx context.Context, queueName, itemID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	items, exists := m.deadLetter[queueName]
	if !exists {
		return ErrItemNotFound
	}

	var found *QueueItem
	newItems := make([]*QueueItem, 0, len(items)-1)
	for _, item := range items {
		if item.ID == itemID {
			found = item
		} else {
			newItems = append(newItems, item)
		}
	}

	if found == nil {
		return ErrItemNotFound
	}

	m.deadLetter[queueName] = newItems

	// Reset item and requeue
	found.Attempts = 0
	found.LeaseID = ""
	found.LeaseExpiry = time.Time{}
	found.LastError = ""
	found.UpdatedAt = time.Now()

	q := m.getOrCreateQueue(queueName)
	q.pending = append(q.pending, found)
	q.stats.Pending++
	q.stats.DeadLetter--
	m.sortPending(q)

	return nil
}

// Purge removes all items from a queue.
func (m *MemoryQueue) Purge(ctx context.Context, queueName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if q, exists := m.queues[queueName]; exists {
		q.pending = make([]*QueueItem, 0)
		q.processing = make(map[string]*QueueItem)
		q.scheduled = make([]*QueueItem, 0)
		q.stats.Pending = 0
		q.stats.Processing = 0
		q.stats.Scheduled = 0
		q.stats.LastUpdated = time.Now()
	}

	delete(m.deadLetter, queueName)

	return nil
}

// Close closes the queue connection (no-op for memory queue).
func (m *MemoryQueue) Close() error {
	return nil
}

// Ensure MemoryQueue implements Queue interface
var _ Queue = (*MemoryQueue)(nil)
