package eventbus

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"github.com/bisheshops/dynamic-crm-engine/internal/metrics"
	"github.com/bisheshops/dynamic-crm-engine/internal/workflow"
	"github.com/prometheus/client_golang/prometheus"
)

var ErrQueueFull = errors.New("event bus queue is full, backpressure applied")

type BatchSaver interface {
	SaveEntityBatch(ctx context.Context, entities []database.BatchEntity) error
	SaveToDLQ(ctx context.Context, events []database.DLQEvent, reason string) error
}

type Event struct {
	SchemaID   int
	SchemaName string
	Payload    map[string]any
}

type Bus struct {
	eventChan    chan Event
	db           BatchSaver
	batchSize    int
	batchTimeout time.Duration

	workflows []workflow.Workflow
	mu        sync.RWMutex
}

func New(db BatchSaver, workerCount, batchSize int, batchTimeout time.Duration, workflows []workflow.Workflow) *Bus {
	b := &Bus{
		eventChan:    make(chan Event, batchSize*workerCount),
		db:           db,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		workflows:    workflows,
	}

	for i := range workerCount {
		go b.worker(i)
	}
	log.Printf("Event bus initialized with %d workers. Batch size: %d", workerCount, batchSize)
	return b
}

func (b *Bus) AddWorkflow(wf workflow.Workflow) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.workflows = append(b.workflows, wf)
	log.Printf("Hot-loaded new workflow '%s'. Total active workflows: %d", wf.Name, len(b.workflows))
}

func (b *Bus) Publish(ev Event) error {
	select {
	case b.eventChan <- ev:
		metrics.QueueLength.Inc()
		return nil
	default:
		return ErrQueueFull
	}
}

func (b *Bus) worker(id int) {
	batch := make([]Event, 0, b.batchSize)

	ticker := time.NewTicker(b.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case ev := <-b.eventChan:
			metrics.QueueLength.Dec()
			b.applyWorkflows(&ev)
			batch = append(batch, ev)

			if len(batch) >= b.batchSize {
				b.flushBatch(id, batch)
				batch = make([]Event, 0, b.batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				b.flushBatch(id, batch)
				batch = make([]Event, 0, b.batchSize)
			}
		}
	}
}

func (b *Bus) flushBatch(workerID int, batch []Event) {
	if len(batch) == 0 {
		return
	}

	dbBatch := make([]database.BatchEntity, len(batch))
	for i, ev := range batch {
		dbBatch[i] = database.BatchEntity{
			SchemaID: ev.SchemaID,
			Data:     ev.Payload,
		}
	}

	timer := prometheus.NewTimer(metrics.BatchFlushDuration)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := b.db.SaveEntityBatch(ctx, dbBatch)
	timer.ObserveDuration()

	if err != nil {
		metrics.EventsProcessed.WithLabelValues("error").Add(float64(len(batch)))
		log.Printf("[Worker %d] ERROR: Failed to flush batch of %d events: %v. Routing to DLQ.", workerID, len(batch), err)

		dlqCtx, dlqCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dlqCancel()
		dlqEvents := make([]database.DLQEvent, len(batch))
		for i, ev := range batch {
			dlqEvents[i] = database.DLQEvent{
				SchemaID:   ev.SchemaID,
				SchemaName: ev.SchemaName,
				Payload:    ev.Payload,
			}
		}
		if dlqErr := b.db.SaveToDLQ(dlqCtx, dlqEvents, err.Error()); dlqErr != nil {
			log.Printf("[Worker %d] FATAL: DLQ write failed! %d events permanently lost: %v", workerID, len(batch), dlqErr)
		}
		return
	}

	metrics.EventsProcessed.WithLabelValues("success").Add(float64(len(batch)))
	log.Printf("[Worker %d] Flushing batch of %d events to Postgres", workerID, len(batch))
}

func (b *Bus) applyWorkflows(ev *Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, wf := range b.workflows {
		if !wf.IsActive || wf.TargetSchema != ev.SchemaName {
			continue
		}
		if wf.Condition.Evaluate(ev.Payload) {
			err := workflow.ApplyActions(ev.Payload, wf.Actions)
			if err != nil {
				log.Printf("Workflow '%s' failed on event for schema '%s': %v", wf.Name, ev.SchemaName, err)
			}
		}
	}
}
