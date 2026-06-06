//go:build linux
// +build linux

package fsnotify

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- GetMask tests ---

func TestGetMask_Create(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_CREATE)
	got := ev.GetMask()
	if got == "" {
		t.Error("GetMask() should not be empty for CREATE")
	}
	if !containsMask(got, "IN_CREATE") {
		t.Errorf("GetMask() = %q, should contain IN_CREATE", got)
	}
}

func TestGetMask_Delete(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_DELETE)
	got := ev.GetMask()
	if !containsMask(got, "FS_DELETE") {
		t.Errorf("GetMask() = %q, should contain FS_DELETE", got)
	}
}

func TestGetMask_Modify(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_MODIFY)
	got := ev.GetMask()
	if !containsMask(got, "IN_MODIFY") {
		t.Errorf("GetMask() = %q, should contain IN_MODIFY", got)
	}
}

func TestGetMask_Attrib(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_ATTRIB)
	got := ev.GetMask()
	if !containsMask(got, "IN_ATTRIB") {
		t.Errorf("GetMask() = %q, should contain IN_ATTRIB", got)
	}
}

func TestGetMask_MovedFrom(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_MOVED_FROM)
	got := ev.GetMask()
	if !containsMask(got, "IN_MOVED_FROM") {
		t.Errorf("GetMask() = %q, should contain IN_MOVED_FROM", got)
	}
}

func TestGetMask_MovedTo(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_MOVED_TO)
	got := ev.GetMask()
	if !containsMask(got, "IN_MOVED_TO") {
		t.Errorf("GetMask() = %q, should contain IN_MOVED_TO", got)
	}
}

func TestGetMask_MoveSelf(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_MOVE_SELF)
	got := ev.GetMask()
	if !containsMask(got, "IN_MOVE_SELF") {
		t.Errorf("GetMask() = %q, should contain IN_MOVE_SELF", got)
	}
}

func TestGetMask_Move(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_MOVE)
	got := ev.GetMask()
	if !containsMask(got, "IN_MOVE") {
		t.Errorf("GetMask() = %q, should contain IN_MOVE", got)
	}
}

func TestGetMask_Open(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_OPEN)
	got := ev.GetMask()
	if !containsMask(got, "IN_OPEN") {
		t.Errorf("GetMask() = %q, should contain IN_OPEN", got)
	}
}

func TestGetMask_DeleteSelf(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_DELETE_SELF)
	got := ev.GetMask()
	if !containsMask(got, "FS_DELETE_SELF") {
		t.Errorf("GetMask() = %q, should contain FS_DELETE_SELF", got)
	}
}

func TestGetMask_Close(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_CLOSE)
	got := ev.GetMask()
	if !containsMask(got, "IN_CLOSE") {
		t.Errorf("GetMask() = %q, should contain IN_CLOSE", got)
	}
}

func TestGetMask_CloseWrite(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_CLOSE_WRITE)
	got := ev.GetMask()
	if !containsMask(got, "IN_CLOSE_WRITE") {
		t.Errorf("GetMask() = %q, should contain IN_CLOSE_WRITE", got)
	}
}

func TestGetMask_UnknownBit(t *testing.T) {
	// Use a bit that doesn't match any known flag
	ev := newEvent("test.txt", 0x02000000)
	got := ev.GetMask()
	if !containsMask(got, "未知操作") {
		t.Errorf("GetMask() = %q, should contain 未知操作 for unknown bits", got)
	}
}

func TestGetMask_Empty(t *testing.T) {
	ev := newEvent("test.txt", 0)
	got := ev.GetMask()
	if got != "" {
		t.Errorf("GetMask() for mask=0 should be empty, got %q", got)
	}
}

func TestGetMask_Combined(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_CREATE|sys_IN_MODIFY)
	got := ev.GetMask()
	if !containsMask(got, "IN_CREATE") {
		t.Errorf("GetMask() = %q, should contain IN_CREATE", got)
	}
	if !containsMask(got, "IN_MODIFY") {
		t.Errorf("GetMask() = %q, should contain IN_MODIFY", got)
	}
}

// --- FileEvent method tests ---

func TestFileEvent_IsCreate(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_CREATE)
	if !ev.IsCreate() {
		t.Error("IN_CREATE should be IsCreate()")
	}
	ev2 := newEvent("test.txt", sys_IN_MOVED_TO)
	if !ev2.IsCreate() {
		t.Error("IN_MOVED_TO should also be IsCreate()")
	}
}

func TestFileEvent_IsDelete(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_DELETE)
	if !ev.IsDelete() {
		t.Error("IN_DELETE should be IsDelete()")
	}
	ev2 := newEvent("test.txt", sys_IN_DELETE_SELF)
	if !ev2.IsDelete() {
		t.Error("IN_DELETE_SELF should also be IsDelete()")
	}
	ev3 := newEvent("test.txt", sys_IN_MOVED_FROM)
	if !ev3.IsDelete() {
		t.Error("IN_MOVED_FROM should also be IsDelete()")
	}
}

