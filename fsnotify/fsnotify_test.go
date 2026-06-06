// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux
// +build linux

package fsnotify

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// newEvent 构造一个带指定 mask 与名称的 FileEvent，便于测试。
func newEvent(name string, mask uint32) *FileEvent {
	return &FileEvent{Name: name, mask: mask}
}

// TestFlagConstants 校验 FSN 标志位的取值与 FSN_ALL 组合关系。
func TestFlagConstants(t *testing.T) {
	if FSN_CREATE != 1 || FSN_MODIFY != 2 || FSN_DELETE != 4 || FSN_RENAME != 8 {
		t.Fatalf("FSN 标志位取值发生变化: CREATE=%d MODIFY=%d DELETE=%d RENAME=%d",
			FSN_CREATE, FSN_MODIFY, FSN_DELETE, FSN_RENAME)
	}

	want := uint32(FSN_CREATE | FSN_MODIFY | FSN_DELETE | FSN_RENAME)
	if FSN_ALL != want {
		t.Fatalf("FSN_ALL 应为 %d, 实际为 %d", want, FSN_ALL)
	}

	// 各标志位互不重叠（两两按位与为 0）。
	flags := []uint32{FSN_CREATE, FSN_MODIFY, FSN_DELETE, FSN_RENAME}
	for i := 0; i < len(flags); i++ {
		for j := i + 1; j < len(flags); j++ {
			if flags[i]&flags[j] != 0 {
				t.Errorf("标志位 %d 与 %d 发生重叠", flags[i], flags[j])
			}
		}
	}
}

// TestFileEventString 校验 String() 在不同事件 mask 下的格式化输出。
func TestFileEventString(t *testing.T) {
	tests := []struct {
		name string
		ev   *FileEvent
		want string
	}{
		{"create", newEvent("a.txt", sys_IN_CREATE), `"a.txt": CREATE`},
		{"moved_to 视为 create", newEvent("a.txt", sys_IN_MOVED_TO), `"a.txt": CREATE`},
		{"delete", newEvent("a.txt", sys_IN_DELETE), `"a.txt": DELETE`},
		{"modify", newEvent("a.txt", sys_IN_MODIFY), `"a.txt": MODIFY`},
		{"rename", newEvent("a.txt", sys_IN_MOVE_SELF), `"a.txt": RENAME`},
		// ATTRIB 同时满足 IsModify 与 IsAttrib。
		{"attrib", newEvent("a.txt", sys_IN_ATTRIB), `"a.txt": MODIFY|ATTRIB`},
		// MOVED_FROM 同时满足 IsDelete 与 IsRename。
		{"moved_from", newEvent("a.txt", sys_IN_MOVED_FROM), `"a.txt": DELETE|RENAME`},
		{"无事件", newEvent("a.txt", 0), `"a.txt": `},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ev.String(); got != tt.want {
				t.Errorf("String() = %q, 期望 %q", got, tt.want)
			}
		})
	}
}

// drainPurge 在独立 goroutine 中运行 purgeEvents，发送给定事件后关闭
// internalEvent，并收集所有经过过滤后送达 Event 通道的事件。
// 当 Event 通道关闭（purgeEvents 返回）后才返回，因此返回后访问
// w.fsnFlags 不存在数据竞争。
func drainPurge(w *Watcher, events []*FileEvent) []*FileEvent {
	done := make(chan []*FileEvent)
	go func() {
		var got []*FileEvent
		for ev := range w.Event {
			got = append(got, ev)
		}
		done <- got
	}()

	go w.purgeEvents()

	for _, ev := range events {
		w.internalEvent <- ev
	}
	close(w.internalEvent)

	return <-done
}

// TestPurgeEventsFilter 校验 purgeEvents 仅放行已注册标志位匹配的事件。
func TestPurgeEventsFilter(t *testing.T) {
	w := &Watcher{
		fsnFlags:      make(map[string]uint32),
		internalEvent: make(chan *FileEvent),
		Event:         make(chan *FileEvent),
	}
	// 只关心 CREATE 与 MODIFY。
	w.fsnFlags["file"] = FSN_CREATE | FSN_MODIFY

	got := drainPurge(w, []*FileEvent{
		newEvent("file", sys_IN_CREATE),    // 放行
		newEvent("file", sys_IN_MODIFY),    // 放行
		newEvent("file", sys_IN_MOVE_SELF), // 过滤（未关注 RENAME）
		newEvent("other", sys_IN_CREATE),   // 过滤（未注册路径，标志位为 0）
	})

	if len(got) != 2 {
		t.Fatalf("期望放行 2 个事件, 实际 %d 个: %v", len(got), got)
	}
	if !got[0].IsCreate() || !got[1].IsModify() {
		t.Errorf("放行的事件类型不符: %v", got)
	}
}

