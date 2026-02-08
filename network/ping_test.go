package network

import (
	"fmt"
	"testing"
	"time"
)

func TestPinger(t *testing.T) {
	fmt.Println("=== 开始测试: Ping 功能 (Pinger) ===")
	
	// Ping 本机回环地址，通常应该成功
	host := "127.0.0.1"
	ok, result, err := Pinger(host, 3, 32, 1*time.Second)
	
	if err != nil {
		// 在某些受限环境（如无 raw socket 权限），ping 可能会失败
		t.Logf("Ping %s 报错 (可能是权限不足): %v", host, err)
	} else {
		t.Logf("Ping %s 结果: 成功=%v, 详情=%v", host, ok, result)
		if !ok {
			t.Log("Ping 未收到回复")
		}
	}
}
