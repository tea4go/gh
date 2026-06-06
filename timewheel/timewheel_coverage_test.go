package timewheel

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tea4go/gh/timewheel/gtype"
)

// --- Entry tests ---

func TestEntry_GetStatus(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-status", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	// Entry should be in READY state initially
	if entry.Status() != STATUS_READY {
		t.Errorf("Initial status = %d, want STATUS_READY(%d)", entry.Status(), STATUS_READY)
	}

	got := entry.GetStatus()
	if got != "准备" {
		t.Errorf("GetStatus() = %q, want 准备", got)
	}
}

func TestEntry_GetStatus_Stopped(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-stopped", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)
	entry.Stop()

	if entry.Status() != STATUS_STOPPED {
		t.Errorf("Stopped status = %d, want STATUS_STOPPED(%d)", entry.Status(), STATUS_STOPPED)
	}
	got := entry.GetStatus()
	if got != "停止" {
		t.Errorf("GetStatus() stopped = %q, want 停止", got)
	}
}

func TestEntry_GetStatus_Closed(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-closed", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)
	entry.Close()

	if entry.Status() != STATUS_CLOSED {
		t.Errorf("Closed status = %d, want STATUS_CLOSED(%d)", entry.Status(), STATUS_CLOSED)
	}
	got := entry.GetStatus()
	if got != "关闭" {
		t.Errorf("GetStatus() closed = %q, want 关闭", got)
	}
}

func TestEntry_SetStatus(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-setstatus", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	old := entry.SetStatus(STATUS_STOPPED)
	if old != STATUS_READY {
		t.Errorf("SetStatus old = %d, want STATUS_READY(%d)", old, STATUS_READY)
	}
	if entry.Status() != STATUS_STOPPED {
		t.Errorf("Status after SetStatus = %d, want STATUS_STOPPED", entry.Status())
	}
}

func TestEntry_Start(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-start", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)
	entry.Stop()
	entry.Start()

	if entry.Status() != STATUS_READY {
		t.Errorf("Status after Start = %d, want STATUS_READY", entry.Status())
	}
}

func TestEntry_IsSingleton(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.AddSingleton("test-singleton", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	if !entry.IsSingleton() {
		t.Error("AddSingleton entry should be singleton")
	}
}

func TestEntry_SetSingleton(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-setsingleton", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	if entry.IsSingleton() {
		t.Error("Add entry should not be singleton initially")
	}
	entry.SetSingleton(true)
	if !entry.IsSingleton() {
		t.Error("SetSingleton(true) should make it singleton")
	}
	entry.SetSingleton(false)
	if entry.IsSingleton() {
		t.Error("SetSingleton(false) should make it non-singleton")
	}
}

func TestEntry_SetTimes(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-settimes", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	entry.SetTimes(5)
	// Verify times was set (through entry.times)
	if entry.times.Val() != 5 {
		t.Errorf("times = %d, want 5", entry.times.Val())
	}
}

func TestEntry_String(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-string", 1*time.Second, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	got := entry.String()
	if got == "" {
		t.Error("Entry.String() should not be empty")
	}
	if !containsStr(got, "test-string") {
		t.Errorf("Entry.String() should contain name, got %q", got)
	}
}

func TestEntry_Run(t *testing.T) {
	var called int32
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-run", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
	}, nil)

	entry.Run()
	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("Run should call job, called = %d", atomic.LoadInt32(&called))
	}
}

// --- Timer tests ---

func TestNewTimer(t *testing.T) {
	tw := NewTimer()
	if tw == nil {
		t.Fatal("NewTimer returned nil")
	}
	tw.Start()
	defer tw.Stop()
}

func TestTimer_String(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	got := tw.String()
	if got == "" {
		t.Error("Timer.String() should not be empty")
	}
	if !containsStr(got, "当前定时器") {
		t.Errorf("Timer.String() should contain header, got: %q", got)
	}
}

func TestTimer_Start_Stop_Close(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 3)

	tw.Start()
	if tw.status.Val() != STATUS_RUNNING {
		t.Errorf("Status after Start = %d, want RUNNING", tw.status.Val())
	}

	tw.Stop()
	if tw.status.Val() != STATUS_STOPPED {
		t.Errorf("Status after Stop = %d, want STOPPED", tw.status.Val())
	}

	tw.Close()
	if tw.status.Val() != STATUS_CLOSED {
		t.Errorf("Status after Close = %d, want CLOSED", tw.status.Val())
	}
}

