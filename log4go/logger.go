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
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

type logWriter struct {
	sync.Mutex
	wi io.Writer
}

func newLogWriter(wr io.Writer) *logWriter {
	return &logWriter{wi: wr}
}

func (lg *logWriter) writeln(msg string) (int, error) {
	lg.Lock()
	defer lg.Unlock()

	// if len(msg) > 0 && msg[len(msg)-1] == '\n' {
	// 	msg = msg[0 : len(msg)-1]
	// }
	// msg = msg + "\n"
	n, err := lg.wi.Write([]byte(msg))
	return n, err
}

// NewLogger 返回一个新的 BeeLogger。
// channelLen 表示通道中的消息数量（用于 asynchronous 为 true 时）。
// 如果缓冲通道已满，日志适配器将写入文件或其他方式。
func NewLogger(channelLens ...int64) *TLogger {
	bl := new(TLogger)
	bl.funcCallDepth = 4
	bl.msgChanLen = append(channelLens, 0)[0]
	if bl.msgChanLen <= 0 {
		bl.msgChanLen = defAsyncMsgLen
	}
	bl.signalChan = make(chan string, 1)
	bl.init_flag = true
	//bl.setLogger(AdapterConsole)
	return bl
}

// Async set the log to asynchronous and start the goroutine
var logMsgPool *sync.Pool

// SetSync 设置同步模式
func (bl *TLogger) SetSync(msgLen ...int64) *TLogger {
	FDebug("SetSync() : 设置异步写入模式")
	bl.lock.Lock()
	defer bl.lock.Unlock()

	if bl.Async_flag {
		return bl
	}

	bl.Async_flag = true
	if len(msgLen) > 0 && msgLen[0] > 0 {
		bl.msgChanLen = msgLen[0]
	}
	bl.msgChan = make(chan *tLogMsg, bl.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &tLogMsg{}
		},
	}
	bl.wg.Add(1)
	go bl.startDaemon()

	return bl
}

// SetLogger 提供给定的日志适配器到 BeeLogger 与配置字符串。
// config 需要是正确的 JSON 字符串：{"interval":360}。
func (bl *TLogger) SetLogger(adapterName string, configs ...string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	if !bl.init_flag {
		bl.outputs = []*nameLogger{}
		bl.init_flag = true
	}
	return bl.setLogger(adapterName, configs...)
}

// setLogger 提供给定的日志适配器到 BeeLogger 与配置字符串。
// config 需要是正确的 JSON 字符串：{"interval":360}。
func (bl *TLogger) setLogger(adapterName string, configs ...string) error {
	config := append(configs, "{}")[0]
	for _, l := range bl.outputs {
		if l.name == adapterName {
			return fmt.Errorf("日志适配器已存在（%s）", adapterName)
		}
	}

	new_logger_func, ok := adapters[adapterName]
	if !ok {
		return fmt.Errorf("未知的日志适配器（%s）", adapterName)
	}

	//FDebug("SetLogger(%s) : %s", adapterName, config)
	lg := new_logger_func()
	err := lg.Init(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志实例错误（%s），%s\n", adapterName, err.Error())
		return err
	}
	bl.outputs = append(bl.outputs,
		&nameLogger{
			name:    adapterName,
			ILogger: lg,
		})
	return nil
}

// DelLogger 删除 BeeLogger 中的日志适配器。
func (bl *TLogger) DelLogger(adapterName string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	newoutputs := []*nameLogger{}
	for _, lg := range bl.outputs {
		if lg.name == adapterName {
			lg.Destroy()
		} else {
			newoutputs = append(newoutputs, lg)
		}
	}
	if len(newoutputs) == len(bl.outputs) {
		return fmt.Errorf("删除日志处理器失败，未知的日志处理器（%s）。", adapterName)
	}
	bl.outputs = newoutputs

	return nil
}

// Write 写入日志
func (bl *TLogger) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// writeMsg will always add a '\n' character
	if p[len(p)-1] == '\n' {
		p = p[0 : len(p)-1]
	}
	// set levelLoggerImpl to ensure all log message will be write out
	err = bl.writeMsg(levelLoggerImpl, string(p))
	if err == nil {
		return len(p), err
	}
	return 0, err
}

