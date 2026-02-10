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

import "time"

var gLogger *TLogger = NewLogger()

// InitGLogger 初始化全局日志记录器
func InitGLogger(level int) *TLogger {
	gLogger.SetLevel(level)
	return gLogger
}

// Reset 重置全局日志记录器
func Reset() {
	gLogger.Reset()
}

// Flush 刷新全局日志记录器
func Flush() {
	gLogger.Flush()
}

// SetSync 设置同步模式
func SetSync(msgLen ...int64) *TLogger {
	return gLogger.SetSync(msgLen...)
}

// SetLevel 设置日志级别
func SetLevel(l int, adapters ...string) {
	if l <= LevelDebug && l >= LevelEmergency {
		gLogger.SetLevel(l, adapters...)
	}
}

// SetFDebug 设置调试模式
func SetFDebug(l bool) {
	IsDebug = l
}

// GetLevel 获取日志级别
func GetLevel(adapter ...string) int {
	return gLogger.GetLevel(adapter...)
}

// GetLastLogTime 获取最后日志时间
func GetLastLogTime() time.Time {
	return gLogger.GetLastLogTime()
}

// SetLogFuncCallDepth 设置日志函数调用深度
func SetLogFuncCallDepth(d int) {
	gLogger.funcCallDepth = d
}

// 设置日志是否输出到标准错误，默认为false
func SetConsole2Stderr(f bool) {
	bstd_err = f
}

// SetLogger 设置日志适配器
func SetLogger(adapter string, config ...string) error {
	err := gLogger.SetLogger(adapter, config...)
	if err != nil {
		return err
	}
	return nil
}

// DelLogger 删除日志适配器
func DelLogger(adapter string) error {
	err := gLogger.DelLogger(adapter)
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

// Print logs a message at debug level.
func Print(f interface{}, v ...interface{}) {
	gLogger.Print(formatLog(f, v...))
}

// Begin logs a message at debug level.
func Begin() {
	gLogger.Begin()
}

// End logs a message at debug level.
func End() {
	gLogger.End()
}