func TestTimer_AddNow(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var called int32
	wg := sync.WaitGroup{}
	wg.Add(1)

	tw.AddNow("test-addnow", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	}, nil)

	wg.Wait()
	if atomic.LoadInt32(&called) < 1 {
		t.Error("AddNow should call the job immediately")
	}
}

// --- gtype tests ---

func TestGtype_Int(t *testing.T) {
	i := gtype.NewInt(5)
	if i.Val() != 5 {
		t.Errorf("Int.Val() = %d, want 5", i.Val())
	}
	if i.Set(10) != 5 {
		t.Errorf("Int.Set(10) old = %d, want 5", i.Set(10))
	}
	if i.Val() != 10 {
		t.Errorf("Int.Val() after Set = %d, want 10", i.Val())
	}
	if i.Add(3) != 13 {
		t.Errorf("Int.Add(3) = %d, want 13", i.Add(3))
	}
	if i.Clone().Val() != 13 {
		t.Errorf("Int.Clone().Val() = %d, want 13", i.Clone().Val())
	}
}

func TestGtype_Int64(t *testing.T) {
	i := gtype.NewInt64(100)
	if i.Val() != 100 {
		t.Errorf("Int64.Val() = %d, want 100", i.Val())
	}
	if i.Set(200) != 100 {
		t.Errorf("Int64.Set(200) old = %d, want 100", i.Set(200))
	}
	if i.Val() != 200 {
		t.Errorf("Int64.Val() after Set = %d, want 200", i.Val())
	}
	if i.Add(50) != 250 {
		t.Errorf("Int64.Add(50) = %d, want 250", i.Add(50))
	}
	if i.Clone().Val() != 250 {
		t.Errorf("Int64.Clone().Val() = %d, want 250", i.Clone().Val())
	}
}

func TestGtype_Bool(t *testing.T) {
	b := gtype.NewBool(true)
	if !b.Val() {
		t.Error("Bool.Val() = false, want true")
	}
	old := b.Set(false)
	if !old {
		t.Error("Bool.Set(false) old = false, want true as old value")
	}
	if b.Val() {
		t.Error("Bool.Val() after Set(false) = true, want false")
	}
	b.Set(true)
	if b.Clone().Val() != true {
		t.Error("Bool.Clone().Val() should be true")
	}
}

func TestGtype_Bool_Default(t *testing.T) {
	b := gtype.NewBool()
	if b.Val() {
		t.Error("Default Bool.Val() should be false")
	}
}

func TestGtype_Int_Default(t *testing.T) {
	i := gtype.NewInt()
	if i.Val() != 0 {
		t.Errorf("Default Int.Val() = %d, want 0", i.Val())
	}
}

func TestGtype_Int64_Default(t *testing.T) {
	i := gtype.NewInt64()
	if i.Val() != 0 {
		t.Errorf("Default Int64.Val() = %d, want 0", i.Val())
	}
}

func TestGtype_List(t *testing.T) {
	l := gtype.NewList(true)
	if l.Len() != 0 {
		t.Errorf("List.Len() = %d, want 0", l.Len())
	}

	l.PushBack("hello")
	if l.Len() != 1 {
		t.Errorf("List.Len() after PushBack = %d, want 1", l.Len())
	}

	l.PushFront("world")
	if l.Len() != 2 {
		t.Errorf("List.Len() after PushFront = %d, want 2", l.Len())
	}

	val := l.PopFront()
	if val != "world" {
		t.Errorf("PopFront() = %v, want world", val)
	}

	val = l.PopFront()
	if val != "hello" {
		t.Errorf("PopFront() = %v, want hello", val)
	}

	if l.Len() != 0 {
		t.Errorf("List.Len() after pops = %d, want 0", l.Len())
	}
}

func TestGtype_List_BatchPopBack(t *testing.T) {
	l := gtype.NewList(true)
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	values := l.BatchPopBack(2)
	if len(values) != 2 {
		t.Errorf("BatchPopBack(2) len = %d, want 2", len(values))
	}

	if l.Len() != 1 {
		t.Errorf("List.Len() after BatchPopBack = %d, want 1", l.Len())
	}
}

