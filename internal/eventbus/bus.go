package eventbus

import (
	"context"
	"errors"
	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"github.com/bisheshops/dynamic-crm-engine/internal/workflow"
	"log"
	"time"
)

var ErrQueueFull = errors.New("event bus queue is full, backpressure applied")

type BatchSaver interface {
	SaveEntityBatch(ctx context.Context, entities []database.BatchEntity) error
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
	workflows    []workflow.Workflow
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

func (b *Bus) Publish(ev Event) error {
	select {
	case b.eventChan <- ev:
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := b.db.SaveEntityBatch(ctx, dbBatch)
	if err != nil {
		log.Printf("[Worker %d] FATAL: Failed to flush batch of %d events: %v", workerID, len(batch), err)
		return
	}
	log.Printf("[Worker %d] Flushing batch of %d events to Postgres", workerID, len(batch))
}

func (b *Bus) applyWorkflows(ev *Event) {
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