func (bl *TLogger) writeMsg(logLevel int, msg string, v ...interface{}) error {
	bl.lastTime = time.Now()

	// 如果没有初始化，则初始化控制台日志
	if !bl.init_flag {
		bl.lock.Lock()
		bl.setLogger(AdapterConsole)
		bl.lock.Unlock()
	}

	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}

	when := GetNow()
	callLevel, funcname, filename, line := bl.GetCallStack()
	if bl.Async_flag {
		// 使用Pool获取 日志对象
		lm := logMsgPool.Get().(*tLogMsg)
		lm.fileName = filename
		lm.fileLine = line
		lm.callLevel = callLevel
		lm.callFunc = funcname
		lm.logLevel = logLevel
		lm.when = when
		lm.msg = msg
		// 把 日志对象 放到 chan
		bl.msgChan <- lm
	} else {
		bl.writeToLoggers(filename, line, callLevel, funcname, logLevel, when, msg)
	}
	return nil
}

// 每个日志处理器，写入日志字符串
func (bl *TLogger) writeToLoggers(fileName string, fileLine int, callLevel int, callFunc string, logLevel int, when time.Time, msg string) {
	for _, l := range bl.outputs {
		err := l.WriteMsg(fileName, fileLine, callLevel, callFunc, logLevel, when, msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "写入日志失败（%s），%s\n", l.name, err)
		}
	}
}

// GetClassName 获取类名
// github.com/tea4go/application/myproxy/service.THTTP.StartServer
func (bl *TLogger) GetClassName(func_name string) string {
	result := ""
	t1 := strings.Split(func_name, "/")
	if len(t1) <= 1 {
		return func_name
	}
	for _, t := range t1[:len(t1)-1] {
		result += fmt.Sprintf("%c.", t[0])
	}
	result += fmt.Sprintf("%s", t1[len(t1)-1])
	return result
}

// GetCallStack 获取调用堆栈
func (bl *TLogger) GetCallStack() (level int, stack string, file string, line int) {
	level = bl.funcCallDepth
	stack = ""
	file = "???"
	line = 0
	for level <= 30 {
		pc, tfile, tline, ok := runtime.Caller(level)
		if ok {
			if level == bl.funcCallDepth {
				_, file = path.Split(tfile)
				line = tline
			}
			f := runtime.FuncForPC(pc)
			fn := strings.Replace(f.Name(), "main.", "", -1)
			fn = strings.Replace(fn, "(*", "", -1)
			fn = strings.Replace(fn, ").", ".", -1)

			if "main.main" == f.Name() {
				stack = "main->" + stack
				break
			}
			stack = fn + "->" + stack
		} else {
			break
		}
		level++
	}
	t1 := strings.Split(stack[:len(stack)-2], "->")
	return level - bl.funcCallDepth, bl.GetClassName(t1[len(t1)-1]), file, line
}

// SetLevel 设置日志消息级别。
// 如果消息级别（如 LevelDebug）高于记录器级别（如 LevelWarning），
// 日志提供程序甚至不会发送该消息。
func (bl *TLogger) SetLevel(l int, adapters ...string) {
	adapters = append(adapters, "")
	adapter_name := adapters[0]
	if l <= LevelDebug && l >= LevelEmergency {
		for _, ll := range bl.outputs {
			//fmt.Println(adapter_name, ll.name)
			if adapter_name == "" || adapter_name == ll.name {
				FDebug("SetLevel(%s) : %s", ll.name, GetLevelName(l))
				ll.SetLevel(l)
			}
		}
	} else {
		FDebug("设置日志级别失败！")
	}
}

// SetFDebug 设置调试模式
func (bl *TLogger) SetFDebug(l bool) {
	//fmt.Println("设置Log4go调试：", l)
	IsDebug = l
}

// GetLevel 获取日志级别
func (bl *TLogger) GetLevel(adapters ...string) int {
	adapters = append(adapters, "")
	adapter_name := adapters[0]
	for _, ll := range bl.outputs {
		//fmt.Printf("GetLevel [%s]=[%s] %d\n", ll.name, adapter_name, ll.GetLevel())
		if adapter_name == "" || ll.name == adapter_name {
			return ll.GetLevel()
		}
	}
	return -1
}

// GetLastLogTime 获取最后日志时间
func (bl *TLogger) GetLastLogTime() time.Time {
	return bl.lastTime
}

// SetLogFuncCallDepth set log funcCallDepth
func (bl *TLogger) SetLogFuncCallDepth(d int) {
	bl.funcCallDepth = d
}

