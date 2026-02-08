package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestUUIDGeneration(t *testing.T) {
	fmt.Println("=== 开始测试: UUID 生成 (TestUUIDGeneration) ===")
	uuid1 := NewUUID()
	uuid2 := NewUUID()

	if uuid1.String() == uuid2.String() {
		t.Error("生成的 UUID 应该是唯一的")
	}

	// Check Time
	// Time is encoded in first 4 bytes.
	// Allow 1-2 second difference
	now := time.Now().Unix()
	uuidTime := uuid1.GetTime().Unix()

	if uuidTime < now-5 || uuidTime > now+5 {
		t.Errorf("UUID 时间看起来不对。得到 %v, 当前 %v", uuidTime, now)
	}
}

func TestUUIDMarshalUnmarshal(t *testing.T) {
	fmt.Println("=== 开始测试: UUID 序列化 (TestUUIDMarshalUnmarshal) ===")
	uuid := NewUUID()
	str := uuid.String()

	uuid2, err := UUIDFromString(str)
	if err != nil {
		t.Fatalf("UUIDFromString 失败: %v", err)
	}

	if uuid.String() != uuid2.String() {
		t.Errorf("往返转换失败。原始 %s, 解析后 %s", uuid.String(), uuid2.String())
	}

	if *uuid != *uuid2 {
		t.Error("UUID 字节不匹配")
	}
}

func TestUUIDInvalid(t *testing.T) {
	fmt.Println("=== 开始测试: UUID 校验 (TestUUIDInvalid) ===")
	// Invalid length
	_, err := UUIDFromString("short")
	if err == nil {
		t.Error("短字符串应该失败")
	}

	// Invalid chars
	// encoding = "0123456789abcdefghijklmnopqrstuv" (base32 custom)
	// 'z' is not in encoding
	invalidStr := "0000000000000000000z"
	if len(invalidStr) != encodedLen {
		// adjust length to match encodedLen (20)
		invalidStr = "0123456789abcdefghiz"
	}

	_, err = UUIDFromString(invalidStr)
	if err == nil {
		t.Error("无效字符应该失败")
	}
}

func TestUUIDComponents(t *testing.T) {
	fmt.Println("=== 开始测试: UUID 组件解析 (TestUUIDComponents) ===")
	uuid := NewUUID()

	// Machine ID (3 bytes)
	machine := uuid.GetMachine()
	if len(machine) != 3 {
		t.Errorf("机器 ID 长度错误。得到 %d", len(machine))
	}

	// PID (2 bytes)
	pid := uuid.GetPID()
	if pid == 0 {
		t.Log("PID 为 0，可能发生但请检查是否正确")
	}
}
