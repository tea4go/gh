package utils

import (
	"runtime"
)

type Glimit struct {
	n int
	c chan int8
}

// initialization Glimit struct
func SetNumGoroutine(n int) *Glimit {
	if n <= 0 {
		n = runtime.NumGoroutine()
	}
	return &Glimit{
		n: n,
		c: make(chan int8, n),
	}
}

// Run f in a new goroutine but with limit.
func (g *Glimit) Run(f func()) {
	g.c <- 1
	go func() {
		f()
		<-g.c
	}()
}
