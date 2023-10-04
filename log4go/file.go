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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// fileLogWriter implements LoggerInterface.
// It writes messages by lines limit, file size limit, or time frequency.
type fileLogWriter struct {
	sync.RWMutex // write log order by order and  atomic incr maxLinesCurLines and maxSizeCurSize
	// The opened file
	Filename   string `json:"filename"`
	fileWriter *os.File

	// Rotate at line （日志滚动处理 - 行）
	MaxLines         int `json:"maxlines"`
	maxLinesCurLines int

	// Rotate at size （日志滚动处理 - 文件数）
	MaxFiles         int `json:"maxfiles"`
	MaxFilesCurFiles int

	// Rotate at size （日志滚动处理 - 大小）
	MaxSize        int `json:"maxsize"`
	maxSizeCurSize int

	// Rotate daily （日志滚动处理 - 日）
	Daily         bool  `json:"daily"`
	MaxDays       int64 `json:"maxdays"`
	dailyOpenDate int
	dailyOpenTime time.Time

	Rotate bool   `json:"rotate"`
	Level  int    `json:"level"`
	Perm   string `json:"perm"`

	fileNameOnly, suffix string // like "project.log", project is fileNameOnly and .log is suffix
}

// newFileWriter create a FileLogWriter returning as LoggerInterface.
func newFileWriter() ILogger {
	w := &fileLogWriter{
		Daily:   true,
		MaxDays: 7,
		Rotate:  true,
		Level:   LevelNotice,
		Perm:    "0660",
	}
	return w
}

// Init file logger with json config.
// jsonConfig like:
//
//		{
//		"filename":"logs/beego.log",
//		"maxLines":10000,
//		"maxsize":1024,
//		"daily":true,
//		"maxDays":15,
//		"rotate":true,
//	 	"perm":"0600"
//		}
func (w *fileLogWriter) Init(jsonConfig string) error {
	err := json.Unmarshal([]byte(jsonConfig), w)
	if err != nil {
		return err
	}
	if len(w.Filename) == 0 {
		return errors.New("配置字符串里面必须有文件名。")
	}
	w.suffix = filepath.Ext(w.Filename)
	w.fileNameOnly = strings.TrimSuffix(w.Filename, w.suffix)
	if w.suffix == "" {
		w.suffix = ".log"
	}
	err = w.startLogger()
	return err
}

// start file logger. create log file and set to locker-inside file writer.
// 启动文件记录器。 创建日志文件并设置为储物柜内部文件编写器。
func (w *fileLogWriter) startLogger() error {
	file, err := w.createLogFile()
	if err != nil {
		return err
	}
	if w.fileWriter != nil {
		w.fileWriter.Close()
	}
	w.fileWriter = file
	return w.initFd()
}

func (w *fileLogWriter) needRotate(size int, day int) bool {
	return (w.MaxLines > 0 && w.maxLinesCurLines >= w.MaxLines) ||
		(w.MaxSize > 0 && w.maxSizeCurSize >= w.MaxSize) ||
		(w.Daily && day != w.dailyOpenDate)
}

// WriteMsg write logger message into file.
func (w *fileLogWriter) WriteMsg(fileName string, fileLine int, callLevel int, callFunc string, logLevel int, when time.Time, msg string) error {
	h := msg + "\n"
	_, d := formatTimeHeader(when)
	if logLevel != LevelPrint {
		h = fmt.Sprintf("[%5d]%s %s:%d(%s) %s> %s\n", os.Getpid(), when.Format("15:04:05"), fileName, fileLine, callFunc, levelPrefix[logLevel], msg)
	}

	if w.Rotate {
		// 判断是否需要更换文件
		w.RLock()
		if w.needRotate(len(h), d) {
			w.RUnlock()

			// 开始更换文件
			w.Lock()
			if w.needRotate(len(h), d) {
				if err := w.doRotate(when); err != nil {
					fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.Filename, err)
				}
			}
			w.Unlock()
		} else {
			w.RUnlock()
		}
	}

	// 写入文件，并且更新计数器
	w.Lock()
	_, err := w.fileWriter.Write([]byte(h))
	if err == nil {
		w.maxLinesCurLines++
		w.maxSizeCurSize += len(h)
	}
	w.Unlock()
	return err
}

func (w *fileLogWriter) createLogFile() (*os.File, error) {
	// Open the log file
	perm, err := strconv.ParseInt(w.Perm, 8, 64)
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile(w.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
		os.Chmod(w.Filename, os.FileMode(perm))
	}
	return fd, err
}

