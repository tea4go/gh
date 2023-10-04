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

// Package logs provide a general log interface
// Usage:
//
// import "github.com/astaxie/beego/logs"
//
//	log := NewLogger(10000)
//	log.SetLogger("console", "")
//
//	> the first params stand for how many channel
//
// Use it like this:
//
//		log.Trace("trace")
//		log.Info("info")
//		log.Warn("warning")
//		log.Debug("debug")
//		log.Critical("critical")
//
//	 more docs http://beego.me/docs/module/logs.md
package logs

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RFC5424 log message levels.
const (
	LevelEmergency = iota //事故
	LevelAlert            //警报
	LevelCritical         //危险
	LevelError            //错误
	LevelWarning          //警告
	LevelNotice           //通知
	LevelInfo             //信息
	LevelDebug            //调试
	LevelPrint            //打印(直接显示内容，不显示前缀)
)

// levelLogLogger is defined to implement log.Logger
// the real log level will be LevelEmergency
const levelLoggerImpl = -1

// Name for adapter with beego official support
const (
	AdapterConsole   = "console"
	AdapterFile      = "file"
	AdapterMultiFile = "multifile"
	AdapterMail      = "smtp"
	AdapterConn      = "conn"
	AdapterEs        = "es"
	AdapterJianLiao  = "jianliao"
	AdapterSlack     = "slack"
)

type newLoggerFunc func() ILogger

// Logger defines the behavior of a log provider.
type ILogger interface {
	Init(config string) error
	SetLevel(l int)
	WriteMsg(fileName string, fileLine int, callLevel int, callFunc string, logLevel int, when time.Time, msg string) error
	Destroy()
	Flush()
}

var adapters = make(map[string]newLoggerFunc)
var levelPrefix = [LevelDebug + 2]string{"[M]", "[A]", "[C]", "[E]", "[W]", "[N]", "[I]", "[D]", "[P]"}
var levelName = [LevelDebug + 2]string{"事故[M]", "警报[A]", "危险[C]", "错误[E]", "警告[W]", "通知[N]", "信息[I]", "调试[D]", "打印[P]"}

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log newLoggerFunc) {
	if log == nil {
		panic("日志注册失败，没有日志处理器。")
	}
	if _, dup := adapters[name]; dup {
		panic("日志注册失败，已经注册过（" + name + "）")
	}
	adapters[name] = log
}

// BeeLogger is default logger in beego application.
// it can contain several providers and log message into all providers.
type TLogger struct {
	lock                sync.Mutex
	level               int
	init                bool
	loggerFuncCallDepth int
	Async_flag          bool
	msgChanLen          int64
	msgChan             chan *tLogMsg
	signalChan          chan string
	wg                  sync.WaitGroup
	outputs             []*nameLogger
	local               *time.Location
	lastTime            time.Time
}

const defAsyncMsgLen = 1e3

type nameLogger struct {
	ILogger
	name string
}

type tLogMsg struct {
	fileName  string
	fileLine  int
	callFunc  string
	callLevel int
	logLevel  int
	msg       string
	when      time.Time
}

var logMsgPool *sync.Pool
var time_local *time.Location = nil

// 获取当前时间
func GetNow() (result time.Time) {
	defer func() {
		if info := recover(); info != nil {
			result = time.Now()
		}
	}()
	if time_local == nil {
		time_local, _ = time.LoadLocation("Asia/Chongqing")
	}
	result = time.Now().In(time_local)
	return
}

// NewLogger returns a new BeeLogger.
// channelLen means the number of messages in chan(used where asynchronous is true).
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channelLens ...int64) *TLogger {
	bl := new(TLogger)
	bl.level = LevelNotice
	bl.loggerFuncCallDepth = 4
	bl.msgChanLen = append(channelLens, 0)[0]
	if bl.msgChanLen <= 0 {
		bl.msgChanLen = defAsyncMsgLen
	}
	bl.signalChan = make(chan string, 1)
	bl.setLogger(AdapterConsole)
	return bl
}

// Async set the log to asynchronous and start the goroutine
func (bl *TLogger) Async(msgLen ...int64) *TLogger {
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
	go bl.startLogger()
	return bl
}

// SetLogger provides a given logger adapter into BeeLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *TLogger) SetLogger(adapterName string, configs ...string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	if !bl.init {
		bl.outputs = []*nameLogger{}
		bl.init = true
	}
	return bl.setLogger(adapterName, configs...)
}

