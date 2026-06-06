package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGlimit(t *testing.T) {
	fmt.Println("=== 开始测试: Goroutine 限制器 (TestGlimit) ===")
	// Limit to 2 concurrent goroutines
	limiter := SetNumGoroutine(2)

	var active int32
	var maxActive int32
	var wg sync.WaitGroup

	// Start 10 tasks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		limiter.Run(func() {
			defer wg.Done()

			current := atomic.AddInt32(&active, 1)

			// Update maxActive safely
			for {
				oldMax := atomic.LoadInt32(&maxActive)
				if current <= oldMax {
					break
				}
				if atomic.CompareAndSwapInt32(&maxActive, oldMax, current) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&active, -1)
		})
	}

	wg.Wait()

	if maxActive > 2 {
		t.Errorf("Max concurrent goroutines exceeded limit. Got %d, limit 2", maxActive)
	}
}

func TestSetNumGoroutineZero(t *testing.T) {
	fmt.Println("=== 开始测试: SetNumGoroutine 零值 ===")
	// When n <= 0, it should use runtime.NumGoroutine()
	limiter := SetNumGoroutine(0)
	if limiter == nil {
		t.Error("SetNumGoroutine(0) should not return nil")
	}
	if limiter.n <= 0 {
		t.Errorf("SetNumGoroutine(0) should set n to runtime.NumGoroutine(), got %d", limiter.n)
	}
	t.Logf("SetNumGoroutine(0) set n to %d", limiter.n)
}

func TestSetNumGoroutineNegative(t *testing.T) {
	fmt.Println("=== 开始测试: SetNumGoroutine 负值 ===")
	limiter := SetNumGoroutine(-5)
	if limiter == nil {
		t.Error("SetNumGoroutine(-5) should not return nil")
	}
	if limiter.n <= 0 {
		t.Errorf("SetNumGoroutine(-5) should set n to runtime.NumGoroutine(), got %d", limiter.n)
	}
}

func TestGlimitSingleGoroutine(t *testing.T) {
	fmt.Println("=== 开始测试: Glimit 单协程 ===")
	limiter := SetNumGoroutine(1)
	var counter int32
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		limiter.Run(func() {
			defer wg.Done()
			atomic.AddInt32(&counter, 1)
		})
	}

	wg.Wait()
	if atomic.LoadInt32(&counter) != 5 {
		t.Errorf("Expected 5 executions, got %d", atomic.LoadInt32(&counter))
	}
}
