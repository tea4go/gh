package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	flag "github.com/spf13/pflag"
	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/network"
	"github.com/tea4go/gh/utils"
)

func filepathJoin(elem ...string) string {
	path := filepath.Join(elem...)
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

const usage = `
  本工具提供 日志测试 的功能。
`

// 标准程序块
var appName string = "log4go"
var appVer string = "v5.5.0"
var IsBeta string = "true"
var BuildTime string

func main() {
	pShow := flag.BoolP(`show`, `s`, false, `显示日志文件名。`)
	flag.Usage = func() {
		fmt.Printf("使用说明: %s\n", utils.GetFileBaseName(os.Args[0]))
		flag.PrintDefaults()
		fmt.Print(usage)
	}

	flag.Parse()

	//#region 处理日志
	logs_file_name := filepathJoin(os.TempDir(), "ulog_"+appName+".txt")
	logs.SetLogger("file", fmt.Sprintf(`{"filename":"%s", "perm": "0666"}`, logs_file_name))
	//#endregion

	network.SetAppVersion(appName, appVer, IsBeta, BuildTime) //设置应用版本号，便于自动更新
	appVer = network.AppVersion
	logs.StartLogger(appName)
	network.StartSelfUpdate()
	if *pShow {
		if runtime.GOOS == "windows" {
			logs_file_name = strings.ReplaceAll(logs_file_name, "/", "\\")
		}
		fmt.Println("tail -f", logs_file_name)
		return
	}

	if flag.NArg() >= 1 {
		logtext := strings.Join(flag.Args(), " ")
		logs.Notice("[%s] %s", appName, logtext)
	}
}
