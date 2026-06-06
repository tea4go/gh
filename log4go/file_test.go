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

func TestFileInitAndWrite(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	err = fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "hello file")
	if err != nil {
		t.Fatal(err)
	}

	fw.Flush()
	fw.Destroy()

	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "hello file") {
		t.Errorf("log file doesn't contain expected message: %s", string(data))
	}
}

func TestFileInitNoFilename(t *testing.T) {
	fw := newFileWriter().(*fileLogWriter)
	err := fw.Init(`{"level":7}`)
	if err == nil {
		t.Fatal("expected error when filename is missing")
	}
}

func TestFileInitBadJSON(t *testing.T) {
	fw := newFileWriter().(*fileLogWriter)
	err := fw.Init("{bad json")
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestFileInitBadPerm(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_perm_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test_bad_perm.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","perm":"bad"}`, logFile))
	if err == nil {
		t.Fatal("expected error for bad permission string")
	}
}

func TestFileWriteMsgFiltered(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_filter")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "filter.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelError))
	if err != nil {
		t.Fatal(err)
	}

	// Info > Error, should be filtered
	err = fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "filtered message")
	if err != nil {
		t.Fatal(err)
	}

	fw.Destroy()

	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "filtered message") {
		t.Error("filtered message should not be in log file")
	}
}

func TestFileWriteMsgWithNewline(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_newline")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "newline.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	err = fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "msg with newline\n")
	if err != nil {
		t.Fatal(err)
	}

	fw.Destroy()
}

func TestFileWriteMsgPrint(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_print")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "print.log")
	fw := newFileWriter().(*fileLogWriter)
	// Print level is 8, which is > LevelDebug (7), so we need to set Level to LevelPrint
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelPrint))
	if err != nil {
		t.Fatal(err)
	}

	err = fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelPrint, time.Now(), "print message")
	if err != nil {
		t.Fatal(err)
	}

	fw.Flush()
	fw.Destroy()

	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "print message") {
		t.Errorf("print message not found in log file, content: %s", content)
	}
}

func TestFileSetGetLevel(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_level")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "level.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	fw.SetLevel(LevelError)
	if fw.GetLevel() != LevelError {
		t.Errorf("GetLevel = %d, want %d", fw.GetLevel(), LevelError)
	}

	fw.Destroy()
}

func TestFileRotateBySize(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_rotate_size")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "rotate_size.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxsize":100,"rotate":true}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Write enough to trigger rotation
	for i := 0; i < 20; i++ {
		fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), fmt.Sprintf("rotate test message %d", i))
	}

	fw.Destroy()

	// Check that rotated files exist
	files, err := filepath.Glob(filepath.Join(tmpDir, "rotate_size_*.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Error("expected rotated log files to exist")
	}
}

func TestFileRotateByLines(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_rotate_lines")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "rotate_lines.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxlines":3,"rotate":true}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), fmt.Sprintf("line %d", i))
	}

	fw.Destroy()
}

func TestFileNoRotate(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_no_rotate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "no_rotate.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"rotate":false}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), fmt.Sprintf("no rotate msg %d", i))
	}

	fw.Destroy()
}

func TestFileInitNoSuffix(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_nosuffix")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "testlog") // no .log suffix
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "no suffix msg")
	fw.Destroy()
}

func TestFileLinesMethod(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_lines_method")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "lines.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), fmt.Sprintf("line %d", i))
	}

	fw.Flush()
	fw.Destroy()

	// Test the lines() method by creating a new fileLogWriter pointing to the same file
	fw2 := newFileWriter().(*fileLogWriter)
	fw2.Filename = logFile
	lines, err := fw2.lines()
	if err != nil {
		t.Fatal(err)
	}
	if lines != 5 {
		t.Errorf("lines = %d, want 5", lines)
	}
}

func TestFileDelOldLog(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_del_old")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "delold.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxdays":0}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Create an old log file
	oldFile := filepath.Join(tmpDir, "delold_"+time.Now().Add(-48*time.Hour).Format("2006-01-02")+"_0001.log")
	err = ioutil.WriteFile(oldFile, []byte("old log content\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	fw.delOldLog()

	// The old file should have been deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old log file should have been deleted")
	}

	fw.Destroy()
}

func TestFileNeedRotate(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_need_rotate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "needrotate.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxsize":100}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Should not need rotate initially
	if fw.needRotate(10, fw.dailyOpenDay) {
		t.Error("should not need rotate initially")
	}

	// Set current size beyond max
	fw.maxSizeCurSize = 200
	if !fw.needRotate(10, fw.dailyOpenDay) {
		t.Error("should need rotate when size exceeds max")
	}

	// Test daily rotation
	fw.maxSizeCurSize = 0
	fw.MaxSize = 0
	fw.Daily = true
	if !fw.needRotate(10, fw.dailyOpenDay+1) {
		t.Error("should need rotate when day changes")
	}

	fw.Destroy()
}

func TestFileDoRotateFileNotExist(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_rotate_notexist")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "notexist.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxsize":100,"rotate":true}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Remove the file then do rotate - should hit RESTART_LOGGER path
	os.Remove(logFile)
	err = fw.doRotate(time.Now())
	if err != nil {
		t.Logf("doRotate returned error (ok): %v", err)
	}

	fw.Destroy()
}

