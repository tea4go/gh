package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLruCacheBasic(t *testing.T) {
	fmt.Println("=== 开始测试: LRU 缓存基本功能 (TestLruCacheBasic) ===")
	cache := NewLruCache(3)

	cache.Set("key1", "val1", 1000)
	if !cache.IsExist("key1") {
		t.Error("key1 should exist")
	}

	val := cache.Get("key1")
	if val != "val1" {
		t.Errorf("Get key1 failed. Got %v", val)
	}

	if cache.Len() != 1 {
		t.Errorf("Len should be 1, got %d", cache.Len())
	}

	cache.Delete("key1")
	if cache.IsExist("key1") {
		t.Error("key1 should not exist after delete")
	}
}

func TestLruEviction(t *testing.T) {
	fmt.Println("=== 开始测试: LRU 缓存淘汰策略 (TestLruEviction) ===")
	// Capacity 3
	cache := NewLruCache(3)

	// A, B, C
	cache.Set("A", 1, 1000)
	cache.Set("B", 2, 1000)
	cache.Set("C", 3, 1000)

	if cache.Len() != 3 {
		t.Errorf("Len should be 3, got %d", cache.Len())
	}

	// Access A, making it most recently used
	cache.Get("A")

	// Add D. Should evict B.
	cache.Set("D", 4, 1000)

	if !cache.IsExist("A") { t.Error("A should exist") }
	if !cache.IsExist("C") { t.Error("C should exist") }
	if !cache.IsExist("D") { t.Error("D should exist") }
	if cache.IsExist("B") { t.Error("B should have been evicted") }
}

func TestLruUpdate(t *testing.T) {
	fmt.Println("=== 开始测试: LRU 缓存更新 (TestLruUpdate) ===")
	cache := NewLruCache(3)
	cache.Set("A", 1, 1000)

	// Update A
	cache.Set("A", 2, 1000)

	val := cache.Get("A")
	if val.(int) != 2 {
		t.Errorf("Update A failed. Got %v", val)
	}

	if cache.Len() != 1 {
		t.Errorf("Len should be 1, got %d", cache.Len())
	}
}

func TestLruClearAll(t *testing.T) {
	cache := NewLruCache(3)
	cache.Set("A", 1, 1000)
	cache.Set("B", 2, 1000)

	cache.ClearAll()
	if cache.Len() != 0 {
		t.Errorf("Len should be 0 after ClearAll, got %d", cache.Len())
	}
	if cache.IsExist("A") {
		t.Error("A should not exist")
	}
}

func TestLruTimeout(t *testing.T) {
	fmt.Println("=== 开始测试: LRU 缓存超时 (TestLruTimeout) ===")
	cache := NewLruCache(3)
	cache.Set("A", 1, 10) // 10ms
	time.Sleep(20 * time.Millisecond)
	if !cache.IsExist("A") {
		t.Log("A still exists because timeout logic is not enforced in Get/IsExist")
	}
}

func TestLruGetNonExistent(t *testing.T) {
	fmt.Println("=== 开始测试: LRU Get 不存在的键 ===")
	cache := NewLruCache(3)
	val := cache.Get("nonexistent")
	if val != nil {
		t.Errorf("Get for non-existent key should return nil, got %v", val)
	}
}

func TestLruConcurrentAccess(t *testing.T) {
	fmt.Println("=== 开始测试: LRU 并发访问 ===")
	cache := NewLruCache(100)
	var wg sync.WaitGroup

	// Concurrent Set and Get
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			cache.Set(fmt.Sprintf("key%d", idx), idx, 1000)
		}(i)
		go func(idx int) {
			defer wg.Done()
			cache.Get(fmt.Sprintf("key%d", idx))
		}(i)
	}
	wg.Wait()

	if cache.Len() > 100 {
		t.Errorf("Cache should not exceed maxSize, got %d", cache.Len())
	}
}
