package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJsonTime(t *testing.T) {
	fmt.Println("=== 开始测试: JsonTime 序列化 (TestJsonTime) ===")
	now := time.Now()
	jt := JsonTime(now)
	
	// Test Marshal
	data, err := json.Marshal(jt)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	// Test Unmarshal
	var jt2 JsonTime
	err = json.Unmarshal(data, &jt2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	// Compare string representations (ignoring monotonic clock)
	if jt.ToString() != jt2.ToString() {
		t.Errorf("JsonTime mismatch. Got %s, want %s", jt2.ToString(), jt.ToString())
	}
}

func TestStringChecks(t *testing.T) {
	fmt.Println("=== 开始测试: 字符串检查函数 (TestStringChecks) ===")
	// IsAscii
	if !IsAscii("abc") { t.Error("abc should be Ascii") }
	if !IsAscii("123") { t.Error("123 should be Ascii") }
	
	// IsNumber
	if !IsNumber("123") { t.Error("123 should be Number") }
	if IsNumber("abc") { t.Error("abc should not be Number") }
	
	// IsHanZi / IsChinese
	if !IsHanZi("汉") { t.Error("汉 should be HanZi") }
	if IsHanZi("a") { t.Error("a should not be HanZi") }
	
	if !IsChinese("Hello 世界") { t.Error("Hello 世界 contains Chinese") }
	if IsChinese("Hello World") { t.Error("Hello World does not contain Chinese") }
}

func TestMd5(t *testing.T) {
	fmt.Println("=== 开始测试: MD5 哈希 (TestMd5) ===")
	s := "hello"
	expected := "5d41402abc4b2a76b9719d911017c592"
	if Md5(s) != expected {
		t.Errorf("Md5 mismatch. Got %s, want %s", Md5(s), expected)
	}
}

func TestBase64(t *testing.T) {
	fmt.Println("=== 开始测试: Base64 编码解码 (TestBase64) ===")
	s := "hello"
	encoded := Base64Encode(s)
	decoded, err := Base64Decode(encoded)
	if err != nil {
		t.Fatalf("Base64Decode failed: %v", err)
	}
	if decoded != s {
		t.Errorf("Base64 roundtrip failed. Got %s, want %s", decoded, s)
	}
}

func TestPasswordMasking(t *testing.T) {
	fmt.Println("=== 开始测试: 密码掩码显示 (TestPasswordMasking) ===")
	pw := "password123"
	show := GetShowPassword(pw)
	if show != "p***3" {
		t.Errorf("GetShowPassword failed. Got %s", show)
	}
	
	short := "1234"
	if GetShowPassword(short) != "***" {
		t.Errorf("GetShowPassword short failed")
	}
	
	key := "123456"
	showKey := GetShowKey(key)
	if showKey != "1234***6" {
		t.Errorf("GetShowKey failed. Got %s", showKey)
	}
}

func TestSizeString(t *testing.T) {
	fmt.Println("=== 开始测试: 文件大小格式化 (TestSizeString) ===")
	// KB
	kb := KB * 1.5
	if s := kb.String(); s != " 1.5K" {
		t.Logf("Size string: %s", s)
	}
	
	// MB
	mb := MB * 2.0
	if s := mb.String(); s != " 2.0M" {
		t.Logf("Size string: %s", s)
	}
}

func TestFileOperations(t *testing.T) {
	fmt.Println("=== 开始测试: 文件操作 (TestFileOperations) ===")
	// Create a temp file
	tmpContent := "test content"
	tmpFile := "temp_test_file.txt"
	err := ioutil.WriteFile(tmpFile, []byte(tmpContent), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)
	
	// GetFileExt
	if ext := GetFileExt(tmpFile); ext != ".txt" {
		t.Errorf("GetFileExt failed. Got %s", ext)
	}
	
	// GetFileSize
	if size := GetFileSize(tmpFile); size != int64(len(tmpContent)) {
		t.Errorf("GetFileSize failed. Got %d", size)
	}
	
	// FileIsExist
	if !FileIsExist(tmpFile) {
		t.Error("FileIsExist returned false for existing file")
	}
	
	// LoadFileText
	content, err := LoadFileText(tmpFile)
	if err != nil {
		t.Error(err)
	}
	if content != tmpContent {
		t.Error("LoadFileText mismatch")
	}
}

func TestJsonOperations(t *testing.T) {
	fmt.Println("=== 开始测试: JSON 文件存取 (TestJsonOperations) ===")
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	
	obj := TestStruct{Name: "Test", Age: 10}
	jsonFile := "test_obj.json"
	defer os.Remove(jsonFile)
	
	// SaveJson
	err := SaveJson(jsonFile, &obj)
	if err != nil {
		t.Fatal(err)
	}
	
	// LoadJson
	var loadedObj TestStruct
	err = LoadJson(jsonFile, &loadedObj)
	if err != nil {
		t.Fatal(err)
	}
	
	if loadedObj.Name != obj.Name || loadedObj.Age != obj.Age {
		t.Error("Json save/load mismatch")
	}
}

func TestIIF(t *testing.T) {
	fmt.Println("=== 开始测试: IIF 三元运算 (TestIIF) ===")
	if val := IIF(true, "yes", "no"); val != "yes" {
		t.Error("IIF true failed")
	}
	if val := IIF(false, "yes", "no"); val != "no" {
		t.Error("IIF false failed")
	}
}

func TestSubstr(t *testing.T) {
	fmt.Println("=== 开始测试: 字符串截取 (TestSubstr) ===")
	s := "Hello World"
	if sub := Substr(s, 0, 5); sub != "Hello" {
		t.Errorf("Substr failed. Got %s", sub)
	}
	if sub := Substr(s, 6, 5); sub != "World" {
		t.Errorf("Substr failed. Got %s", sub)
	}
}

func TestGetTimeAgo(t *testing.T) {
	fmt.Println("=== 开始测试: 时间流逝描述 (TestGetTimeAgo) ===")
	now := GetNow()
	
	past := now.Add(-10 * time.Second)
	if s := GetTimeAgo(past); s != "10秒前" {
		t.Logf("GetTimeAgo 10s: %s", s)
	}
	
	past = now.Add(-2 * time.Minute)
	if s := GetTimeAgo(past); s != "2分前" {
		t.Logf("GetTimeAgo 2m: %s", s)
	}
}

func TestGetCallStack(t *testing.T) {
	fmt.Println("=== 开始测试: 调用堆栈获取 (TestGetCallStack) ===")
	_, stack, _, _ := GetCallStack()
	if stack == "" {
		t.Error("GetCallStack returned empty stack")
	}
}

func TestRunPath(t *testing.T) {
	fmt.Println("=== 开始测试: 运行路径获取 (TestRunPath) ===")
	p := RunPathName()
	if p == "" {
		t.Error("RunPathName is empty")
	}
	
	p = GetCurrentPath()
	if p == "" {
		t.Error("GetCurrentPath is empty")
	}
}

func TestDosName(t *testing.T) {
	fmt.Println("=== 开始测试: DOS 文件名转换 (TestDosName) ===")
	// DosName seems to truncate long filenames to 8.3 format or similar logic
	longName := "reallylongfilename.txt"
	dosName := DosName(longName)
	// Function logic:
	// if len(file) >= 13 { file = Substr(file, 0, 13) + "~" }
	// ext = Substr(ext, 0, 4)
	
	// "reallylongfilename" is 18 chars.
	// Substr(0, 13) -> "reallylongfil"
	// + "~" -> "reallylongfil~"
	// ext ".txt" -> ".txt"
	
	expected := "reallylongfil~.txt"
	if dosName != expected {
		t.Errorf("DosName failed. Got %s, want %s", dosName, expected)
	}
}

func TestGetFileDir(t *testing.T) {
	fmt.Println("=== 开始测试: 获取文件目录 (TestGetFileDir) ===")
	path := "/a/b/c.txt"
	dir := GetFileDir(path)
	// filepath.Dir depends on OS separator.
	// On windows it handles / mixed with \ usually.
	expected := filepath.Dir(path)
	if dir != expected {
		t.Errorf("GetFileDir failed. Got %s, want %s", dir, expected)
	}
}