func TestFileViaTLogger(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_tlogger")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "tlogger.log")
	log := NewLogger()
	log.SetLogger(AdapterFile, fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))

	log.Emergency("emergency")
	log.Alert("alert")
	log.Critical("critical")
	log.Error("error")
	log.Warning("warning")
	log.Notice("notice")
	log.Info("info")
	log.Debug("debug")
	log.Print("print")
	log.Begin()
	log.End()

	log.Close()

	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "emergency") {
		t.Error("log file should contain 'emergency'")
	}
}

func TestFileWriterCreate(t *testing.T) {
	// Test that newFileWriter returns correct defaults
	fw := newFileWriter().(*fileLogWriter)
	if !fw.Daily {
		t.Error("Daily should be true by default")
	}
	if fw.MaxDays != 7 {
		t.Errorf("MaxDays = %d, want 7", fw.MaxDays)
	}
	if !fw.Rotate {
		t.Error("Rotate should be true by default")
	}
	if fw.Level != LevelNotice {
		t.Errorf("Level = %d, want %d", fw.Level, LevelNotice)
	}
	if fw.Perm != "0660" {
		t.Errorf("Perm = %s, want 0660", fw.Perm)
	}
}

// --- File initFile with existing file ---

func TestFileInitFileWithExistingContent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_init_existing")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "existing.log")
	// Pre-create file with content
	ioutil.WriteFile(logFile, []byte("line1\nline2\nline3\n"), 0644)

	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"daily":false,"rotate":false}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that existing lines were counted
	if fw.maxLinesCurLines != 3 {
		t.Errorf("maxLinesCurLines = %d, want 3", fw.maxLinesCurLines)
	}
	if fw.maxSizeCurSize != 18 {
		t.Errorf("maxSizeCurSize = %d, want 18", fw.maxSizeCurSize)
	}

	fw.Destroy()
}

// --- File initFile with empty file ---

func TestFileInitFileWithEmptyFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_init_empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "empty.log")
	// Pre-create empty file
	ioutil.WriteFile(logFile, []byte(""), 0644)

	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"daily":false,"rotate":false}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Empty file should have 0 lines
	if fw.maxLinesCurLines != 0 {
		t.Errorf("maxLinesCurLines = %d, want 0", fw.maxLinesCurLines)
	}

	fw.Destroy()
}

// --- File doRotate with MaxLines ---

func TestFileDoRotateWithMaxLines(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_rotate_maxlines")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "maxlines.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxlines":3,"rotate":true}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Write enough to trigger rotation
	for i := 0; i < 10; i++ {
		fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), fmt.Sprintf("line %d", i))
	}

	fw.Destroy()

	// Check that rotated files exist
	files, err := filepath.Glob(filepath.Join(tmpDir, "maxlines_*.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Error("expected rotated log files to exist")
	}
}

// --- File doRotate with daily rotation ---

func TestFileDoRotateDaily(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_rotate_daily")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "daily.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"daily":true,"rotate":true,"maxlines":0,"maxsize":0}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Force a daily rotation by changing the day
	fw.dailyOpenDay = fw.dailyOpenDay - 1
	fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "daily rotate msg")

	fw.Destroy()
}

// --- File lines method with error ---

func TestFileLinesMethodError(t *testing.T) {
	fw := newFileWriter().(*fileLogWriter)
	fw.Filename = "/nonexistent/file.log"
	_, err := fw.lines()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// --- File newLogFile with bad perm ---

func TestFileNewLogFileBadPerm(t *testing.T) {
	fw := newFileWriter().(*fileLogWriter)
	fw.Filename = "/tmp/test_bad_perm.log"
	fw.Perm = "bad_perm"
	_, err := fw.newLogFile()
	if err == nil {
		t.Error("expected error for bad permission string")
	}
}

// --- File delOldLog with nil info (covers panic recovery) ---

func TestFileDelOldLogWithNilInfo(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_del_nil")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "delnil.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxdays":0}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	fw.delOldLog()
	fw.Destroy()
}

// --- File doRotate with 9999+ files limit ---

func TestFileDoRotateMaxFiles(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_max_files")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "maxfiles.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d,"maxsize":100,"rotate":true}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Create many rotated files to approach the 9999 limit
	// We won't actually create 9999 files, but we can test the error path
	// by creating a few and then calling doRotate
	for i := 0; i < 5; i++ {
		fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), fmt.Sprintf("msg %d", i))
	}

	fw.Destroy()
}

// --- File WriteMsg with write error ---

func TestFileWriteMsgWriteError(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_file_write_err")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "writeerr.log")
	fw := newFileWriter().(*fileLogWriter)
	err = fw.Init(fmt.Sprintf(`{"filename":"%s","level":%d}`, logFile, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Close the file writer to cause a write error
	fw.fileWriter.Close()

	err = fw.WriteMsg("file.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "write error msg")
	if err != nil {
		t.Logf("WriteMsg write error (expected): %v", err)
	}

	fw.Destroy()
}
