package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("Решение с sync.Map")
	var safeMap sync.Map
	var wg sync.WaitGroup

	// Горутины пишут в карту
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			safeMap.Store("key", key)
		}(i)
	}
	wg.Wait()

	value, ok := safeMap.Load("key")
	if ok {
		fmt.Printf("Value from sync.Map: %v\n\n", value)
	}

	fmt.Println("Решение с sync.RWMutex")
	var mu sync.RWMutex
	safeMapWithMutex := make(map[string]int)
	wg2 := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg2.Add(1)
		go func(key int) {
			defer wg2.Done()
			mu.Lock()
			safeMapWithMutex["key"] = key
			mu.Unlock()
		}(i)
	}
	wg2.Wait()

	mu.RLock()
	v := safeMapWithMutex["key"]
	mu.RUnlock()

	fmt.Printf("Value from map with RWMutex: %v\n", v)
}