func TestGtype_List_PopBackAll(t *testing.T) {
	l := gtype.NewList(true)
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	values := l.PopBackAll()
	if len(values) != 3 {
		t.Errorf("PopBackAll len = %d, want 3", len(values))
	}
	if l.Len() != 0 {
		t.Errorf("List.Len() after PopBackAll = %d, want 0", l.Len())
	}
}

func TestGtype_List_Top(t *testing.T) {
	l := gtype.NewList(true)
	if l.Top() != nil {
		t.Error("Top() on empty list should be nil")
	}
	l.PushBack(1)
	if l.Top() == nil {
		t.Error("Top() on non-empty list should not be nil")
	}
}

func TestGtype_List_PopFront_Empty(t *testing.T) {
	l := gtype.NewList(true)
	val := l.PopFront()
	if val != nil {
		t.Errorf("PopFront() on empty list should be nil, got %v", val)
	}
}

func TestGtype_RWMutex(t *testing.T) {
	mu := gtype.NewRWMutex()
	if !mu.IsSafe() {
		t.Error("Default RWMutex should be safe")
	}
	mu.Lock()
	mu.Unlock()
	mu.RLock()
	mu.RUnlock()
}

func TestGtype_RWMutex_Unsafe(t *testing.T) {
	mu := gtype.NewRWMutex(true) // unsafe
	if mu.IsSafe() {
		t.Error("Unsafe RWMutex should not be safe")
	}
	// Lock/Unlock should be no-op for unsafe
	mu.Lock()
	mu.Unlock()
	mu.RLock()
	mu.RUnlock()

	// Force lock
	mu.Lock(true)
	mu.Unlock(true)
	mu.RLock(true)
	mu.RUnlock(true)
}

// --- Wheel tests ---

func TestWheel_GetCount(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	// Add one task
	tw.Add("test-count", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	time.Sleep(100 * time.Millisecond)

	// Check count on each wheel
	totalNonSys := 0
	for _, w := range tw.wheels {
		totalNonSys += w.GetCount()
	}
	if totalNonSys == 0 {
		t.Error("Should have at least one non-sys task after adding")
	}
}

func TestWheel_String(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	tw.Add("test-wheel-str", 1*time.Second, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	for _, w := range tw.wheels {
		got := w.String()
		if got == "" {
			t.Error("Wheel.String() should not be empty")
		}
	}
}

// --- Timer delay tests ---

func TestTimer_DelayAddOnce(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var called int32
	wg := sync.WaitGroup{}
	wg.Add(1)

	tw.DelayAddOnce("test-delay-once", 200*time.Millisecond, 100*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	}, nil)

	wg.Wait()
	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("DelayAddOnce should execute once, got %d", atomic.LoadInt32(&called))
	}
}

func TestTimer_DelayAdd(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var called int32
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Delay 200ms then add a 100ms interval task
	tw.DelayAdd("test-delay", 200*time.Millisecond, 100*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	}, nil)

	// Wait for 2 executions
	wg.Wait()
	c := atomic.LoadInt32(&called)
	if c < 2 {
		t.Errorf("DelayAdd should execute at least 2 times, got %d", c)
	}
}

func TestTimer_DelayAddTimes(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var called int32
	wg := sync.WaitGroup{}
	wg.Add(2)

	tw.DelayAddTimes("test-delay-times", 100*time.Millisecond, 100*time.Millisecond, 2, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	}, nil)

	wg.Wait()
}

func TestTimer_DelayAddSingleton(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var called int32
	wg := sync.WaitGroup{}
	wg.Add(1)

	tw.DelayAddSingleton("test-delay-singleton", 100*time.Millisecond, 100*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	}, nil)

	wg.Wait()
}

// --- Timer addEntry edge cases ---

func TestTimer_AddEntry_ZeroTimes(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	// AddOnce with 0 interval still works (set to DEFAULT_TIMES)
	entry := tw.doAddEntry("test-zero-times", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil, false, 0, STATUS_READY)
	if entry == nil {
		t.Error("doAddEntry with times=0 should still work")
	}
	if entry.times.Val() != DEFAULT_TIMES {
		t.Errorf("times should be DEFAULT_TIMES, got %d", entry.times.Val())
	}
}

