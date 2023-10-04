package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shiena/ansicolor"
)

/*
格式：\033[显示方式;前景色;背景色m

说明：
前景色            背景色           颜色
---------------------------------------
30                40              黑色
31                41              红色
32                42              绿色
33                43              黃色
34                44              蓝色
35                45              紫红色
36                46              青蓝色
37                47              白色
显示方式           意义
-------------------------
0                终端默认设置
1                高亮显示
4                使用下划线
5                闪烁
7                反白显示
8                不可见

例子：
\033[1;31;40m
\033[0m
*/

// brush is a color join function
type brush func(string) string

// newBrush return a fix color Brush
func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []brush{
	newBrush("1;37;41"), // Emergency          white
	newBrush("1;37;45"), // Alert              cyan
	newBrush("1;33;46"), // Critical           magenta
	newBrush("1;31"),    // Error              red
	newBrush("1;33"),    // Warning            yellow
	newBrush("1;32"),    // Notice             green
	newBrush("1;34"),    // Info      blue
	newBrush("1;37"),    // Debug              blue
	newBrush("1;37"),    // Print              blue
}

// consoleWriter implements LoggerInterface and writes messages to terminal.
type consoleWriter struct {
	lgwi      *logWriter
	Level     int  `json:"level"`
	ColorFlag bool `json:"color"` //this filed is useful only when system's terminal supports color
}

// NewConsole create ConsoleWriter returning as LoggerInterface.
func NewConsole() ILogger {
	cw := &consoleWriter{
		lgwi:      newLogWriter(ansicolor.NewAnsiColorWriter(os.Stdout)),
		Level:     LevelDebug,
		ColorFlag: true, //runtime.GOOS != "windows",
	}
	return cw
}

// Init init console logger.
// jsonConfig like '{"level":LevelTrace}'.
func (c *consoleWriter) Init(jsonConfig string) error {
	if len(jsonConfig) == 0 {
		return nil
	}
	err := json.Unmarshal([]byte(jsonConfig), c)
	return err
}

// 打印内容：
// server.go:192            [N]> ==>网络协议： udp
func (c *consoleWriter) WriteMsg(fileName string, fileLine int, callLevel int, callFunc string, logLevel int, when time.Time, msg string) error {
	if (logLevel > c.Level) && (logLevel != LevelPrint) {
		return nil
	}

	h := msg + "\n"
	if logLevel != LevelPrint {
		head := fmt.Sprintf("(%s:%d)", fileName, fileLine)
		h = fmt.Sprintf("%s %-25s %s> %s\n", when.Format("15.04.05"), head, levelPrefix[logLevel], msg)
	}
	if c.ColorFlag {
		h = colors[logLevel](h)
	}

	c.lgwi.wi.Write([]byte(h))
	return nil
}

// Destroy implementing method. empty.
func (c *consoleWriter) Destroy() {

}

// Flush implementing method. empty.
func (c *consoleWriter) Flush() {

}

func (w *consoleWriter) SetLevel(l int) {
	w.Level = l
}

func init() {
	Register(AdapterConsole, NewConsole)
}
