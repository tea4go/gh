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
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

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

func GetLevelName(level int) string {
	if level <= LevelDebug && level >= LevelEmergency {
		return levelName[level]
	}
	return "无效"
}

// 获取当前时间
var time_local *time.Location = nil

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

var IsDebug bool = true

func FDebug(f string, v ...interface{}) {
	if IsDebug {
		// writeMsg will always add a '\n' character
		if len(f) > 0 && f[len(f)-1] == '\n' {
			f = f[0 : len(f)-1]
		}
		fmt.Printf(f, v...)
		fmt.Println()
	}
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

// 获取文件行数
func GetFileLines(filename string) (int, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		// 通过统计字符出现次数，这个算法统计效率比较高
		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}
