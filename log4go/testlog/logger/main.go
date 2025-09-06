package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	flag "github.com/spf13/pflag"
	logs "github.com/tea4go/gh/log4go"
)

func filepathJoin(elem ...string) string {
	path := filepath.Join(elem...)
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

func main() {
	flag.Parse()

	//#region 处理日志
	logs_file_name := filepathJoin(os.TempDir(), "ulog_tea4go.txt")
	logs.SetLogger("file", fmt.Sprintf(`{"filename":"%s", "perm": "0666", "log_level":7}`, logs_file_name))
	//#endregion
	logs.StartLogger("tea4go")

	if flag.NArg() >= 1 {
		logtext := strings.Join(flag.Args(), " ")
		logs.Notice("[tea4go] ", logtext)
	}
}
