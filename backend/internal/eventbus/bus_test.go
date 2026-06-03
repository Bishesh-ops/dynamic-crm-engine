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
	DLQBatches      int // Tracks how many times the DLQ was hit
	DLQEvents       int // Tracks total events routed to DLQ
}

func (m *MockDB) SaveEntityBatch(ctx context.Context, entities []database.BatchEntity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BatchesReceived++
	m.TotalEvents += len(entities)
	return nil
}

func (m *MockDB) SaveToDLQ(ctx context.Context, events []database.DLQEvent, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DLQBatches++
	m.DLQEvents += len(events)
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

	err := bus.Publish(Event{SchemaID: 0, Payload: map[string]any{"msg": "wait for me"}})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	err = bus.Publish(Event{SchemaID: 0, Payload: map[string]any{"msg": "me too"}})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
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