// TestPurgeEventsDeleteCleanup 校验 DELETE 事件放行后会清理 fsnFlags 中的条目。
func TestPurgeEventsDeleteCleanup(t *testing.T) {
	w := &Watcher{
		fsnFlags:      make(map[string]uint32),
		internalEvent: make(chan *FileEvent),
		Event:         make(chan *FileEvent),
	}
	w.fsnFlags["file"] = FSN_ALL

	got := drainPurge(w, []*FileEvent{
		newEvent("file", sys_IN_DELETE),
	})

	if len(got) != 1 || !got[0].IsDelete() {
		t.Fatalf("期望放行 1 个 DELETE 事件, 实际: %v", got)
	}
	if _, ok := w.fsnFlags["file"]; ok {
		t.Errorf("DELETE 后 fsnFlags 仍残留 \"file\" 条目")
	}
}

// TestPurgeEventsClosesEvent 校验 internalEvent 关闭后 Event 通道随之关闭。
func TestPurgeEventsClosesEvent(t *testing.T) {
	w := &Watcher{
		fsnFlags:      make(map[string]uint32),
		internalEvent: make(chan *FileEvent),
		Event:         make(chan *FileEvent),
	}

	go w.purgeEvents()
	close(w.internalEvent)

	select {
	case _, ok := <-w.Event:
		if ok {
			t.Fatal("未发送事件却收到了数据")
		}
	case <-time.After(time.Second):
		t.Fatal("internalEvent 关闭后 Event 通道未关闭")
	}
}

// TestWatchAndRemove 校验 Watch/WatchFlags/RemoveWatch 对 fsnFlags 与底层
// inotify 监视集的维护，使用真实的 inotify 实例。
func TestWatchAndRemove(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher 失败: %v", err)
	}
	defer w.Close()

	dir := t.TempDir()

	// Watch 默认使用 FSN_ALL。
	if err := w.Watch(dir); err != nil {
		t.Fatalf("Watch 失败: %v", err)
	}
	w.fsnmut.Lock()
	flags := w.fsnFlags[dir]
	w.fsnmut.Unlock()
	if flags != FSN_ALL {
		t.Errorf("Watch 后标志位应为 FSN_ALL(%d), 实际 %d", uint32(FSN_ALL), flags)
	}

	// WatchFlags 覆盖为指定标志位。
	if err := w.WatchFlags(dir, FSN_CREATE); err != nil {
		t.Fatalf("WatchFlags 失败: %v", err)
	}
	w.fsnmut.Lock()
	flags = w.fsnFlags[dir]
	w.fsnmut.Unlock()
	if flags != FSN_CREATE {
		t.Errorf("WatchFlags 后标志位应为 FSN_CREATE(%d), 实际 %d", FSN_CREATE, flags)
	}

	// RemoveWatch 移除条目。
	if err := w.RemoveWatch(dir); err != nil {
		t.Fatalf("RemoveWatch 失败: %v", err)
	}
	w.fsnmut.Lock()
	_, ok := w.fsnFlags[dir]
	w.fsnmut.Unlock()
	if ok {
		t.Errorf("RemoveWatch 后 fsnFlags 仍残留条目")
	}
}

// TestRemoveWatchNonExistent 校验移除未注册路径时返回错误。
func TestRemoveWatchNonExistent(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher 失败: %v", err)
	}
	defer w.Close()

	if err := w.RemoveWatch("/path/never/watched"); err == nil {
		t.Error("移除未注册路径应返回错误, 实际返回 nil")
	}
}

// TestWatchReceiveCreateEvent 端到端校验：在被监视目录下新建文件，
// 应能从 Event 通道收到 CREATE 事件。
func TestWatchReceiveCreateEvent(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher 失败: %v", err)
	}
	defer w.Close()

	dir := t.TempDir()
	if err := w.Watch(dir); err != nil {
		t.Fatalf("Watch 失败: %v", err)
	}

	target := filepath.Join(dir, "new.txt")
	f, err := os.Create(target)
	if err != nil {
		t.Fatalf("创建文件失败: %v", err)
	}
	f.Close()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case ev := <-w.Event:
			if ev.Name == target && ev.IsCreate() {
				return // 成功收到目标 CREATE 事件。
			}
		case err := <-w.Error:
			t.Fatalf("监视过程中出错: %v", err)
		case <-deadline:
			t.Fatal("超时: 未收到 CREATE 事件")
		}
	}
}
