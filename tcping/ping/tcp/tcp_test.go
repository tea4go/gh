package tcp_test

import (
	"context"
	"testing"

	"github.com/tea4go/gh/tcping/ping"
	"github.com/tea4go/gh/tcping/ping/tcp"
)

func TestPing(t *testing.T) {
	ph := tcp.NewTCP("google.com", 80, &ping.TOption{})
	_ = ph.Ping(context.Background())
}

func TestPing_Failed(t *testing.T) {
	ph := tcp.NewTCP("127.0.0.1", 1, &ping.TOption{})
	stats := ph.Ping(context.Background())
	if stats.Connected {
		t.Fatalf("it should be connected refused error")
	}
}
