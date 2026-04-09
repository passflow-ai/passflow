# NATS EventBus Implementation

This package provides a NATS JetStream implementation of the EventBus interface for Passflow.

## Status

**Skeleton / Design Phase** — This is a stub implementation for evaluation purposes (ADR-002). The structure and interface are defined, but most methods are not yet implemented. This allows for:

1. Design validation
2. Benchmark planning
3. Migration strategy development
4. Interface compatibility testing

**Do not use in production** until the full implementation is completed and tested.

## Architecture

### Subject-based Routing

NATS JetStream uses hierarchical subjects for routing messages. Passflow maps event types to subjects:

```
Event Type              → Subject
-----------------       → --------------------
agent.started           → passflow.agents.started
agent.completed         → passflow.agents.completed
task.created            → passflow.tasks.created
incident.detected       → passflow.incidents.detected
trigger.agent           → passflow.triggers.agent
```

### Stream Configuration

The EventBus creates a single JetStream stream named `PASSFLOW` with:

- **Subjects**: `passflow.>` (all events)
- **Storage**: File-based (configurable to memory)
- **Replicas**: 3 (for HA in cluster)
- **Retention**: 24 hours (configurable)
- **Max Size**: 10GB (configurable)

### Consumer Groups

Consumers are organized into consumer groups (default: `passflow-agents`) for load balancing:

```
┌──────────────────────────────────────┐
│      NATS JetStream Stream           │
│         (PASSFLOW)                   │
│                                      │
│  Subject: passflow.agents.started    │
└──────────────┬───────────────────────┘
               │
         ┌─────┴─────┐
         │           │
    ┌────▼────┐ ┌────▼────┐
    │Consumer │ │Consumer │
    │   #1    │ │   #2    │
    └─────────┘ └─────────┘
       (Load balanced)
```

## Configuration

```go
config := nats.NATSConfig{
    URL:             "nats://nats-cluster:4222",
    StreamPrefix:    "passflow",
    ConsumerGroup:   "passflow-agents",
    ConsumerName:    "agent-consumer-1",
    MaxRetries:      3,
    SubjectStrategy: "domain", // or "flat"
    Replicas:        3,
    Storage:         "file",
    MaxAge:          86400,    // 24 hours
    MaxBytes:        10 * 1024 * 1024 * 1024, // 10GB
}

eventBus, err := nats.NewNATSEventBus(config, logger)
```

## Subject Strategies

### Domain Strategy (Recommended)

Events are routed to domain-specific subjects based on their type:

```go
SubjectStrategy: "domain"

// Routing:
agent.started → passflow.agents.started
task.created  → passflow.tasks.created
```

**Benefits:**
- Consumers can subscribe to specific domains (e.g., only agent events)
- Reduces noise in consumer processing
- Enables fine-grained access control
- Better monitoring per domain

### Flat Strategy

All events go to a single subject:

```go
SubjectStrategy: "flat"

// Routing:
agent.started → passflow.events
task.created  → passflow.events
```

**Use case:**
- Simple setup for small deployments
- Testing/development
- When all consumers need all events

## Implementation Roadmap

Based on ADR-002, the implementation will proceed in phases:

### Phase 1: Core Implementation (Week 1)
- [ ] Connection management
- [ ] Stream creation and configuration
- [ ] Publish/PublishAsync
- [ ] Subject routing logic
- [ ] Unit tests with embedded NATS server

### Phase 2: Subscription & Consumers (Week 2)
- [ ] Subscribe/SubscribeAll
- [ ] Durable consumer creation
- [ ] Consumer groups
- [ ] Ack/Nack handling
- [ ] Integration tests

### Phase 3: Advanced Features (Week 3)
- [ ] Exactly-once delivery (message ID deduplication)
- [ ] Dead letter queue
- [ ] Retry logic with backoff
- [ ] Metrics and observability
- [ ] Performance benchmarks

### Phase 4: Migration (Week 4)
- [ ] Dual-write capability (Redis + NATS)
- [ ] Consumer migration strategy
- [ ] Deployment guides
- [ ] Production readiness review

## Comparison: NATS vs Redis

| Feature | NATS JetStream | Redis Streams |
|---------|----------------|---------------|
| Clustering | Native RAFT | Requires Cluster/Sentinel |
| Subject routing | Built-in wildcards | Manual filtering |
| Exactly-once | Yes | No |
| Cross-region | Leafnodes | Custom |
| Memory footprint | ~50MB/pod | ~100MB/pod |
| Throughput | 10M msgs/sec | 5M msgs/sec |
| Latency p99 | <2ms | <5ms |

## Testing

Once implemented, tests will cover:

```bash
# Unit tests with embedded NATS
go test -v ./...

# Integration tests with NATS container
docker-compose -f test/docker-compose.yml up -d
go test -v -tags=integration ./...

# Benchmarks
go test -bench=. -benchmem ./...
```

## Deployment

### Kubernetes (OKE)

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nats
spec:
  replicas: 3
  serviceName: nats
  template:
    spec:
      containers:
      - name: nats
        image: nats:2.10-alpine
        args:
        - "--cluster_name=passflow"
        - "--jetstream"
        - "--store_dir=/data"
        ports:
        - containerPort: 4222
        - containerPort: 6222
        - containerPort: 8222
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

## References

- [NATS JetStream Documentation](https://docs.nats.io/nats-concepts/jetstream)
- [NATS Go Client](https://github.com/nats-io/nats.go)
- ADR-002: NATS JetStream Evaluation
- ADR-001: EDA & Agentic AI Alignment
