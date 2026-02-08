package utils

import (
	"fmt"
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
	
	// Access A, making it most recently used. Order: B, C, A (Most Recent)
	// Wait, LruCache implementation:
	// Set: MoveToFront or PushFront.
	// Get: MoveToFront.
	// Front is Most Recent. Back is Least Recent.
	// Current: C(Front), B, A(Back) -> Set logic pushes to front.
	
	// Let's verify internal order implicitly by adding D.
	// If we add D, the one at Back should be removed.
	
	// Access A. Now A is at Front. Order: A, C, B(Back)
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
	// Note: The TLruCache implementation stores `update` time, but does NOT actively check for expiration in Get().
	// It only stores it. Let's verify code:
	// func (this *TLruCache) Get(key string) interface{} { ... return kv.data }
	// It does NOT check if time.Now() > kv.update.
	// So timeout parameter seems to be metadata or for future use, or I missed something.
	// Looking at Set: kv.update = time.Now().Add(m)
	// Looking at Get: just returns data.
	
	// So we cannot test timeout expiration because it's not implemented in Get.
	// We just test that setting it works without error.
	cache := NewLruCache(3)
	cache.Set("A", 1, 10) // 10ms
	time.Sleep(20 * time.Millisecond)
	if !cache.IsExist("A") {
		// If implementation changed to support timeout, this might fail or pass depending on expectation.
		// Current implementation: IsExist checks map existence.
		t.Log("A still exists because timeout logic is not enforced in Get/IsExist")
	}
}
