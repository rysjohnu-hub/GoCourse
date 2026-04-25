package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
}

type IdempotencyStore interface {
	Get(ctx context.Context, key string) (*CachedResponse, error)
	StartProcessing(ctx context.Context, key string, processingTTL time.Duration) (bool, error)
	Finish(ctx context.Context, key string, status int, body []byte, resultTTL time.Duration) error
	Delete(ctx context.Context, key string) error
}

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

func (rs *RedisStore) Get(ctx context.Context, key string) (*CachedResponse, error) {
	val, err := rs.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var resp CachedResponse
	if err := json.Unmarshal([]byte(val), &resp); err == nil {
		resp.Completed = true
		return &resp, nil
	}

	if val == "processing" {
		return &CachedResponse{Completed: false}, nil
	}

	return nil, fmt.Errorf("invalid cached value format")
}

func (rs *RedisStore) StartProcessing(ctx context.Context, key string, processingTTL time.Duration) (bool, error) {
	success, err := rs.client.SetNX(ctx, key, "processing", processingTTL).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx error: %w", err)
	}

	return success, nil
}

func (rs *RedisStore) Finish(ctx context.Context, key string, status int, body []byte, resultTTL time.Duration) error {
	resp := CachedResponse{
		StatusCode: status,
		Body:       body,
		Completed:  true,
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	if err := rs.client.Set(ctx, key, respJSON, resultTTL).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (rs *RedisStore) Delete(ctx context.Context, key string) error {
	if err := rs.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

func (rs *RedisStore) Close() error {
	return rs.client.Close()
}

type MemoryStore struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]string),
	}
}

func (ms *MemoryStore) Get(ctx context.Context, key string) (*CachedResponse, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, exists := ms.data[key]
	if !exists {
		return nil, nil
	}

	var resp CachedResponse
	if err := json.Unmarshal([]byte(val), &resp); err == nil {
		resp.Completed = true
		return &resp, nil
	}

	if val == "processing" {
		return &CachedResponse{Completed: false}, nil
	}

	return nil, fmt.Errorf("invalid cached value format")
}

func (ms *MemoryStore) StartProcessing(ctx context.Context, key string, processingTTL time.Duration) (bool, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.data[key]; exists {
		return false, nil
	}

	ms.data[key] = "processing"
	return true, nil
}

func (ms *MemoryStore) Finish(ctx context.Context, key string, status int, body []byte, resultTTL time.Duration) error {
	resp := CachedResponse{
		StatusCode: status,
		Body:       body,
		Completed:  true,
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data[key] = string(respJSON)
	return nil
}

func (ms *MemoryStore) Delete(ctx context.Context, key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.data, key)
	return nil
}

func (ms *MemoryStore) Close() error {
	return nil
}
