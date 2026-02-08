package network

import (
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
)

type TestMsg struct {
	ID   int
	Data string
}

func TestGobConnection(t *testing.T) {
	fmt.Println("=== 开始测试: Gob 连接 (GobConnection) ===")

	// 注册消息类型
	RegisterType(reflect.TypeOf(TestMsg{}))

	// 创建内存管道模拟网络连接
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// 服务端读取
	go func() {
		defer wg.Done()
		conn := NewGobConnection(server)
		defer conn.Close()

		msg, err := conn.Read()
		if err != nil {
			t.Errorf("服务端读取失败: %v", err)
			return
		}
		
		val := msg.Interface()
		if tm, ok := val.(TestMsg); ok {
			t.Logf("服务端收到消息: %+v", tm)
			if tm.ID != 123 || tm.Data != "hello" {
				t.Errorf("消息内容不匹配")
			}
		} else {
			t.Errorf("收到错误的消息类型: %T", val)
		}
		msg.Recovery()
	}()

	// 客户端发送
	conn := NewGobConnection(client)
	defer conn.Close()

	err := conn.Write(TestMsg{ID: 123, Data: "hello"})
	if err != nil {
		t.Fatalf("客户端发送失败: %v", err)
	}

	wg.Wait()
}

func TestGetMsgType(t *testing.T) {
	fmt.Println("=== 开始测试: 获取消息类型 (GetMsgType) ===")
	typ := reflect.TypeOf(TestMsg{})
	RegisterType(typ)

	got, err := GetMsgType(typ.String())
	if err != nil {
		t.Errorf("获取消息类型失败: %v", err)
	}
	if got != typ {
		t.Errorf("获取到的类型不匹配")
	}

	all := GetMsgAllType()
	t.Logf("所有注册的类型: %v", all)
	
	DeleteType(typ.String())
	_, err = GetMsgType(typ.String())
	if err == nil {
		t.Error("删除后仍能获取到类型")
	}
}
