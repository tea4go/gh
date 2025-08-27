// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logs

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestFilePerm(t *testing.T) {
	t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	log := NewLogger()
	//log.SetSync(100)
	// use 0666 as test perm cause the default umask is 022
	log.SetLogger("file", `{"filename":"test.log", "perm": "0666"}`)

	log.SetLevel(LevelDebug)
	testConsoleCalls(log)
	log.Close()
	time.Sleep(1 * time.Second)

	file, err := os.Stat("test.log")
	if err != nil {
		t.Fatal(err)
	}
	if file.Mode() != 0666 {
		t.Fatal("日志文件的权限与预期不相符！")
	}
	err = os.Remove("test.log")
	if err != nil {
		t.Fatal("删除测试日志文件失败，" + err.Error())
	}
}

func TestFileLineNum(t *testing.T) {
	t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	log := NewLogger()
	//log.SetSync(100)
	log.SetLogger("file", `{"filename":"test.log"}`)

	log.SetLevel(LevelDebug)
	testConsoleCalls(log)
	log.Close()

	lineNum, err := GetFileLines("test.log")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove("test.log")
	if err != nil {
		t.Fatal("删除测试日志文件失败，" + err.Error())
	}

	var expected = LevelDebug + 1
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
}

func TestFile(t *testing.T) {
	t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	log := NewLogger()
	//log.SetSync(100)
	log.SetLogger("file", fmt.Sprintf(`{"filename":"test.log","level":%d}`, LevelError))
	testConsoleCalls(log)
	log.Close()
	time.Sleep(1 * time.Second)

	lineNum, err := GetFileLines("test.log")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove("test.log")
	if err != nil {
		t.Fatal("删除测试日志文件失败，" + err.Error())
	}
	var expected = LevelError + 1
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
}

func TestFileRotate_00(t *testing.T) {
	//t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	log := NewLogger()
	log.SetLogger("file", fmt.Sprintf(`{"filename":"test2.log","level":%d,"maxsize":400}`, LevelDebug))
	time.Sleep(20 * time.Microsecond)

	testConsoleCalls(log)
	time.Sleep(20 * time.Microsecond)

	log.Close()
	time.Sleep(1 * time.Second)

	rotateName := "test2" + fmt.Sprintf("_%s_%04d.log", time.Now().Format("2006-01-02"), 1)
	b, err := exists(rotateName)
	if !b || err != nil {
		os.Remove("test2.log")
		t.Fatal("rotate not generated")
	}
	os.Remove(rotateName)
	os.Remove("test2.log")
}

func TestFileRotate_01(t *testing.T) {
	//t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	log := NewLogger()
	log.SetLogger("file", fmt.Sprintf(`{"filename":"test3.log","level":%d,"maxlines":4}`, LevelDebug))
	time.Sleep(20 * time.Microsecond)

	testConsoleCalls(log)
	time.Sleep(20 * time.Microsecond)

	log.Close()
	time.Sleep(1 * time.Second)

	rotateName := "test3" + fmt.Sprintf("_%s_%04d.log", time.Now().Format("2006-01-02"), 1)
	b, err := exists(rotateName)
	if !b || err != nil {
		os.Remove("test3.log")
		t.Fatal("rotate not generated")
	}
	os.Remove(rotateName)
	os.Remove("test3.log")
}

func TestFileRotate_02(t *testing.T) {
	//t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	fn1 := "rotate_day_02.log"
	fn2 := "rotate_day_02_" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + "_0001.log"
	testFileRotate(t, fn1, fn2)
}

func TestFileRotate_03(t *testing.T) {
	//t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	fn1 := "rotate_day_03.log"
	fn0 := "rotate_day_03_" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + "_0001.log"
	fnd, _ := os.Create(fn0)
	fn2 := "rotate_day_03_" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + "_0002.log"
	testFileRotate(t, fn1, fn2)
	fnd.Close()
	os.Remove(fn0)
}

