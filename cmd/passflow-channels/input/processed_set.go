package input

import "sync"

// processedSet is a bounded, thread-safe set of message IDs used to prevent
// duplicate event dispatch when the IMAP Store call fails and a message is not
// marked as seen on the server.
//
// When the set reaches its capacity, the oldest insertion is evicted to keep
// memory bounded. This is a simple FIFO eviction; a full LRU is unnecessary
// given typical email polling volumes.
type processedSet struct {
	mu       sync.Mutex
	ids      map[string]struct{}
	order    []string // insertion order for FIFO eviction
	capacity int
}

// newProcessedSet creates a processedSet with the given capacity.
func newProcessedSet(capacity int) *processedSet {
	return &processedSet{
		ids:      make(map[string]struct{}, capacity),
		order:    make([]string, 0, capacity),
		capacity: capacity,
	}
}

// add marks id as processed. Empty IDs are silently ignored.
// If the set is at capacity, the oldest entry is evicted before inserting.
func (p *processedSet) add(id string) {
	if id == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.ids[id]; exists {
		return // idempotent
	}

	// Evict oldest entry when at capacity.
	for len(p.ids) >= p.capacity {
		oldest := p.order[0]
		p.order = p.order[1:]
		delete(p.ids, oldest)
	}

	p.ids[id] = struct{}{}
	p.order = append(p.order, id)
}

// contains reports whether id has been processed.
func (p *processedSet) contains(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.ids[id]
	return ok
}

// size returns the current number of tracked IDs.
func (p *processedSet) size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.ids)
}
