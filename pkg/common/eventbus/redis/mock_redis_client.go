package redis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

// MockRedisClient implements a mock for testing without real Redis.
type MockRedisClient struct {
	mu            sync.RWMutex
	streams       map[string][]redis.XMessage
	sets          map[string]map[string]float64
	hashes        map[string]map[string]string
	subscriptions map[string][]chan *redis.Message
	messageID     int
	closed        bool
	shouldFail    bool
	failureError  error
}

// NewMockRedisClient creates a new mock Redis client.
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		streams:       make(map[string][]redis.XMessage),
		sets:          make(map[string]map[string]float64),
		hashes:        make(map[string]map[string]string),
		subscriptions: make(map[string][]chan *redis.Message),
		messageID:     0,
	}
}

// XAdd mocks the Redis XADD command.
func (m *MockRedisClient) XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewStringCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	m.messageID++
	msgID := generateMockMessageID(m.messageID)

	values, ok := a.Values.(map[string]interface{})
	if !ok {
		values = make(map[string]interface{})
	}
	msg := redis.XMessage{
		ID:     msgID,
		Values: values,
	}

	if _, exists := m.streams[a.Stream]; !exists {
		m.streams[a.Stream] = []redis.XMessage{}
	}
	m.streams[a.Stream] = append(m.streams[a.Stream], msg)

	cmd.SetVal(msgID)
	return cmd
}

// XReadGroup mocks the Redis XREADGROUP command.
func (m *MockRedisClient) XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewXStreamSliceCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	var results []redis.XStream
	for _, stream := range a.Streams {
		if stream == ">" {
			continue
		}
		if messages, exists := m.streams[stream]; exists && len(messages) > 0 {
			count := a.Count
			if count == 0 || count > int64(len(messages)) {
				count = int64(len(messages))
			}
			results = append(results, redis.XStream{
				Stream:   stream,
				Messages: messages[:count],
			})
		}
	}

	cmd.SetVal(results)
	return cmd
}

// XAck mocks the Redis XACK command.
func (m *MockRedisClient) XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	acked := int64(0)
	if messages, exists := m.streams[stream]; exists {
		var remaining []redis.XMessage
		for _, msg := range messages {
			shouldRemove := false
			for _, id := range ids {
				if msg.ID == id {
					shouldRemove = true
					acked++
					break
				}
			}
			if !shouldRemove {
				remaining = append(remaining, msg)
			}
		}
		m.streams[stream] = remaining
	}

	cmd.SetVal(acked)
	return cmd
}

// XGroupCreateMkStream mocks the Redis XGROUP CREATE command.
func (m *MockRedisClient) XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}
	cmd.SetVal("OK")
	return cmd
}

// XLen mocks the Redis XLEN command.
func (m *MockRedisClient) XLen(ctx context.Context, stream string) *redis.IntCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if messages, exists := m.streams[stream]; exists {
		cmd.SetVal(int64(len(messages)))
		return cmd
	}

	cmd.SetVal(0)
	return cmd
}

// HSet mocks the Redis HSET command.
func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if _, exists := m.hashes[key]; !exists {
		m.hashes[key] = make(map[string]string)
	}

	for i := 0; i < len(values); i += 2 {
		field := values[i].(string)
		value := values[i+1].(string)
		m.hashes[key][field] = value
	}

	cmd.SetVal(int64(len(values) / 2))
	return cmd
}

// HGet mocks the Redis HGET command.
func (m *MockRedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd := redis.NewStringCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if hash, exists := m.hashes[key]; exists {
		if value, fieldExists := hash[field]; fieldExists {
			cmd.SetVal(value)
			return cmd
		}
	}

	cmd.SetErr(redis.Nil)
	return cmd
}

// HGetAll mocks the Redis HGETALL command.
func (m *MockRedisClient) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd := redis.NewStringStringMapCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if hash, exists := m.hashes[key]; exists {
		cmd.SetVal(hash)
		return cmd
	}

	cmd.SetVal(make(map[string]string))
	return cmd
}