func testFileRotate(t *testing.T, fn1, fn2 string) {
	fw := &fileLogWriter{
		Daily:   true,
		MaxDays: 7,
		Rotate:  true,
		Level:   LevelDebug,
		Perm:    "0660",
	}
	FDebug("===> 初始化日志 ......")
	fw.Init(fmt.Sprintf(`{"filename":"%v","maxdays":1}`, fn1))
	time.Sleep(20 * time.Microsecond)

	FDebug("===> 调整日志的当前天 ......")
	fw.dailyOpenTime = time.Now().Add(-24 * time.Hour)
	fw.dailyOpenDay = fw.dailyOpenTime.Day()
	FDebug("= 当前日志所在天（%d号）", fw.dailyOpenDay)
	FDebug("= 预计 1 秒后会自动切换文件(写日志时会自动触发切换)")
	time.Sleep(1 * time.Second)

	FDebug("===> 准备写入日志 ......")
	fw.WriteMsg("file.go", 10, 4, "main()", LevelDebug, time.Now(), "this is a msg for test")

	time.Sleep(1 * time.Second)
	FDebug("===> 准备释放日志 ......")
	fw.Destroy()

	time.Sleep(1 * time.Second)
	for _, file := range []string{fn1, fn2} {
		_, err := os.Stat(file)
		if err != nil {
			t.Fatal(err.Error())
		}
		os.Remove(file)
	}
	fw.Destroy()
}

func TestFileRotate_04(t *testing.T) {
	//t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	fn1 := "rotate_day_04.log"
	fn2 := "rotate_day_04_" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + "_0001.log"
	testFileDailyRotate(t, fn1, fn2)
}

func TestFileRotate_05(t *testing.T) {
	//t.SkipNow()
	FDebug("")
	FDebug("==========================================================================")
	FDebug("")
	fn1 := "rotate_day_05.log"
	fn0 := "rotate_day_05_" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + "_0001.log"
	fnd, _ := os.Create(fn0)
	fn2 := "rotate_day_05_" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + "_0002.log"
	testFileDailyRotate(t, fn1, fn2)
	fnd.Close()
	os.Remove(fn0)
}

func testFileDailyRotate(t *testing.T, fn1, fn2 string) {
	fw := &fileLogWriter{
		Daily:   true,
		MaxDays: 7,
		Rotate:  true,
		Level:   LevelDebug,
		Perm:    "0660",
	}
	FDebug("===> 初始化日志 ......")
	fw.Init(fmt.Sprintf(`{"filename":"%v","maxdays":1}`, fn1))
	time.Sleep(20 * time.Microsecond)

	FDebug("===> 调整日志的当前天 ......")
	fw.dailyOpenTime = time.Now().Add(-24 * time.Hour)
	fw.dailyOpenDay = fw.dailyOpenTime.Day()
	today, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), fw.dailyOpenTime.Location())
	today = today.Add(-1 * time.Second)
	FDebug("= 当前日志所在天（%d号），当前时间：%s", fw.dailyOpenDay, today.Format("2006-01-02 15:04:05"))
	FDebug("===> 准备写入日志 ......")
	fw.WriteMsg("file.go", 10, 4, "main()", LevelDebug, time.Now(), "this is a msg for test")

	FDebug("===> 准备分割日志 ......")
	FDebug("= 预计 1 秒后会自动切换文件(定时器会自动切换，这里是手动触发)")
	time.Sleep(1 * time.Second)
	fw.timerRotate(today)
	fw.Destroy()
	//time.Sleep(1 * time.Second)
	for _, file := range []string{fn1, fn2} {
		_, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}
		os.Remove(file)
	}
	fw.Destroy()
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func BenchmarkFile(b *testing.B) {
	b.SkipNow()
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileAsync(b *testing.B) {
	b.SkipNow()
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	log.SetSync()
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileCallDepth(b *testing.B) {
	b.SkipNow()
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	log.SetLogFuncCallDepth(2)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileAsyncCallDepth(b *testing.B) {
	b.SkipNow()
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	log.SetLogFuncCallDepth(2)
	log.SetSync()
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileOnGoroutine(b *testing.B) {
	b.SkipNow()
	log := NewLogger(100000)
	log.SetLogger("file", `{"filename":"test4.log"}`)
	for i := 0; i < b.N; i++ {
		go log.Debug("debug")
	}
	os.Remove("test4.log")
}