func (w *fileLogWriter) initFd() error {
	fd := w.fileWriter
	fInfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败，%s\n", err.Error())
	}

	w.maxSizeCurSize = int(fInfo.Size())

	// 按天滚动文件
	w.dailyOpenTime = GetNow()
	w.dailyOpenDate = w.dailyOpenTime.Day()
	if w.Daily {
		go w.dailyRotate(w.dailyOpenTime)
	}

	// 统计文件行数
	w.maxLinesCurLines = 0
	if fInfo.Size() > 0 {
		count, err := w.lines()
		if err != nil {
			return err
		}
		w.maxLinesCurLines = count
	}
	return nil
}

func (w *fileLogWriter) dailyRotate(openTime time.Time) {
	// 获取第二天的年月日
	y, m, d := openTime.Add(24 * time.Hour).Date()
	nextDay := time.Date(y, m, d, 0, 0, 0, 0, openTime.Location())

	tm := time.NewTimer(time.Duration(nextDay.UnixNano() - openTime.UnixNano() + 100))
	select {
	case <-tm.C:
		w.Lock()
		if w.needRotate(0, GetNow().Day()) {
			if err := w.doRotate(GetNow()); err != nil {
				fmt.Fprintf(os.Stderr, "文件切割失败(%q)，%s\n", w.Filename, err.Error())
			}
		}
		w.Unlock()
	}
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

// 获取文件行数
func (w *fileLogWriter) lines() (int, error) {
	fd, err := os.Open(w.Filename)
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

// DoRotate means it need to write file in new file.
// new file name like xx.2013-01-01.log (daily) or xx.001.log (by line or size)
func (w *fileLogWriter) doRotate(logTime time.Time) error {
	// 查找下一个可用数
	num := 1
	fName := ""

	_, err := os.Lstat(w.Filename)
	if err != nil {
		goto RESTART_LOGGER // 如果文件不存在，则重新计数
	}

	if w.MaxLines > 0 || w.MaxSize > 0 {
		for ; err == nil && num <= 999; num++ {
			fName = w.fileNameOnly + fmt.Sprintf("_%s_%03d%s", logTime.Format("2006-01-02"), num, w.suffix)
			_, err = os.Lstat(fName)
		}
	} else {
		fName = fmt.Sprintf("%s_%s%s", w.fileNameOnly, w.dailyOpenTime.Format("2006-01-02"), w.suffix)
		_, err = os.Lstat(fName)
		for ; err == nil && num <= 999; num++ {
			fName = w.fileNameOnly + fmt.Sprintf("_%s_%03d%s", w.dailyOpenTime.Format("2006-01-02"), num, w.suffix)
			_, err = os.Lstat(fName)
		}
	}
	// return error if the last file checked still existed
	if err == nil {
		return fmt.Errorf("Rotate: Cannot find free log number to rename %s\n", w.Filename)
	}

	// close fileWriter before rename
	w.fileWriter.Close()

	// Rename the file to its new found name
	// even if occurs error,we MUST guarantee to  restart new logger
	err = os.Rename(w.Filename, fName)
	err = os.Chmod(fName, os.FileMode(440))

RESTART_LOGGER: // 开始新的日志文件
	newlgerr := w.startLogger()
	go w.deleteOldLog()

	if newlgerr != nil {
		return fmt.Errorf("新建日志文件错误，%s\n", newlgerr.Error())
	}
	if err != nil {
		return fmt.Errorf("Rotate: %s\n", err)
	}
	return nil

}

func (w *fileLogWriter) deleteOldLog() {
	dir := filepath.Dir(w.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete old log '%s', error: %v\n", path, r)
			}
		}()

		if info == nil {
			return
		}

		if !info.IsDir() && info.ModTime().Add(24*time.Hour*time.Duration(w.MaxDays)).Before(GetNow()) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(w.fileNameOnly)) &&
				strings.HasSuffix(filepath.Base(path), w.suffix) {
				os.Remove(path)
			}
		}
		return
	})
}

// Destroy close the file description, close file writer.
func (w *fileLogWriter) Destroy() {
	w.fileWriter.Close()
}

// Flush flush file logger.
// there are no buffering messages in file logger in memory.
// flush file means sync file from disk.
func (w *fileLogWriter) Flush() {
	w.fileWriter.Sync()
}

func (w *fileLogWriter) SetLevel(l int) {
	w.Level = l
}

func init() {
	Register(AdapterFile, newFileWriter)
}
