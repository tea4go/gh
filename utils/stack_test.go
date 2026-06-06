package utils

import (
	"fmt"
	"testing"
)

func TestStackBasic(t *testing.T) {
	fmt.Println("=== 开始测试: 栈基本操作 (TestStackBasic) ===")
	s := NewStack()
	if !s.Empty() {
		t.Error("New stack should be empty")
	}

	s.Push("A")
	s.Push("B")

	if s.Size() != 2 {
		t.Errorf("Size should be 2, got %d", s.Size())
	}

	if s.Top() != "B" {
		t.Errorf("Top should be B, got %s", s.Top())
	}

	if err := s.Pop(); err != nil {
		t.Error(err)
	}

	if s.Top() != "A" {
		t.Errorf("Top should be A, got %s", s.Top())
	}

	s.Pop()
	if !s.Empty() {
		t.Error("Stack should be empty")
	}

	if err := s.Pop(); err == nil {
		t.Error("Pop empty stack should error")
	}
}

func TestStackFind(t *testing.T) {
	fmt.Println("=== 开始测试: 栈查找 (TestStackFind) ===")
	s := NewStack()
	s.Push("A")
	s.Push("B")

	if !s.Find("A") {
		t.Error("Should find A")
	}
	if !s.Find("B") {
		t.Error("Should find B")
	}
	if s.Find("C") {
		t.Error("Should not find C")
	}
}

func TestStackPopByValue(t *testing.T) {
	fmt.Println("=== 开始测试: 按值出栈 (TestStackPopByValue) ===")
	s := NewStack()
	s.Push("A")
	s.Push("B")
	s.Push("C")

	err := s.PopByValue("B")
	if err != nil {
		t.Error(err)
	}

	if s.Size() != 2 {
		t.Errorf("Size should be 2, got %d", s.Size())
	}
	if s.Top() != "B" {
		t.Errorf("Top should be B, got %s", s.Top())
	}

	// Check if C is gone
	if s.Find("C") {
		t.Error("C should be gone")
	}
}

func TestStackPopByValueNotFound(t *testing.T) {
	fmt.Println("=== 开始测试: PopByValue 未找到 ===")
	s := NewStack()
	s.Push("A")
	s.Push("B")

	err := s.PopByValue("Z")
	if err == nil {
		t.Error("PopByValue for non-existent value should return error")
	}
}

func TestStackPopByValueEmpty(t *testing.T) {
	fmt.Println("=== 开始测试: PopByValue 空栈 ===")
	s := NewStack()
	err := s.PopByValue("A")
	if err == nil {
		t.Error("PopByValue on empty stack should return error")
	}
}

func TestStackSwap(t *testing.T) {
	fmt.Println("=== 开始测试: 栈交换 (TestStackSwap) ===")
	s1 := NewStack()
	s1.Push("A")

	s2 := NewStack()
	s2.Push("B")

	s1.Swap(s2)

	if s1.Top() != "B" {
		t.Errorf("s1 Top should be B, got %s", s1.Top())
	}
	if s2.Top() != "A" {
		t.Errorf("s2 Top should be A, got %s", s2.Top())
	}
}

func TestStackSwapEmptyBoth(t *testing.T) {
	fmt.Println("=== 开始测试: 栈交换 双空 ===")
	s1 := NewStack()
	s2 := NewStack()
	s1.Swap(s2)
	if !s1.Empty() || !s2.Empty() {
		t.Error("Both empty stacks should remain empty after swap")
	}
}

func TestStackSwapOneEmpty(t *testing.T) {
	fmt.Println("=== 开始测试: 栈交换 一方空 ===")
	// s1 has items, s2 is empty
	s1 := NewStack()
	s1.Push("A")
	s1.Push("B")
	s2 := NewStack()

	s1.Swap(s2)
	if !s1.Empty() {
		t.Errorf("s1 should be empty after swap, size=%d", s1.Size())
	}
	if s2.Size() != 2 {
		t.Errorf("s2 should have 2 items after swap, size=%d", s2.Size())
	}
	if s2.Top() != "B" {
		t.Errorf("s2 Top should be B, got %s", s2.Top())
	}

	// Reverse: s1 is now empty, s2 has items
	s1.Swap(s2)
	if !s2.Empty() {
		t.Errorf("s2 should be empty after reverse swap, size=%d", s2.Size())
	}
	if s1.Size() != 2 {
		t.Errorf("s1 should have 2 items after reverse swap, size=%d", s1.Size())
	}
}

func TestStackGetSet(t *testing.T) {
	fmt.Println("=== 开始测试: 栈索引操作 (TestStackGetSet) ===")
	s := NewStack()
	s.Push("A")

	if s.Get(0).(string) != "A" {
		t.Error("Get failed")
	}

	s.Set(0, "B")
	if s.Get(0).(string) != "B" {
		t.Error("Set failed")
	}

	if s.Get(10) != nil {
		t.Error("Get OOB should return nil")
	}
}

func TestStackSetInvalid(t *testing.T) {
	fmt.Println("=== 开始测试: Set 无效索引 ===")
	s := NewStack()
	s.Push("A")

	// Negative index
	if err := s.Set(-1, "B"); err == nil {
		t.Error("Set with negative index should fail")
	}

	// Out of bounds
	if err := s.Set(10, "B"); err == nil {
		t.Error("Set with OOB index should fail")
	}
}

func TestStackGetInvalid(t *testing.T) {
	fmt.Println("=== 开始测试: Get 无效索引 ===")
	s := NewStack()
	// Empty stack
	if v := s.Get(0); v != nil {
		t.Error("Get on empty stack should return nil")
	}
	// Negative index
	if v := s.Get(-1); v != nil {
		t.Error("Get with negative index should return nil")
	}
}

func TestStackTopEmpty(t *testing.T) {
	fmt.Println("=== 开始测试: Top 空栈 ===")
	s := NewStack()
	if v := s.Top(); v != "" {
		t.Errorf("Top on empty stack should return empty string, got %q", v)
	}
}

func TestStackEmptyNil(t *testing.T) {
	fmt.Println("=== 开始测试: Empty nil Element ===")
	s := NewStack()
	s.Element = nil
	if !s.Empty() {
		t.Error("Stack with nil Element should be empty")
	}
}

func TestStackEmptyNonEmpty(t *testing.T) {
	fmt.Println("=== 开始测试: Empty 非空栈 ===")
	s := NewStack()
	s.Push("A")
	if s.Empty() {
		t.Error("Non-empty stack should not be empty")
	}
}

func TestStackPrint(t *testing.T) {
	fmt.Println("=== 开始测试: Print ===")
	s := NewStack()
	s.Push("A")
	s.Push("B")
	s.Push("C")
	// Just test that Print doesn't panic
	s.Print()
}

func TestStackSize(t *testing.T) {
	fmt.Println("=== 开始测试: Size ===")
	s := NewStack()
	if s.Size() != 0 {
		t.Errorf("Empty stack size should be 0, got %d", s.Size())
	}
	s.Push("X")
	if s.Size() != 1 {
		t.Errorf("Stack size should be 1, got %d", s.Size())
	}
}

func TestStackNewStack(t *testing.T) {
	fmt.Println("=== 开始测试: NewStack ===")
	s := NewStack()
	if s == nil {
		t.Error("NewStack should not return nil")
	}
	if s.Size() != 0 {
		t.Error("New stack should have size 0")
	}
}