// SetLogger provides a given logger adapter into BeeLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *TLogger) setLogger(adapterName string, configs ...string) error {
	config := append(configs, "{}")[0]
	for _, l := range bl.outputs {
		if l.name == adapterName {
			return fmt.Errorf("重复的适配器名称（%s）", adapterName)
		}
	}

	new_logger_func, ok := adapters[adapterName]
	if !ok {
		return fmt.Errorf("未知的适配器名称（%s）", adapterName)
	}

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

// DelLogger remove a logger adapter in BeeLogger.
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
	if !bl.init {
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
		lm := logMsgPool.Get().(*tLogMsg)
		lm.fileName = filename
		lm.fileLine = line
		lm.callLevel = callLevel
		lm.callFunc = funcname
		lm.logLevel = logLevel
		lm.when = when
		lm.msg = msg
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

func (bl *TLogger) GetCallStack() (level int, stack string, file string, line int) {
	level = bl.loggerFuncCallDepth
	stack = ""
	file = "???"
	line = 0
	for level <= 30 {
		pc, tfile, tline, ok := runtime.Caller(level)
		if ok {
			if level == bl.loggerFuncCallDepth {
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
	return level - bl.loggerFuncCallDepth, bl.GetClassName(t1[len(t1)-1]), file, line
}

// SetLevel Set log message level.
// If message level (such as LevelDebug) is higher than logger level (such as LevelWarning),
// log providers will not even be sent the message.
func (bl *TLogger) SetLevel(l int) {
	if l <= LevelDebug && l >= LevelEmergency {
		bl.level = l
		for _, ll := range bl.outputs {
			ll.SetLevel(l)
		}
	} else {
		fmt.Println("设置日志级别失败！")
	}
}

func (bl *TLogger) GetLevel() int {
	return bl.level
}

func (bl *TLogger) GetLastLogTime() time.Time {
	return bl.lastTime
}

// SetLogFuncCallDepth set log funcCallDepth
func (bl *TLogger) SetLogFuncCallDepth(d int) {
	bl.loggerFuncCallDepth = d
}

// GetLogFuncCallDepth return log funcCallDepth for wrapper
func (bl *TLogger) GetLogFuncCallDepth() int {
	return bl.loggerFuncCallDepth
}

// start logger chan reading.
// when chan is not empty, write logs.
func (bl *TLogger) startLogger() {
	gameOver := false
	for {
		select {
		case bm := <-bl.msgChan:
			bl.writeToLoggers(bm.fileName, bm.fileLine, bm.callLevel, bm.callFunc, bm.logLevel, bm.when, bm.msg)
			logMsgPool.Put(bm)
		case sg := <-bl.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
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
}

// Emergency Log EMERGENCY level message.
func (bl *TLogger) Emergency(format string, v ...interface{}) {
	if LevelEmergency > bl.level {
		return
	}
	bl.writeMsg(LevelEmergency, format, v...)
}

// Alert Log ALERT level message.
func (bl *TLogger) Alert(format string, v ...interface{}) {
	if LevelAlert > bl.level {
		return
	}
	bl.writeMsg(LevelAlert, format, v...)
}

// Critical Log CRITICAL level message.
func (bl *TLogger) Critical(format string, v ...interface{}) {
	if LevelCritical > bl.level {
		return
	}
	bl.writeMsg(LevelCritical, format, v...)
}

// Error Log ERROR level message.
func (bl *TLogger) Error(format string, v ...interface{}) {
	if LevelError > bl.level {
		return
	}
	bl.writeMsg(LevelError, format, v...)
}

// Warning Log WARNING level message.
func (bl *TLogger) Warning(format string, v ...interface{}) {
	if LevelWarning > bl.level {
		return
	}
	bl.writeMsg(LevelWarning, format, v...)
}

// Notice Log NOTICE level message.
func (bl *TLogger) Notice(format string, v ...interface{}) {
	if LevelNotice > bl.level {
		return
	}
	bl.writeMsg(LevelNotice, format, v...)
}

// Info Log Info level message.
func (bl *TLogger) Info(format string, v ...interface{}) {
	if LevelInfo > bl.level {
		return
	}
	bl.writeMsg(LevelInfo, format, v...)
}

// Debug Log DEBUG level message.
func (bl *TLogger) Debug(format string, v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	bl.writeMsg(LevelDebug, format, v...)
}

// Debug Log DEBUG level message.
func (bl *TLogger) Print(format string, v ...interface{}) {
	if LevelNotice > bl.level {
		return
	}
	bl.writeMsg(LevelPrint, format, v...)
}

// Debug Log DEBUG level message.
func (bl *TLogger) Begin() {
	if LevelDebug > bl.level {
		return
	}
	bl.writeMsg(LevelDebug, "Begin")
}

// Debug Log DEBUG level message.
func (bl *TLogger) End() {
	if LevelDebug > bl.level {
		return
	}
	bl.writeMsg(LevelDebug, "End\n")
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
		bl.signalChan <- "close"
		bl.wg.Wait()
		close(bl.msgChan)
	} else {
		bl.flush()
		for _, l := range bl.outputs {
			l.Destroy()
		}
		bl.outputs = nil
	}
	close(bl.signalChan)
}

// Reset close all outputs, and set bl.outputs to nil
func (bl *TLogger) Reset() {
	bl.Flush()
	for _, l := range bl.outputs {
		l.Destroy()
	}
	bl.outputs = nil
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

// beeLogger references the used application logger.
var gLogger *TLogger = NewLogger()

// GetLogger returns the default BeeLogger
func InitGLogger(level int) *TLogger {
	gLogger.SetLevel(level)
	return gLogger
}

// Reset will remove all the adapter
func Reset() {
	gLogger.Reset()
}

func Async(msgLen ...int64) *TLogger {
	return gLogger.Async(msgLen...)
}

// SetLevel sets the global log level used by the simple logger.
func SetLevel(l int) {
	if l <= LevelDebug && l >= LevelEmergency {
		gLogger.SetLevel(l)
	}
}

func GetLevel() int {
	return gLogger.GetLevel()
}

func GetLastLogTime() time.Time {
	return gLogger.GetLastLogTime()
}

func GetLevelName(level int) string {
	if level <= LevelDebug && level >= LevelEmergency {
		return levelName[level]
	}
	return "无效"
}

// SetLogFuncCallDepth set log funcCallDepth
func SetLogFuncCallDepth(d int) {
	gLogger.loggerFuncCallDepth = d
}

// SetLogger sets a new logger.
func SetLogger(adapter string, config ...string) error {
	err := gLogger.SetLogger(adapter, config...)
	if err != nil {
		return err
	}
	return nil
}

// Emergency logs a message at emergency level.
func Emergency(f interface{}, v ...interface{}) {
	gLogger.Emergency(formatLog(f, v...))
}

// Alert logs a message at alert level.
func Alert(f interface{}, v ...interface{}) {
	gLogger.Alert(formatLog(f, v...))
}

// Critical logs a message at critical level.
func Critical(f interface{}, v ...interface{}) {
	gLogger.Critical(formatLog(f, v...))
}

// Error logs a message at error level.
func Error(f interface{}, v ...interface{}) {
	gLogger.Error(formatLog(f, v...))
}

// Warning logs a message at warning level.
func Warning(f interface{}, v ...interface{}) {
	gLogger.Warning(formatLog(f, v...))
}

// Notice logs a message at notice level.
func Notice(f interface{}, v ...interface{}) {
	gLogger.Notice(formatLog(f, v...))
}

// Info logs a message at info level.
func Info(f interface{}, v ...interface{}) {
	gLogger.Info(formatLog(f, v...))
}

// Debug logs a message at debug level.
func Debug(f interface{}, v ...interface{}) {
	gLogger.Debug(formatLog(f, v...))
}

// Debug logs a message at debug level.
func Print(f interface{}, v ...interface{}) {
	gLogger.Print(formatLog(f, v...))
}

// Debug logs a message at debug level.
func Begin() {
	gLogger.Begin()
}

// Debug logs a message at debug level.
func End() {
	gLogger.End()
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}

func CheckError(pos string, err error) {
	if err != nil {
		where := fmt.Sprintf("在%s失败，原因：%s", pos, err.Error())
		panic(where)
	}
}

func GetParamString(name string, flag_value, default_value string) string {
	//从环境变量读取参数
	env_value := os.Getenv(name)

	if env_value != "" {
		return env_value
	}

	//从命令行读参数
	if default_value != "" && flag_value == "" {
		return default_value
	}

	return flag_value
}

func GetParamInt(name string, flag_value int) int {
	//从环境变量读取参数
	env_value := os.Getenv(name)
	env_int_value, err := strconv.Atoi(env_value)

	if err == nil {
		return env_int_value
	}

	//从命令行读参数
	return flag_value
}

func GetParamBool(name string, flag_value bool) bool {
	//从环境变量读取参数
	env_value := os.Getenv(name)

	if env_value != "" {
		if strings.ToLower(env_value) == "true" {
			return true
		} else {
			return false
		}
	}

	//从命令行读参数
	return flag_value
}
