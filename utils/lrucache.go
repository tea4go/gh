package utils

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

//Least Recently Used 近期最少使用算法
//	lru := algorithm.NewLRUCache(3)

//  A - lru.Set(10,"value1")  <== 当D加进来时A被更新为value4
//	B - lru.Set(20,"value2")  <== 当E加进来时B被清除
//	C - lru.Set(30,"value3")
//	D - lru.Set(10,"value4")
//	E - lru.Set(50,"value5")

//内部数据对象
type nodeItem struct {
	name   string
	update time.Time
	data   interface{}
}

//container/list 简单介绍
//  Element.Value interface{}   //在元素中存储的值
//  func (l *List) MoveToBack(e *Element) //将元素e移动到list的末尾，如果e不属于list，则list不改变。
//  func (l *List) MoveToFront(e *Element)//将元素e移动到list的首部，如果e不属于list，则list不改变。
//  func (l *List) PushFront(v interface{}) *Element//在list的首部插入值为v的元素，并返回该元素。
//  URL: http://studygolang.com/articles/4371

type TLruCache struct {
	itemList *list.List
	itemMap  map[string]*list.Element
	maxSize  int
	lock     sync.Mutex
}

func (this *TLruCache) Get(key string) interface{} {
	elem, ok := this.itemMap[key]
	if !ok {
		return nil
	}
	//将元素e移动到list的首部
	this.itemList.MoveToFront(elem)

	//elem.Value得到元素的值
	kv := elem.Value.(*nodeItem)
	return kv.data
}

func (this *TLruCache) Set(key string, val interface{}, timeout int64) {
	this.lock.Lock()
	defer this.lock.Unlock()

	m, _ := time.ParseDuration(fmt.Sprintf("%dms", timeout))

	elem, ok := this.itemMap[key]

	if ok {
		this.itemList.MoveToFront(elem)
		kv := elem.Value.(*nodeItem)
		kv.data = val
		kv.update = time.Now().Add(m)
	} else {
		elem := this.itemList.PushFront(&nodeItem{name: key, data: val, update: time.Now().Add(m)})
		this.itemMap[key] = elem

		if this.itemList.Len() > this.maxSize {
			delElem := this.itemList.Back()
			kv := delElem.Value.(*nodeItem)
			this.itemList.Remove(delElem)
			delete(this.itemMap, kv.name)
		}
	}
}

func (this *TLruCache) Delete(key string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	elem, ok := this.itemMap[key]
	if ok {
		this.itemList.Remove(elem)
		delete(this.itemMap, key)
	}
}

func (this *TLruCache) IsExist(key string) bool {
	if _, ok := this.itemMap[key]; ok {
		return true
	}
	return false
}

func (this *TLruCache) ClearAll() {
	this.lock.Lock()
	defer this.lock.Unlock()

	for k, e := range this.itemMap {
		this.itemList.Remove(e)
		delete(this.itemMap, k)
	}
}

func (this *TLruCache) Len() int {
	return this.itemList.Len()
}

func NewLruCache(maxSize int) *TLruCache {
	return &TLruCache{
		itemList: list.New(),
		itemMap:  make(map[string]*list.Element),
		maxSize:  maxSize,
	}
}
