package logs

import (
	"testing"
)

func TestLogConstants(t *testing.T) {
	if LevelEmergency != 0 {
		t.Errorf("LevelEmergency = %d, want 0", LevelEmergency)
	}
	if LevelAlert != 1 {
		t.Errorf("LevelAlert = %d, want 1", LevelAlert)
	}
	if LevelCritical != 2 {
		t.Errorf("LevelCritical = %d, want 2", LevelCritical)
	}
	if LevelError != 3 {
		t.Errorf("LevelError = %d, want 3", LevelError)
	}
	if LevelWarning != 4 {
		t.Errorf("LevelWarning = %d, want 4", LevelWarning)
	}
	if LevelNotice != 5 {
		t.Errorf("LevelNotice = %d, want 5", LevelNotice)
	}
	if LevelInfo != 6 {
		t.Errorf("LevelInfo = %d, want 6", LevelInfo)
	}
	if LevelDebug != 7 {
		t.Errorf("LevelDebug = %d, want 7", LevelDebug)
	}
	if LevelPrint != 8 {
		t.Errorf("LevelPrint = %d, want 8", LevelPrint)
	}
}

func TestAdapterConstants(t *testing.T) {
	adapters := map[string]string{
		"console":   AdapterConsole,
		"file":      AdapterFile,
		"multifile": AdapterMultiFile,
		"smtp":      AdapterMail,
		"conn":      AdapterConn,
		"es":        AdapterEs,
		"jianliao":  AdapterJianLiao,
		"slack":     AdapterSlack,
	}
	for expected, actual := range adapters {
		if actual != expected {
			t.Errorf("adapter constant = %s, want %s", actual, expected)
		}
	}
}

func TestILoggerInterfaceConsole(t *testing.T) {
	logger := NewConsole()
	logger.Init(`{"level":7}`)
	_ = logger.GetLevel()
	logger.SetLevel(LevelDebug)
	logger.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, GetNow(), "test")
	logger.Flush()
	logger.Destroy()
}

func TestILoggerInterfaceConn(t *testing.T) {
	logger := NewConn()
	logger.Init("") // No addr, so connect fails but Init handles it
	_ = logger.GetLevel()
	logger.SetLevel(LevelDebug)
	logger.Flush()
	logger.Destroy()
}

func TestILoggerInterfaceFile(t *testing.T) {
	// File adapter needs a filename, so we skip direct interface test here
	// and test it in file_test.go
}

func TestILoggerInterfaceSMTP(t *testing.T) {
	logger := newSMTPWriter()
	logger.Init(`{"username":"u","password":"p","host":"h:25","level":7}`)
	_ = logger.GetLevel()
	logger.SetLevel(LevelDebug)
	logger.Flush()
	logger.Destroy()
}

func TestLevelPrefixArray(t *testing.T) {
	if len(levelPrefix) != LevelDebug+2 {
		t.Errorf("levelPrefix length = %d, want %d", len(levelPrefix), LevelDebug+2)
	}
}

func TestLevelNameArray(t *testing.T) {
	if len(levelName) != LevelDebug+2 {
		t.Errorf("levelName length = %d, want %d", len(levelName), LevelDebug+2)
	}
}

func TestTLoggerStruct(t *testing.T) {
	bl := NewLogger()
	if bl.funcCallDepth != 4 {
		t.Errorf("funcCallDepth = %d, want 4", bl.funcCallDepth)
	}
	if !bl.init_flag {
		t.Error("init_flag should be true")
	}
	if bl.Async_flag {
		t.Error("Async_flag should be false initially")
	}
	if bl.msgChanLen != int64(defAsyncMsgLen) {
		t.Errorf("msgChanLen = %d, want %d", bl.msgChanLen, int64(defAsyncMsgLen))
	}
}
