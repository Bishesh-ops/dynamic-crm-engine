package eventbus

import (
	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"log"
	"time"
)

type Event struct {
	SchemaID int
	Payload  map[string]any
}

type Bus struct {
	eventChan    chan Event
	db           *database.DB
	batchSize    int
	batchTimeout time.Duration
}

func New(db *database.DB, workerCount, batchSize int, batchTimeout time.Duration) *Bus {
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
	log.Printf("[Worker %d] Flushing batch of %d events to Postgres", workerID, len(batch))
}
