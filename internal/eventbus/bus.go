package eventbus

import (
	"context"
	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"log"
	"time"
)

type BatchSaver interface {
	SaveEntityBatch(ctx context.Context, entities []database.BatchEntity) error
}
type Event struct {
	SchemaID int
	Payload  map[string]any
}

type Bus struct {
	eventChan    chan Event
	db           BatchSaver
	batchSize    int
	batchTimeout time.Duration
}

func New(db BatchSaver, workerCount, batchSize int, batchTimeout time.Duration) *Bus {
	b := &Bus{
		eventChan:    make(chan Event, batchSize*workerCount),
		db:           db,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
	}

	for i := range workerCount {
		go b.worker(i)
	}
	log.Printf("Event bus initialized with %d workers. Batch size: %d", workerCount, batchSize)
	return b
}

func (b *Bus) Publish(ev Event) {
	b.eventChan <- ev
}

func (b *Bus) worker(id int) {
	batch := make([]Event, 0, b.batchSize)

	ticker := time.NewTicker(b.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case ev := <-b.eventChan:
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
