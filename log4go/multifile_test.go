package logs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultiFileInit(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_multifile_init")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "multi.log")
	mfw := newFilesWriter().(*multiFileLogWriter)
	err = mfw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"separate":["error","warning"]}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}
	mfw.Destroy()
}

func TestMultiFileInitBadJSON(t *testing.T) {
	mfw := newFilesWriter().(*multiFileLogWriter)
	err := mfw.Init("{bad json")
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestMultiFileInitNoFilename(t *testing.T) {
	mfw := newFilesWriter().(*multiFileLogWriter)
	err := mfw.Init(`{"level":7}`)
	if err == nil {
		t.Fatal("expected error for missing filename")
	}
}

func TestMultiFileWriteMsg(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_multifile_write")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "multi_write.log")
	mfw := newFilesWriter().(*multiFileLogWriter)
	err = mfw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"separate":["error","warning","info"]}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	mfw.WriteMsg("test.go", 10, 4, "TestFunc", LevelError, GetNow(), "error message")
	mfw.WriteMsg("test.go", 10, 4, "TestFunc", LevelWarning, GetNow(), "warning message")
	mfw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, GetNow(), "info message")
	mfw.WriteMsg("test.go", 10, 4, "TestFunc", LevelDebug, GetNow(), "debug message")

	mfw.Flush()
	mfw.Destroy()

	// Check the full log file
	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "error message") {
		t.Error("full log should contain error message")
	}

	// Check the error-specific log file
	errFile := filepath.Join(tmpDir, "multi_write.error.log")
	data, err = ioutil.ReadFile(errFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "error message") {
		t.Error("error log should contain error message")
	}
}

func TestMultiFileDestroy(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_multifile_destroy")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "destroy.log")
	mfw := newFilesWriter().(*multiFileLogWriter)
	err = mfw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"separate":["error"]}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}
	mfw.Destroy() // Should not panic
}

func TestMultiFileFlush(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_multifile_flush")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "flush.log")
	mfw := newFilesWriter().(*multiFileLogWriter)
	err = mfw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"separate":["error"]}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}
	mfw.Flush() // Should not panic
	mfw.Destroy()
}

func TestMultiFileSetGetLevel(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_multifile_setlevel")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "setlevel.log")
	mfw := newFilesWriter().(*multiFileLogWriter)
	err = mfw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"separate":["emergency","error"]}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	mfw.SetLevel(LevelError)
	// GetLevel will try writers[0] which is the emergency writer (set via separate)
	// and return its level
	level := mfw.GetLevel()
	_ = level // Just verify it doesn't panic and returns a valid int
	mfw.Destroy()
}

// Note: multiFileLogWriter.GetLevel() panics on zero-initialized struct
// because writers[0] is nil. This is a bug in the source code.

func TestMultiFileWriteMsgNoFullWriter(t *testing.T) {
	mfw := &multiFileLogWriter{}
	err := mfw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, GetNow(), "no writer")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMultiFileViaTLogger(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_multifile_tlogger")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "tlogger_multi.log")
	log := NewLogger()
	log.SetLogger(AdapterMultiFile, fmt.Sprintf(`{"filename":"%s","level":%d,"separate":["emergency","alert","critical","error","warning","notice","info","debug"]}`, logFile, LevelDebug))

	log.Emergency("emergency")
	log.Alert("alert")
	log.Critical("critical")
	log.Error("error")
	log.Warning("warning")
	log.Notice("notice")
	log.Info("info")
	log.Debug("debug")

	log.Close()

	// Verify the main log file
	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "emergency") {
		t.Error("log file should contain 'emergency'")
	}
}

func TestMultiFileAllSeparateLevels(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_multifile_all_sep")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "all_sep.log")
	mfw := newFilesWriter().(*multiFileLogWriter)
	err = mfw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"separate":["emergency","alert","critical","error","warning","notice","info","debug"]}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Write to all levels
	for i := LevelEmergency; i <= LevelDebug; i++ {
		mfw.WriteMsg("test.go", 10, 4, "TestFunc", i, GetNow(), fmt.Sprintf("%s level message", levelNames[i]))
	}

	mfw.Flush()
	mfw.Destroy()

	// Verify each separate file exists
	for _, name := range levelNames {
		sepFile := filepath.Join(tmpDir, "all_sep."+name+".log")
		data, err := ioutil.ReadFile(sepFile)
		if err != nil {
			t.Errorf("failed to read %s: %v", sepFile, err)
			continue
		}
		if !strings.Contains(string(data), name+" level message") {
			t.Errorf("%s log should contain its message", name)
		}
	}
}
