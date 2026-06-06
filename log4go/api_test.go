package logs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- api.go global wrapper functions ---

func TestInitGLogger(t *testing.T) {
	Reset()
	lg := InitGLogger(LevelDebug)
	if lg == nil {
		t.Fatal("InitGLogger returned nil")
	}
	Reset()
}

func TestGlobalSetSync(t *testing.T) {
	Reset()
	SetLogger("console", `{"level":7}`)
	lg := SetSync(100)
	if lg == nil {
		t.Fatal("SetSync returned nil")
	}
	gLogger.Close()
	gLogger = NewLogger()
}

func TestGlobalSetLevel(t *testing.T) {
	Reset()
	SetLogger("console", `{"level":7}`)
	SetLevel(LevelDebug)
	level := GetLevel()
	if level != LevelDebug {
		t.Fatalf("expected level %d, got %d", LevelDebug, level)
	}
	SetLevel(LevelError)
	level = GetLevel()
	if level != LevelError {
		t.Fatalf("expected level %d, got %d", LevelError, level)
	}
	Reset()
}

func TestGlobalSetLevelInvalid(t *testing.T) {
	Reset()
	SetLogger("console", `{"level":7}`)
	// Out of range values should be ignored (no change)
	SetLevel(100)
	SetLevel(-5)
	Reset()
}

func TestGlobalSetFDebug(t *testing.T) {
	old := IsDebug
	SetFDebug(true)
	if !IsDebug {
		t.Fatal("IsDebug should be true")
	}
	SetFDebug(false)
	if IsDebug {
		t.Fatal("IsDebug should be false")
	}
	IsDebug = old
}

func TestGlobalSetConsole2Stderr(t *testing.T) {
	old := bstd_err
	SetConsole2Stderr(true)
	if !bstd_err {
		t.Fatal("bstd_err should be true")
	}
	SetConsole2Stderr(false)
	if bstd_err {
		t.Fatal("bstd_err should be false")
	}
	bstd_err = old
}

func TestGlobalSetLogger(t *testing.T) {
	Reset()
	err := SetLogger("console", `{"level":7}`)
	if err != nil {
		t.Fatal(err)
	}
	Reset()
}

func TestGlobalSetLoggerUnknown(t *testing.T) {
	Reset()
	err := SetLogger("nonexistent_adapter", "")
	if err == nil {
		t.Fatal("expected error for unknown adapter")
	}
	Reset()
}

func TestGlobalDelLogger(t *testing.T) {
	Reset()
	SetLogger("console", `{"level":7}`)
	err := DelLogger("console")
	if err != nil {
		t.Fatal(err)
	}
	Reset()
}

func TestGlobalDelLoggerUnknown(t *testing.T) {
	Reset()
	err := DelLogger("nonexistent_adapter")
	if err == nil {
		t.Fatal("expected error for unknown adapter")
	}
	Reset()
}

func TestGlobalLogFunctions(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_test_api")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "global_test.log")
	Reset()
	SetLogger("file", fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))

	Emergency("emergency msg")
	Alert("alert msg")
	Critical("critical msg")
	Error("error msg")
	Warning("warning msg")
	Notice("notice msg")
	Info("info msg")
	Debug("debug msg")
	Print("print msg")
	Begin()
	End()

	time.Sleep(100 * time.Millisecond)
	Flush()

	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "emergency msg") {
		t.Fatal("missing emergency msg in log")
	}
	if !strings.Contains(content, "debug msg") {
		t.Fatal("missing debug msg in log")
	}
	Reset()
}

func TestGlobalGetLastLogTime(t *testing.T) {
	Reset()
	SetLogger("console", `{"level":7}`)
	Info("test")
	tm := GetLastLogTime()
	if tm.IsZero() {
		t.Fatal("GetLastLogTime should not be zero after logging")
	}
	Reset()
}

func TestGlobalSetLogFuncCallDepth(t *testing.T) {
	Reset()
	SetLogger("console", `{"level":7}`)
	SetLogFuncCallDepth(5)
	if gLogger.funcCallDepth != 5 {
		t.Fatal("funcCallDepth should be 5")
	}
	SetLogFuncCallDepth(4)
	Reset()
}

func TestGlobalGetLogName(t *testing.T) {
	name := GetLogName()
	_ = name // just verify it doesn't panic
}

func TestGlobalGetLevelNoAdapters(t *testing.T) {
	Reset()
	level := GetLevel()
	if level != -1 {
		t.Errorf("GetLevel with no adapters should return -1, got %d", level)
	}
}
