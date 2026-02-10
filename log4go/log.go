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
	// Init 初始化
	Init(config string) error
	// SetLevel 设置级别
	SetLevel(l int)
	// GetLevel 获取级别
	GetLevel() int
	// WriteMsg 写入消息
	WriteMsg(fileName string, fileLine int, callLevel int, callFunc string, logLevel int, when time.Time, msg string) error
	// Destroy 销毁
	Destroy()
	// Flush 刷新
	Flush()
}

// BeeLogger is default logger in beego application.
// it can contain several providers and log message into all providers.
type TLogger struct {
	lock          sync.Mutex
	init_flag     bool          // 是否初始化
	funcCallDepth int           // 函数调用深度
	Async_flag    bool          // 是否异步消息
	msgChanLen    int64         // 日志对象通道大小
	msgChan       chan *tLogMsg // 日志对象通道
	signalChan    chan string   // 控制（flush / close）消息通道
	lastTime      time.Time     // 最后写入日志时间
	wg            sync.WaitGroup
	outputs       []*nameLogger
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
