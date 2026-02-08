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
		t.Error("Generated UUIDs should be unique")
	}
	
	// Check Time
	// Time is encoded in first 4 bytes.
	// Allow 1-2 second difference
	now := time.Now().Unix()
	uuidTime := uuid1.GetTime().Unix()
	
	if uuidTime < now-5 || uuidTime > now+5 {
		t.Errorf("UUID time seems wrong. Got %v, Now %v", uuidTime, now)
	}
}

func TestUUIDMarshalUnmarshal(t *testing.T) {
	fmt.Println("=== 开始测试: UUID 序列化 (TestUUIDMarshalUnmarshal) ===")
	uuid := NewUUID()
	str := uuid.String()
	
	uuid2, err := UUIDFromString(str)
	if err != nil {
		t.Fatalf("UUIDFromString failed: %v", err)
	}
	
	if uuid.String() != uuid2.String() {
		t.Errorf("Roundtrip failed. Original %s, Parsed %s", uuid.String(), uuid2.String())
	}
	
	if *uuid != *uuid2 {
		t.Error("UUID bytes mismatch")
	}
}

func TestUUIDInvalid(t *testing.T) {
	// Invalid length
	_, err := UUIDFromString("short")
	if err == nil {
		t.Error("Should fail for short string")
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
		t.Error("Should fail for invalid characters")
	}
}

func TestUUIDComponents(t *testing.T) {
	fmt.Println("=== 开始测试: UUID 组件解析 (TestUUIDComponents) ===")
	uuid := NewUUID()
	
	// Machine ID (3 bytes)
	machine := uuid.GetMachine()
	if len(machine) != 3 {
		t.Errorf("Machine ID length wrong. Got %d", len(machine))
	}
	
	// PID (2 bytes)
	pid := uuid.GetPID()
	if pid == 0 {
		t.Log("PID is 0, possible but check if correct")
	}
}
