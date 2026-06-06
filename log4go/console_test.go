package logs

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestConsoleInit(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	err := cw.Init("")
	if err != nil {
		t.Fatal(err)
	}
	cw.Destroy()
}

func TestConsoleInitWithLevel(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	err := cw.Init(`{"level":3}`)
	if err != nil {
		t.Fatal(err)
	}
	if cw.Level != 3 {
		t.Errorf("Level = %d, want 3", cw.Level)
	}
	cw.Destroy()
}

func TestConsoleInitBadJSON(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	err := cw.Init("{bad json")
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestConsoleWriteMsgFiltered(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init(`{"level":3}`) // LevelError
	cw.ColorFlag = false

	// Info > Error, so should be filtered
	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "filtered")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConsoleWriteMsgPrint(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init(`{"level":3}`)
	cw.ColorFlag = false

	// Print level should always pass through even when level is high
	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelPrint, time.Now(), "print msg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConsoleWriteMsgWithNewline(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init("")
	cw.ColorFlag = false

	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "msg\n")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConsoleWriteMsgLogShort(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init(`{"level":7,"log_short":true}`)
	cw.ColorFlag = false

	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "short msg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConsoleWriteMsgWithColor(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init(`{"level":7,"color":true}`)

	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelError, time.Now(), "colored msg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConsoleDestroy(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init("")
	cw.Destroy() // Should not panic
}

func TestConsoleFlush(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init("")
	cw.Flush() // Should not panic
	cw.Destroy()
}

func TestConsoleSetGetLevel(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init("")
	cw.SetLevel(LevelError)
	if cw.GetLevel() != LevelError {
		t.Errorf("GetLevel = %d, want %d", cw.GetLevel(), LevelError)
	}
	cw.Destroy()
}

func TestConsoleStderr(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	err := cw.Init(`{"level":7,"stderr":true}`)
	if err != nil {
		t.Fatal(err)
	}
	cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "stderr msg")
	cw.Destroy()
}

func TestNewBrush(t *testing.T) {
	brush := newBrush("1;31")
	result := brush("test message")
	if !strings.Contains(result, "test message") {
		t.Errorf("brush result doesn't contain message: %s", result)
	}
	if !strings.HasPrefix(result, "\033[") {
		t.Errorf("brush result doesn't have ANSI prefix: %s", result)
	}
	if !strings.Contains(result, "\033[0m") {
		t.Errorf("brush result doesn't have reset: %s", result)
	}
}

func TestNewBrushWithNewline(t *testing.T) {
	brush := newBrush("1;31")
	result := brush("test message\n")
	// The newline should be stripped from the message and re-added after reset
	if !strings.Contains(result, "test message") {
		t.Errorf("brush result doesn't contain message: %s", result)
	}
	if !strings.HasSuffix(result, "\n") {
		t.Errorf("brush result should end with newline: %q", result)
	}
}

func TestConsoleAllLevels(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init(`{"level":7,"color":false}`)

	levels := []int{LevelEmergency, LevelAlert, LevelCritical, LevelError, LevelWarning, LevelNotice, LevelInfo, LevelDebug}
	for _, level := range levels {
		err := cw.WriteMsg("test.go", 10, 4, "TestFunc", level, time.Now(), "test msg")
		if err != nil {
			t.Errorf("WriteMsg at level %d returned error: %v", level, err)
		}
	}
	cw.Destroy()
}

func TestConsoleWriteMsgPrintNoPrefix(t *testing.T) {
	// Capture output by using a buffer-based logWriter
	cw := NewConsole().(*consoleWriter)
	cw.Init(`{"level":7,"color":false}`)

	var buf bytes.Buffer
	cw.lgwi = newLogWriter(&buf)

	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelPrint, time.Now(), "raw print")
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "raw print") {
		t.Errorf("output doesn't contain message: %s", output)
	}
	// Print level should not have level prefix like [I]>
	if strings.Contains(output, "[P]>") {
		t.Errorf("print level should not have prefix: %s", output)
	}
	cw.Destroy()
}

func TestConsoleWriteMsgNonPrintWithPrefix(t *testing.T) {
	cw := NewConsole().(*consoleWriter)
	cw.Init(`{"level":7,"color":false}`)

	var buf bytes.Buffer
	cw.lgwi = newLogWriter(&buf)

	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "info msg")
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "[I]>") {
		t.Errorf("info level should have [I]> prefix: %s", output)
	}
	cw.Destroy()
}