// GetLogFuncCallDepth return log funcCallDepth for wrapper
func (bl *TLogger) GetLogFuncCallDepth() int {
	return bl.funcCallDepth
}

// start logger chan reading.
// when chan is not empty, write logs.
func (bl *TLogger) startDaemon() {
	FDebug("StartDaemon() : Begin")
	gameOver := false
	for {
		select {
		case bm := <-bl.msgChan:
			// 异步处理日志消息
			FDebug("StartDaemon() : 处理消息")
			bl.writeToLoggers(bm.fileName, bm.fileLine, bm.callLevel, bm.callFunc, bm.logLevel, bm.when, bm.msg)
			logMsgPool.Put(bm)
		case sg := <-bl.signalChan:
			FDebug("StartDaemon() : 接收消息(%s)", sg)
			bl.flush()
			if sg == "close" {
				for _, l := range bl.outputs {
					l.Destroy()
				}
				bl.outputs = nil
				gameOver = true
			}
			bl.wg.Done()
		}
		if gameOver {
			break
		}
	}
	FDebug("StartDaemon() : End")
}

// Emergency Log EMERGENCY level message.
func (bl *TLogger) Emergency(format string, v ...interface{}) {
	bl.writeMsg(LevelEmergency, format, v...)
}

// Alert Log ALERT level message.
func (bl *TLogger) Alert(format string, v ...interface{}) {
	bl.writeMsg(LevelAlert, format, v...)
}

// Critical Log CRITICAL level message.
func (bl *TLogger) Critical(format string, v ...interface{}) {
	bl.writeMsg(LevelCritical, format, v...)
}

// Error Log ERROR level message.
func (bl *TLogger) Error(format string, v ...interface{}) {
	bl.writeMsg(LevelError, format, v...)
}

// Warning Log WARNING level message.
func (bl *TLogger) Warning(format string, v ...interface{}) {
	bl.writeMsg(LevelWarning, format, v...)
}

// Notice Log NOTICE level message.
func (bl *TLogger) Notice(format string, v ...interface{}) {
	bl.writeMsg(LevelNotice, format, v...)
}

// Info Log Info level message.
func (bl *TLogger) Info(format string, v ...interface{}) {
	bl.writeMsg(LevelInfo, format, v...)
}

// Debug Log DEBUG level message.
func (bl *TLogger) Debug(format string, v ...interface{}) {
	bl.writeMsg(LevelDebug, format, v...)
}

// Print Log DEBUG level message.
func (bl *TLogger) Print(format string, v ...interface{}) {
	bl.writeMsg(LevelPrint, format, v...)
}

// Begin Log DEBUG level message.
func (bl *TLogger) Begin() {
	bl.writeMsg(LevelDebug, "Begin")
}

// End Log DEBUG level message.
func (bl *TLogger) End() {
	bl.writeMsg(LevelDebug, "End\n")
}

func (bl *TLogger) flush() {
	if bl.Async_flag {
		for {
			if len(bl.msgChan) > 0 {
				bm := <-bl.msgChan
				bl.writeToLoggers(bm.fileName, bm.fileLine, bm.callLevel, bm.callFunc, bm.logLevel, bm.when, bm.msg)
				logMsgPool.Put(bm)
				continue
			}
			break
		}
	}
	for _, l := range bl.outputs {
		l.Flush()
	}
}

// Flush flush all chan data.
func (bl *TLogger) Flush() {
	if bl.Async_flag {
		bl.signalChan <- "flush"
		bl.wg.Wait()
		bl.wg.Add(1)
		return
	}
	bl.flush()
}

// Close close logger, flush all chan data and destroy all adapters in BeeLogger.
func (bl *TLogger) Close() {
	if bl.Async_flag {
		FDebug("Close() : 关闭日志")
		bl.signalChan <- "close"
		bl.wg.Wait()
		bl.wg.Add(1)
		if bl.msgChan != nil {
			close(bl.msgChan)
			bl.msgChan = nil
		}
	} else {
		bl.flush()
		for _, l := range bl.outputs {
			l.Destroy()
		}
		bl.outputs = nil
	}
	if bl.signalChan != nil {
		close(bl.signalChan)
		bl.signalChan = nil
	}
}

// Reset close all outputs, and set bl.outputs to nil
func (bl *TLogger) Reset() {
	bl.Flush()
	for _, l := range bl.outputs {
		l.Destroy()
	}
	bl.outputs = nil
}
