package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"github.com/bisheshops/dynamic-crm-engine/internal/workflow"
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
	bus := New(mockDB, 1, 5, 10*time.Second, []workflow.Workflow{})

	for i := range 6 {
		bus.Publish(Event{SchemaID: 1, Payload: map[string]any{"test": i}})
	}

	deadline := time.Now().Add(2 * time.Second)
	success := false

	for time.Now().Before(deadline) {
		mockDB.mu.Lock()
		batches := mockDB.BatchesReceived
		events := mockDB.TotalEvents
		mockDB.mu.Unlock()

		if batches == 1 && events == 5 {
			success = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !success {
		mockDB.mu.Lock()
		defer mockDB.mu.Unlock()
		t.Fatalf("Timeout waiting for worker batch. Got %d batches, %d events", mockDB.BatchesReceived, mockDB.TotalEvents)
	}
}

func TestEventBus_TimeoutTrigger(t *testing.T) {
	mockDB := &MockDB{}
	bus := New(mockDB, 1, 100, 100*time.Millisecond, []workflow.Workflow{})

	bus.Publish(Event{SchemaID: 1, Payload: map[string]any{"msg": "wait for me"}})
	bus.Publish(Event{SchemaID: 1, Payload: map[string]any{"msg": "me too"}})

	deadline := time.Now().Add(2 * time.Second)
	success := false

	for time.Now().Before(deadline) {
		mockDB.mu.Lock()
		batches := mockDB.BatchesReceived
		events := mockDB.TotalEvents
		mockDB.mu.Unlock()

		if batches == 1 && events == 2 {
			success = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !success {
		mockDB.mu.Lock()
		defer mockDB.mu.Unlock()
		t.Fatalf("Timeout waiting for ticker batch. Got %d batches, %d events", mockDB.BatchesReceived, mockDB.TotalEvents)
	}
}
