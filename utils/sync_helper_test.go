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
	
	// Theoretically, maxActive should not exceed 2.
	// However, since "active" increment/decrement is not perfectly synchronized with the limiter channel inside Run,
	// (limiter acquires token -> launches go func -> executes f -> releases token)
	// The token is acquired BEFORE f starts. So strictly speaking, only 2 goroutines can be IN f().
	// But our "active" counter is inside f. So it should be accurate.
	
	if maxActive > 2 {
		t.Errorf("Max concurrent goroutines exceeded limit. Got %d, limit 2", maxActive)
	}
}