func TestFileEvent_IsModify(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_MODIFY)
	if !ev.IsModify() {
		t.Error("IN_MODIFY should be IsModify()")
	}
	ev2 := newEvent("test.txt", sys_IN_ATTRIB)
	if !ev2.IsModify() {
		t.Error("IN_ATTRIB should also be IsModify()")
	}
}

func TestFileEvent_IsRename(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_MOVE_SELF)
	if !ev.IsRename() {
		t.Error("IN_MOVE_SELF should be IsRename()")
	}
	ev2 := newEvent("test.txt", sys_IN_MOVED_FROM)
	if !ev2.IsRename() {
		t.Error("IN_MOVED_FROM should also be IsRename()")
	}
}

func TestFileEvent_IsAttrib(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_ATTRIB)
	if !ev.IsAttrib() {
		t.Error("IN_ATTRIB should be IsAttrib()")
	}
}

func TestFileEvent_ignoreLinux_Ignored(t *testing.T) {
	ev := newEvent("test.txt", sys_IN_IGNORED)
	if !ev.ignoreLinux() {
		t.Error("IN_IGNORED should be ignored")
	}
}

func TestFileEvent_ignoreLinux_DeletedFile(t *testing.T) {
	// Create a temp file that we'll delete
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "deleteme.txt")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	ev := newEvent(tmpFile, sys_IN_MODIFY)
	// File exists, so not ignored
	if ev.ignoreLinux() {
		t.Error("MODIFY on existing file should not be ignored")
	}

	// Delete the file
	os.Remove(tmpFile)
	ev2 := newEvent(tmpFile, sys_IN_MODIFY)
	// File no longer exists and it's not DELETE or RENAME, so it should be ignored
	if !ev2.ignoreLinux() {
		t.Error("MODIFY on deleted file should be ignored")
	}
}

func TestFileEvent_ignoreLinux_DeleteNotIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "somefile.txt")
	os.Remove(tmpFile)

	ev := newEvent(tmpFile, sys_IN_DELETE)
	// DELETE events are never ignored
	if ev.ignoreLinux() {
		t.Error("DELETE event should not be ignored even for non-existent file")
	}
}

func TestFileEvent_ignoreLinux_RenameNotIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "somefile.txt")
	os.Remove(tmpFile)

	ev := newEvent(tmpFile, sys_IN_MOVED_FROM)
	// RENAME events are never ignored
	if ev.ignoreLinux() {
		t.Error("RENAME event should not be ignored even for non-existent file")
	}
}

// --- Watcher Close test ---

func TestWatcher_CloseTwice(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("First Close failed: %v", err)
	}
	// Second close should return nil
	if err := w.Close(); err != nil {
		t.Errorf("Second Close should return nil, got: %v", err)
	}
}

// --- Watch modify/delete/rename events ---

func TestWatchReceiveModifyEvent(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	defer w.Close()

	dir := t.TempDir()
	if err := w.Watch(dir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a file first
	target := filepath.Join(dir, "modify.txt")
	f, err := os.Create(target)
	if err != nil {
		t.Fatalf("Create file failed: %v", err)
	}
	f.Close()

	// Wait for CREATE event to drain
	drainEvents(w, 500*time.Millisecond)

	// Modify the file
	f, err = os.OpenFile(target, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("Open file failed: %v", err)
	}
	f.WriteString("hello")
	f.Close()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case ev := <-w.Event:
			if ev.Name == target && ev.IsModify() {
				return
			}
		case err := <-w.Error:
			t.Fatalf("Error during watch: %v", err)
		case <-deadline:
			t.Fatal("Timeout: did not receive MODIFY event")
		}
	}
}

func TestWatchReceiveDeleteEvent(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	defer w.Close()

	dir := t.TempDir()
	if err := w.Watch(dir); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create and then delete
	target := filepath.Join(dir, "delete_me.txt")
	f, err := os.Create(target)
	if err != nil {
		t.Fatalf("Create file failed: %v", err)
	}
	f.Close()

	// Wait for CREATE
	drainEvents(w, 500*time.Millisecond)

	// Delete the file
	os.Remove(target)

	deadline := time.After(5 * time.Second)
	for {
		select {
		case ev := <-w.Event:
			if ev.Name == target && ev.IsDelete() {
				return
			}
		case err := <-w.Error:
			t.Fatalf("Error during watch: %v", err)
		case <-deadline:
			t.Fatal("Timeout: did not receive DELETE event")
		}
	}
}

func TestAddWatch_ClosedWatcher(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	w.Close()

	err = w.addWatch("/tmp", sys_AGNOSTIC_EVENTS)
	if err == nil {
		t.Error("addWatch on closed watcher should return error")
	}
}

// --- Helper functions ---

func containsMask(mask, flag string) bool {
	return len(mask) > 0 && containsStr(mask, flag)
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func drainEvents(w *Watcher, timeout time.Duration) {
	deadline := time.After(timeout)
	for {
		select {
		case <-w.Event:
		case <-w.Error:
		case <-deadline:
			return
		}
	}
}