// ZAdd mocks the Redis ZADD command.
func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if _, exists := m.sets[key]; !exists {
		m.sets[key] = make(map[string]float64)
	}

	added := int64(0)
	for _, member := range members {
		memberStr := member.Member.(string)
		if _, exists := m.sets[key][memberStr]; !exists {
			added++
		}
		m.sets[key][memberStr] = member.Score
	}

	cmd.SetVal(added)
	return cmd
}

// ZRangeByScore mocks the Redis ZRANGEBYSCORE command.
func (m *MockRedisClient) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd := redis.NewStringSliceCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	var results []string
	if set, exists := m.sets[key]; exists {
		for member := range set {
			results = append(results, member)
		}
	}

	cmd.SetVal(results)
	return cmd
}

// ZRem mocks the Redis ZREM command.
func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	removed := int64(0)
	if set, exists := m.sets[key]; exists {
		for _, member := range members {
			memberStr := member.(string)
			if _, memberExists := set[memberStr]; memberExists {
				delete(set, memberStr)
				removed++
			}
		}
	}

	cmd.SetVal(removed)
	return cmd
}

// Publish mocks the Redis PUBLISH command.
func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if subs, exists := m.subscriptions[channel]; exists {
		for _, sub := range subs {
			go func(ch chan *redis.Message) {
				ch <- &redis.Message{
					Channel: channel,
					Payload: message.(string),
				}
			}(sub)
		}
	}

	cmd.SetVal(int64(len(m.subscriptions[channel])))
	return cmd
}

// Close mocks closing the Redis client.
func (m *MockRedisClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// SetShouldFail configures the mock to fail operations.
func (m *MockRedisClient) SetShouldFail(fail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
	m.failureError = err
}

// GetStreamMessages returns messages from a stream.
func (m *MockRedisClient) GetStreamMessages(stream string) []redis.XMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.streams[stream]
}

// GetHashValue returns a value from a hash.
func (m *MockRedisClient) GetHashValue(key, field string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if hash, exists := m.hashes[key]; exists {
		if value, fieldExists := hash[field]; fieldExists {
			return value, true
		}
	}
	return "", false
}

// XRevRange mocks the Redis XREVRANGE command.
func (m *MockRedisClient) XRevRange(ctx context.Context, stream, start, stop string) *redis.XMessageSliceCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd := redis.NewXMessageSliceCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if messages, exists := m.streams[stream]; exists {
		// Return messages in reverse order
		reversed := make([]redis.XMessage, len(messages))
		for i, msg := range messages {
			reversed[len(messages)-1-i] = msg
		}
		cmd.SetVal(reversed)
		return cmd
	}

	cmd.SetVal([]redis.XMessage{})
	return cmd
}

// XDel mocks the Redis XDEL command.
func (m *MockRedisClient) XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	deleted := int64(0)
	if messages, exists := m.streams[stream]; exists {
		var remaining []redis.XMessage
		for _, msg := range messages {
			shouldDelete := false
			for _, id := range ids {
				if msg.ID == id {
					shouldDelete = true
					deleted++
					break
				}
			}
			if !shouldDelete {
				remaining = append(remaining, msg)
			}
		}
		m.streams[stream] = remaining
	}

	cmd.SetVal(deleted)
	return cmd
}

// Del mocks the Redis DEL command.
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	deleted := int64(0)
	for _, key := range keys {
		if _, exists := m.hashes[key]; exists {
			delete(m.hashes, key)
			deleted++
		}
	}

	cmd.SetVal(deleted)
	return cmd
}

// XInfoStream mocks the Redis XINFO STREAM command.
func (m *MockRedisClient) XInfoStream(ctx context.Context, stream string) *redis.XInfoStreamCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd := redis.NewXInfoStreamCmd(ctx, stream)
	if m.shouldFail {
		cmd.SetErr(m.failureError)
		return cmd
	}

	if messages, exists := m.streams[stream]; exists {
		info := &redis.XInfoStream{
			Length: int64(len(messages)),
		}
		if len(messages) > 0 {
			info.FirstEntry = messages[0]
			info.LastEntry = messages[len(messages)-1]
		}
		cmd.SetVal(info)
		return cmd
	}

	cmd.SetVal(&redis.XInfoStream{Length: 0})
	return cmd
}

func generateMockMessageID(seq int) string {
	return "1234567890-" + string(rune('0'+seq%10))
}