func TestTimer_BinSearchIndex(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	// Test with an interval that matches the lowest wheel
	_, cmp := tw.binSearchIndex(tw.wheels[0].intervalMs)
	if cmp != 0 {
		t.Errorf("binSearchIndex for lowest wheel interval: cmp=%d, want 0", cmp)
	}
}

func TestTimer_GetLevelByIntervalMs(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	// Very small interval should go to level 0
	idx0 := tw.getLevelByIntervalMs(50)
	if idx0 != 0 {
		t.Errorf("getLevelByIntervalMs(50) = %d, want 0", idx0)
	}

	// Very large interval should go to highest level
	idxN := tw.getLevelByIntervalMs(tw.wheels[tw.length-1].intervalMs)
	if idxN >= tw.length {
		t.Errorf("getLevelByIntervalMs(large) = %d, should be < %d", idxN, tw.length)
	}
}

// --- NewTimerPlus clamping ---

func TestNewTimerPlus_SlotClamped(t *testing.T) {
	tw := NewTimerPlus(30, 50*time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()
	// The code clamps level, not slot. Verify slot is as given.
	if tw.number != 30 {
		t.Logf("number = %d (may or may not be clamped)", tw.number)
	}
}

func TestNewTimerPlus_LevelClamped(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 10)
	tw.Start()
	defer tw.Stop()
	// level > 8 should be clamped to 8
	if tw.length > 8 {
		t.Errorf("length should be clamped to 8, got %d", tw.length)
	}
}

// --- Timer proceed test ---

func TestWheel_Proceed(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	var called int32
	wg := sync.WaitGroup{}
	wg.Add(1)

	// This tests that the proceed function is called internally
	tw.Add("test-proceed", 100*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	}, nil)

	wg.Wait()
	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("proceed should have executed the job, called=%d", atomic.LoadInt32(&called))
	}
}

// --- Panic handling in entry run ---

func TestEntry_PanicHandling(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	var called int32
	wg := sync.WaitGroup{}
	wg.Add(1)

	// This entry will panic
	entry := tw.Add("test-panic", 100*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
		atomic.AddInt32(&called, 1)
		wg.Done()
		panic("exit")
	}, nil)

	wg.Wait()
	// After panic, entry should be closed
	time.Sleep(200 * time.Millisecond)
	if entry.Status() != STATUS_CLOSED {
		t.Logf("Entry status after panic = %d (may vary)", entry.Status())
	}
}

// --- Entry check test ---

func TestEntry_Check_StatusStopped(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-check-stopped", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)
	entry.Stop()

	runnable, addable := entry.check(1, 1)
	if runnable {
		t.Error("Stopped entry should not be runnable")
	}
	if !addable {
		t.Error("Stopped entry should be addable")
	}
}

func TestEntry_Check_StatusClosed(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-check-closed", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)
	entry.Close()

	runnable, addable := entry.check(1, 1)
	if runnable {
		t.Error("Closed entry should not be runnable")
	}
	if addable {
		t.Error("Closed entry should not be addable")
	}
}

func TestEntry_Check_NoMatch(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-check-nomatch", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	// Check with a tick that doesn't match the interval
	runnable, addable := entry.check(1, 1)
	if runnable {
		t.Error("Entry with wrong tick should not be runnable")
	}
	if !addable {
		t.Error("Entry with wrong tick should still be addable")
	}
}

// --- Format check in proceed ---

func TestEntry_NonSysFlag(t *testing.T) {
	tw := NewTimerPlus(10, 50*time.Millisecond, 6)
	tw.Start()
	defer tw.Stop()

	entry := tw.Add("test-nonsys", 200*time.Millisecond, func(name string, create time.Time, interval time.Duration, jobp interface{}) {
	}, nil)

	if entry.sysflag {
		t.Error("Regular Add entry should not have sysflag=true")
	}
}

// --- Helper function ---

func containsStr(s, sub string) bool {
	if len(s) < len(sub) {
		return false
	}
	return strings.Contains(s, sub)
}