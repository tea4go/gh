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
	
	// PopByValue implementation:
	// Find value from top, reset slice to [:i+1] ???
	// Let's check source:
	/*
	func (stack *TStack) PopByValue(value string) (err error) {
		for i := len(stack.Element) - 1; i >= 0; i-- {
			if value == stack.Element[i] {
				stack.Element = stack.Element[:i+1] // Wait, this keeps everything UP TO i (inclusive) ??
				// If stack is [A, B, C], and we PopByValue("B") (i=1)
				// stack.Element = stack.Element[:2] -> [A, B]
				// So "C" is removed. "B" stays at top.
				// This behavior seems to be "Revert to state where 'value' is top"?
				// The name "PopByValue" suggests removing 'value', but implementation looks like "RollbackTo".
				return nil
			}
		}
		return errors.New("Stack为空.") // Actually "Not Found" but returns "Stack empty" msg
	}
	*/
	
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
