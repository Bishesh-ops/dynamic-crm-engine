package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bisheshops/dynamic-crm-engine/internal/database"
)

type MockDB struct {
	mu              sync.Mutex
	BatchesReceived int
	TotalEvents     int
}

func (m *MockDB) SaveEntityBatch(ctx context.Context, entities []database.BatchEntity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BatchesReceived++
	m.TotalEvents += len(entities)
	return nil
}

func TestEventBus_BatchSizeTrigger(t *testing.T) {
	mockDB := &MockDB{}
	bus := New(mockDB, 1, 5, 10*time.Second)
	for i := range 5 {
		bus.Publish(Event{SchemaID: 1, Payload: map[string]any{"test": i}})
	}

	time.Sleep(50 * time.Millisecond)

	mockDB.mu.Lock()
	defer mockDB.mu.Unlock()

	if mockDB.BatchesReceived != 1 {
		t.Errorf("Expected 1 batch to be flushed, got %d", mockDB.BatchesReceived)
	}
	if mockDB.TotalEvents != 5 {
		t.Errorf("Expected 5 total events saved, got %d", mockDB.TotalEvents)
	}
}

func TestEventBus_TimeoutTrigger(t *testing.T) {
	mockDB := &MockDB{}

	bus := New(mockDB, 1, 100, 100*time.Millisecond)

	bus.Publish(Event{SchemaID: 1, Payload: map[string]any{"msg": "wait for me"}})
	bus.Publish(Event{SchemaID: 1, Payload: map[string]any{"msg": "me too"}})

	time.Sleep(150 * time.Millisecond)

	mockDB.mu.Lock()
	defer mockDB.mu.Unlock()

	if mockDB.BatchesReceived != 1 {
		t.Errorf("Expected ticker to flush 1 batch, got %d", mockDB.BatchesReceived)
	}
	if mockDB.TotalEvents != 2 {
		t.Errorf("Expected 2 events flushed by timeout, got %d", mockDB.TotalEvents)
	}
}
