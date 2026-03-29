package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func problemWithRaceCondition() {
	fmt.Println("Counter с Race Condition")
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}
	wg.Wait()
	fmt.Printf("Expected: 1000, Got: %d\n\n", counter)
}

func solution1_mutex() {
	fmt.Println("Решение с sync.Mutex")
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("Counter with Mutex: %d\n\n", counter)
}

func solution2_atomic() {
	fmt.Println("Решение с sync/atomic")
	var counter int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}
	wg.Wait()
	fmt.Printf("Counter with Atomic: %d\n\n", counter)
}

func main() {
	problemWithRaceCondition()
	solution1_mutex()
	solution2_atomic()
}
