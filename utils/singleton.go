package utils

import (
	"sync"
	"sync/atomic"
)

type SingletonInitFunc func() (interface{}, error)

// 接口，用于访问单一对象
type ISingleton interface {
	// Return the encapsulated singleton
	GetSingleton() (interface{}, error)
}

func NewSingleton(init SingletonInitFunc) ISingleton {
	return &TSingleton{init_func: init}
}

type TSingleton struct {
	sync.Mutex
	data      interface{}
	init_func SingletonInitFunc
	init_flag atomic.Bool
}

func (s *TSingleton) GetSingleton() (interface{}, error) {
	if s.init_flag.Load() {
		return s.data, nil
	}

	s.Lock()
	defer s.Unlock()

	if s.init_flag.Load() {
		return s.data, nil
	}

	var err error
	s.data, err = s.init_func()
	if err != nil {
		return nil, err
	}

	s.init_flag.Store(true)
	return s.data, nil
}
