package utils

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSingletonBasic(t *testing.T) {
	fmt.Println("=== 开始测试: 单例模式基本功能 (TestSingletonBasic) ===")
	count := 0
	init := func() (interface{}, error) {
		count++
		return "data", nil
	}
	
	s := NewSingleton(init)
	
	// First call
	val, err := s.GetSingleton()
	if err != nil {
		t.Fatalf("GetSingleton failed: %v", err)
	}
	if val.(string) != "data" {
		t.Errorf("Unexpected data: %v", val)
	}
	if count != 1 {
		t.Errorf("Init called %d times, want 1", count)
	}
	
	// Second call
	val, _ = s.GetSingleton()
	if val.(string) != "data" {
		t.Errorf("Unexpected data: %v", val)
	}
	if count != 1 {
		t.Errorf("Init called %d times, want 1", count)
	}
}

func TestSingletonConcurrency(t *testing.T) {
	fmt.Println("=== 开始测试: 单例模式并发安全 (TestSingletonConcurrency) ===")
	var count int32
	init := func() (interface{}, error) {
		time.Sleep(10 * time.Millisecond) // Simulate work
		atomic.AddInt32(&count, 1)
		return "shared", nil
	}
	
	s := NewSingleton(init)
	
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := s.GetSingleton()
			if err != nil {
				t.Errorf("Concurrent Get failed: %v", err)
			}
			if val != "shared" {
				t.Errorf("Got wrong value")
			}
		}()
	}
	wg.Wait()
	
	if c := atomic.LoadInt32(&count); c != 1 {
		t.Errorf("Init called %d times, want 1", c)
	}
}

func TestSingletonError(t *testing.T) {
	fmt.Println("=== 开始测试: 单例模式错误处理 (TestSingletonError) ===")
	init := func() (interface{}, error) {
		return nil, errors.New("init failed")
	}
	
	s := NewSingleton(init)
	
	_, err := s.GetSingleton()
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	// Should retry if failed?
	// Implementation:
	/*
	s.data, err = s.init_func()
	if err != nil {
		return nil, err
	}
	s.init_flag = true
	*/
	// If it fails, init_flag is NOT set. Next call will retry.
	
	// Change init func behavior for second call
	fail := true
	initRetry := func() (interface{}, error) {
		if fail {
			fail = false
			return nil, errors.New("fail once")
		}
		return "success", nil
	}
	
	s2 := NewSingleton(initRetry)
	_, err = s2.GetSingleton()
	if err == nil {
		t.Error("First call should fail")
	}
	
	val, err := s2.GetSingleton()
	if err != nil {
		t.Errorf("Second call should succeed, got %v", err)
	}
	if val != "success" {
		t.Error("Got wrong value")
	}
}
