package utils

import (
	"errors"
	"fmt"
)

type TStack struct {
	Element []string
}

func NewStack() *TStack {
	return &TStack{}
}

func (stack *TStack) Push(value string) {
	stack.Element = append(stack.Element, value)
}

func (stack *TStack) Find(value string) bool {
	for i := len(stack.Element) - 1; i >= 0; i-- {
		if value == stack.Element[i] {
			return true
		}
	}

	return false
}

//返回下一个元素
func (stack *TStack) Top() (value string) {
	if stack.Size() > 0 {
		return stack.Element[stack.Size()-1]
	}
	return "" //read empty stack
}

//返回下一个元素,并从Stack移除元素
func (stack *TStack) PopByValue(value string) (err error) {
	for i := len(stack.Element) - 1; i >= 0; i-- {
		if value == stack.Element[i] {
			stack.Element = stack.Element[:i+1]
			return nil
		}
	}
	return errors.New("Stack为空.") //read empty stack
}

//返回下一个元素,并从Stack移除元素
func (stack *TStack) Pop() (err error) {
	if stack.Size() > 0 {
		stack.Element = stack.Element[:stack.Size()-1]
		return nil
	}
	return errors.New("Stack为空.") //read empty stack
}

//交换值
func (stack *TStack) Swap(other *TStack) {
	switch {
	case stack.Size() == 0 && other.Size() == 0:
		return
	case other.Size() == 0:
		other.Element = stack.Element[:stack.Size()]
		stack.Element = nil
	case stack.Size() == 0:
		stack.Element = other.Element
		other.Element = nil
	default:
		stack.Element, other.Element = other.Element, stack.Element
	}
	return
}

//修改指定索引的元素
func (stack *TStack) Set(idx int, value string) (err error) {
	if idx >= 0 && stack.Size() > 0 && stack.Size() > idx {
		stack.Element[idx] = value
		return nil
	}
	return errors.New("Set失败!")
}

//返回指定索引的元素
func (stack *TStack) Get(idx int) (value interface{}) {
	if idx >= 0 && stack.Size() > 0 && stack.Size() > idx {
		return stack.Element[idx]
	}
	return nil //read empty stack
}

//Stack的size
func (stack *TStack) Size() int {
	return len(stack.Element)
}

//是否为空
func (stack *TStack) Empty() bool {
	if stack.Element == nil || stack.Size() == 0 {
		return true
	}
	return false
}

//打印
func (stack *TStack) Print() {
	for i := len(stack.Element) - 1; i >= 0; i-- {
		fmt.Println(i, "=>", stack.Element[i])
	}
}
