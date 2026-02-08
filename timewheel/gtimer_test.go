package timewheel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestBasicFunctionality 测试基本的时间轮功能
func TestBasicFunctionality(t *testing.T) {
	fmt.Println("=== 开始测试: 基本定时任务功能 (TestBasicFunctionality) ===")
	// 创建时间轮，10个槽，每个槽10ms，6层
	tw := NewTimerPlus(10, 10*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)

	start := time.Now()
	// 添加一个200ms后执行的任务
	tw.Add("test-basic", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		defer wg.Done()
		fmt.Printf("Task %s executed\n", name)
	}, nil)

	wg.Wait()
	duration := time.Since(start)

	// 允许一定的误差，比如 +/- 50ms
	if duration < 150*time.Millisecond || duration > 300*time.Millisecond {
		t.Errorf("Expected execution around 200ms, but got %v", duration)
	}
}

// TestAddOnce 测试只执行一次的任务
func TestAddOnce(t *testing.T) {
	fmt.Println("=== 开始测试: 一次性任务功能 (TestAddOnce) ===")
	tw := NewTimerPlus(10, 10*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var count int32
	wg := sync.WaitGroup{}
	wg.Add(1)

	// AddOnce 应该只执行一次
	tw.AddOnce("test-once", 50*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&count, 1)
		wg.Done()
	}, nil)

	wg.Wait()

	// 等待一段时间，确保不会再次执行
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("Expected execution count 1, but got %d", count)
	}
}

// TestAddTimes 测试执行指定次数的任务
func TestAddTimes(t *testing.T) {
	fmt.Println("=== 开始测试: 指定次数任务功能 (TestAddTimes) ===")
	tw := NewTimerPlus(10, 10*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var count int32
	expectedTimes := 3
	wg := sync.WaitGroup{}
	wg.Add(expectedTimes)

	tw.AddTimes("test-times", 50*time.Millisecond, expectedTimes, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&count, 1)
		wg.Done()
	}, nil)

	wg.Wait()

	// 等待一段时间，确保不会再次执行
	time.Sleep(200 * time.Millisecond)

	if int(atomic.LoadInt32(&count)) != expectedTimes {
		t.Errorf("Expected execution count %d, but got %d", expectedTimes, count)
	}
}

// TestAddSingleton 测试单例任务
// 单例任务是指同一时间只能有一个该任务在运行。
// 如果任务执行时间超过间隔时间，下一次执行会被跳过或推迟，直到当前执行完成。
func TestAddSingleton(t *testing.T) {
	fmt.Println("=== 开始测试: 单例任务功能 (TestAddSingleton) ===")
	tw := NewTimerPlus(10, 10*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var count int32
	// 间隔 50ms
	// 任务执行耗时 150ms
	// 总共运行 500ms
	// 理论上非单例模式会触发约 10 次 (500/50)
	// 单例模式下，由于每次执行耗时150ms，且check间隔，实际执行次数会少很多。
	// 0ms start -> 150ms finish. Next valid tick?

	// 为了测试方便，我们让任务执行时间明显大于间隔

	done := make(chan struct{})

	tw.AddSingleton("test-singleton", 50*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&count, 1)
		time.Sleep(150 * time.Millisecond) // 模拟耗时任务
	}, nil)

	go func() {
		time.Sleep(600 * time.Millisecond)
		close(done)
	}()

	<-done

	c := atomic.LoadInt32(&count)
	fmt.Printf("Singleton task executed %d times in 600ms (interval 50ms, duration 150ms)\n", c)

	// 简单验证次数是否合理。
	// 正常 50ms 一次，600ms 应该是 12 次。
	// 单例模式下，每次耗时 150ms。
	// Start at 50ms (duration 150ms) -> ends at 200ms.
	// Next check at 250ms (or depends on implementation details of next tick check)
	// 应该远小于 12 次，大概 3-4 次。
	if c > 8 {
		t.Errorf("Singleton task executed too many times: %d, expected significantly less than theoretical max", c)
	}
}

// TestDelayAdd 测试延迟添加任务
func TestDelayAdd(t *testing.T) {
	fmt.Println("=== 开始测试: 延迟任务功能 (TestDelayAdd) ===")
	tw := NewTimerPlus(10, 10*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)

	start := time.Now()

	// 延迟 200ms 添加，然后间隔 100ms 执行一次（因为是AddOnce，其实这里DelayAdd底层调用AddOnce添加wrapper，wrapper里再Add）
	// 注意：DelayAdd系列方法在 wrapper 里的实现是 t.Add(...)，也就是会变成周期性任务。
	// 根据源码:
	// func (t *Timer) DelayAdd(...) {
	//     t.AddOnce(..., func(...) {
	//         t.Add(...)
	//     }, ...)
	// }
	// 所以它会变成一个周期性任务。我们这里只测试它第一次执行的时间。

	// 为了简化测试，我们只关心第一次执行的时间是否包含了 delay + interval
	// Delay 200ms, Interval 100ms. First execution should be around 300ms.

	firstRun := true
	tw.DelayAdd("test-delay", 200*time.Millisecond, 100*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		if firstRun {
			defer wg.Done()
			firstRun = false
			fmt.Println("Delay task executed")
		}
		// 这是一个周期性任务，我们需要手动停止它，或者让测试结束自动销毁
	}, nil)

	wg.Wait()
	duration := time.Since(start)

	if duration < 250*time.Millisecond {
		t.Errorf("Expected delay execution > 250ms (200 delay + ~interval), got %v", duration)
	}
}

// TestCancelTask 测试取消任务
func TestCancelTask(t *testing.T) {
	fmt.Println("=== 开始测试: 任务取消功能 (TestCancelTask) ===")
	tw := NewTimerPlus(10, 10*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var count int32
	// 添加一个周期性任务
	entry := tw.Add("test-cancel", 50*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&count, 1)
	}, nil)

	// 让它跑一会
	time.Sleep(120 * time.Millisecond)

	// 取消任务
	entry.Stop() // 或者 entry.Close()

	currentCount := atomic.LoadInt32(&count)
	fmt.Printf("Task executed %d times before cancel\n", currentCount)

	if currentCount == 0 {
		t.Error("Task should have executed at least once")
	}

	// 再等待一会，确认计数不再增加
	time.Sleep(200 * time.Millisecond)

	finalCount := atomic.LoadInt32(&count)
	if finalCount != currentCount {
		t.Errorf("Task continued to execute after cancellation. Before: %d, After: %d", currentCount, finalCount)
	}
}
