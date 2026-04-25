package storage

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryStoreBasic(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	success, err := store.StartProcessing(ctx, "key1", 5*time.Minute)
	if err != nil {
		t.Fatalf("StartProcessing failed: %v", err)
	}
	if !success {
		t.Fatal("StartProcessing should return true for new key")
	}

	success, err = store.StartProcessing(ctx, "key1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Second StartProcessing failed: %v", err)
	}
	if success {
		t.Fatal("StartProcessing should return false for existing key")
	}
}

func TestMemoryStoreFinish(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	key := "key2"
	statusCode := 200
	responseBody := []byte(`{"status":"paid","amount":1000}`)

	_, _ = store.StartProcessing(ctx, key, 5*time.Minute)

	err := store.Finish(ctx, key, statusCode, responseBody, 24*time.Hour)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	cached, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cached == nil {
		t.Fatal("Expected non-nil cached response")
	}
	if cached.StatusCode != statusCode {
		t.Errorf("Expected status code %d, got %d", statusCode, cached.StatusCode)
	}
	if !cached.Completed {
		t.Fatal("Expected Completed to be true")
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	key := "key3"

	store.StartProcessing(ctx, key, 5*time.Minute)
	cached, _ := store.Get(ctx, key)
	if cached == nil {
		t.Fatal("Key should exist after StartProcessing")
	}

	err := store.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	cached, _ = store.Get(ctx, key)
	if cached != nil {
		t.Fatal("Key should not exist after deletion")
	}
}

func TestMemoryStoreRaceCondition(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	key := "concurrent-key"

	var mu sync.Mutex
	successCount := 0
	failureCount := 0

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			success, _ := store.StartProcessing(ctx, key, 5*time.Minute)
			mu.Lock()
			if success {
				successCount++
			} else {
				failureCount++
			}
			mu.Unlock()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	mu.Lock()
	defer mu.Unlock()

	if successCount != 1 {
		t.Errorf("Expected exactly 1 success, got %d", successCount)
	}
	if failureCount != 9 {
		t.Errorf("Expected exactly 9 failures, got %d", failureCount)
	}
}

func TestMemoryStoreProcessingState(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	key := "state-key"

	store.StartProcessing(ctx, key, 5*time.Minute)

	cached, _ := store.Get(ctx, key)
	if cached == nil {
		t.Fatal("Expected cached response during processing")
	}
	if cached.Completed {
		t.Fatal("Expected Completed to be false during processing")
	}

	store.Finish(ctx, key, 200, []byte("result"), 24*time.Hour)
	cached, _ = store.Get(ctx, key)
	if cached == nil {
		t.Fatal("Expected cached response after completion")
	}
	if !cached.Completed {
		t.Fatal("Expected Completed to be true after finishing")
	}
}
