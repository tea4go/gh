package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/k0kubun/go-ansi"
	"github.com/minio/selfupdate"
	"github.com/schollz/progressbar/v3"
	flag "github.com/spf13/pflag"
	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/gh/utils"
)

var AppName string
var AppVersion string
var BuildTime string
var VerServer string = "http://nj.yj2025.icu:23432" // 更新服务器基础URL

type progressReader struct {
	reader io.Reader
	bar    *progressbar.ProgressBar
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.bar.Add(n)
	}
	return n, err
}

func SetAppVersion(appname, appver, isbeta, buildtime string) {
	// 设置应用程序名称
	AppName = appname
	// 设置应用程序版本
	AppVersion = appver
	// 设置应用程序编译时间
	BuildTime = buildtime
	isbeta = strings.ToLower(isbeta)
	if isbeta != "" && isbeta != "false" && isbeta != "f" && buildtime != "" {
		const inputLayout = "2006-01-02(15:04:05)"
		t, err := time.Parse(inputLayout, buildtime)
		if err == nil {
			const outputLayout = "20060102_150405"
			AppVersion = fmt.Sprintf("%s_%s", appver, t.Format(outputLayout))
		} else {
			AppVersion = fmt.Sprintf("%s_%s", appver, buildtime)
		}
	}
}

func SetVerServer(serv_url string) {
	// 设置版本服务器地址
	VerServer = serv_url
}

//curl -X POST ^
//  -F "version=v3.0.7_20250427" ^
//  -F "verpath=/update/F112" ^
//  -F "verfile=@C:\DevDisk\Other\MiniXplorer\f112.exe" ^
//  http://localhost:8080/publish?key=tvQ2YthGoV2wymjWVkyc ^
//  | jq

// PublishSoftware 发布软件函数
func PublishSoftware() error {
	// 超级管理员权限验证
	superadmin := logs.GetParamString("BASH_KEY", "", "Null") == "rfoMzV4D8O9owOET33vJ"
	if !superadmin {
		return fmt.Errorf("没有版本发布的权限")
	}

	fmt.Printf("准备发布新版本 ...... %s.%s.%s.%s\n", AppName, AppVersion, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("= 文件名：%s 编译时间：%s\n", utils.RunFileName(), BuildTime)
	fmt.Printf("= 发布到：%s/update/%s\n", VerServer, AppName)

	// 创建一个缓冲区用于存储multipart表单数据
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	// 添加表单字段
	_ = writer.WriteField("version", AppVersion)
	_ = writer.WriteField("appname", AppName)
	_ = writer.WriteField("verpath", "/update/"+AppName)
	_ = writer.WriteField("GOOS", runtime.GOOS)
	_ = writer.WriteField("GOARCH", runtime.GOARCH)
	_ = writer.WriteField("servurl", VerServer)
	_ = writer.WriteField("key", "tvQ2YthGoV2wymjWVkyc")

	puturl := `发布的地址：
curl -X POST ^
 -F "version=%s" ^
 -F "appname=%s" ^
 -F "verpath=/update/%s" ^
 -F "GOOS=%s" ^
 -F "GOARCH=%s" ^
 -F "key=tvQ2YthGoV2wymjWVkyc" ^
 -F "verfile=@%s" ^
 %s/publish ^
 | jq`
	logs.Debug(puturl+"\n", AppVersion, AppName, AppName, runtime.GOOS, runtime.GOARCH, utils.RunFileName(), VerServer)

	// 添加文件
	file, err := os.Open(utils.RunFileName())
	if err != nil {
		return fmt.Errorf("打开文件错误，%v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("verfile", file.Name())
	if err != nil {
		return fmt.Errorf("创建表单文件错误，%v", err)
	}

	// 创建进度条 - 修正后的版本
	bar := progressbar.NewOptions64(
		utils.GetFileSize(os.Args[0]),
		progressbar.OptionSetDescription("正在发布中 ......"),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(1000*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n发布新版本 ...... OK\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	// 包装响应体以跟踪进度
	progressReader := &progressReader{
		reader: file,
		bar:    bar,
	}
	_, err = io.Copy(part, progressReader)
	if err != nil {
		return fmt.Errorf("复制文件内容错误，%v", err)
	}
	// 关闭writer以完成表单写入
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("关闭multipart写入错误，%v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/publish", VerServer), &requestBody)
	if err != nil {
		return fmt.Errorf("创建请求错误，%v", err)
	}
	// 设置Content-Type头部，包含boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求错误，%v", utils.GetNetError(err))
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求错误，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
	return nil
}

// checkForUpdate 检查是否有新版本可用
func CheckForUpdate() (string, string, string, error) {
	if AppName == "" || AppVersion == "" {
		return "", "", "", fmt.Errorf("未设置AppName或AppVersion")
	}

	// 获取版本信息文件
	downurl := fmt.Sprintf("%s/update/%s/%s.%s.%s.txt", VerServer, AppName, AppName, runtime.GOOS, runtime.GOARCH)
	logs.Debug("检测新版本，版本地址: %s", downurl)

	resp, err := http.Get(downurl)
	if err != nil {
		return "", "", "", fmt.Errorf("检测新版本失败，%v", utils.GetNetError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("无法连接到版本服务器，返回HTTP状态码(%d)", resp.StatusCode)
	}

	// 读取版本信息
	versionData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("读取版本信息错误，%v", err)
	}

	latestVersion := ""
	checksum := ""

	// 按行分割内容
	lines := strings.Split(string(versionData), "\n")
	if len(lines) >= 2 {
		latestVersion = strings.TrimSpace(lines[0])
		checksum = strings.TrimSpace(lines[1])
	} else {
		latestVersion = lines[0]
	}
	if latestVersion == "" {
		return "", "", "", fmt.Errorf("版本文件格式不正确！")
	}

	// 检查是否比当前版本新
	if compareVersions(latestVersion, AppVersion) > 0 {
		downurl = fmt.Sprintf("%s/update/%s/%s.%s.%s.%s", VerServer, AppName, AppName, runtime.GOOS, runtime.GOARCH, latestVersion)
		if runtime.GOOS == "windows" {
			downurl += ".exe"
		}
		logs.Debug("发现新版本，下载地址: %s", downurl)
		return latestVersion, checksum, downurl, nil
	}

	logs.Debug("当前已经是最新版本： %s，编译时间：%s，服务器版本： %s ", AppVersion, BuildTime, latestVersion)
	return "", "", "", nil
}

// doUpdate 执行更新操作
func DoUpdate(downurl, checksum string) error {
	logs.Debug("开始下载更新版本(%s)", downurl)

	// 确保下载链接是有效的
	if !strings.HasPrefix(downurl, "http") {
		return fmt.Errorf("无效的更新URL(%s)", downurl)
	}

	// 下载更新文件
	resp, err := http.Get(downurl)
	if err != nil {
		return fmt.Errorf("下载请求错误，%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 404 {
			return fmt.Errorf("文件不存在(%d)", resp.StatusCode)
		}
		return fmt.Errorf("文件下载错误(%d)", resp.StatusCode)
	}

	// 应用更新
	opts := selfupdate.Options{}
	if checksum != "" {
		opts.Checksum, _ = hex.DecodeString(checksum)
	}
	if err := selfupdate.Apply(resp.Body, opts); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("无法回滚，请手动删除版本！(%v)", rerr)
		}
		if strings.Contains(err.Error(), "Updated file has wrong checksum") {
			return fmt.Errorf("版本文件已损坏！")
		}
		return err
	}

	return nil
}

func DoUpdateWithProgress(downurl, checksum string) error {
	logs.Debug("开始下载更新版本(%s)", downurl)

	// 确保下载链接是有效的
	if !strings.HasPrefix(downurl, "http") {
		return fmt.Errorf("无效的更新URL(%s)", downurl)
	}

	// 创建自定义HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	// 发起HEAD请求获取文件大小
	resp, err := client.Head(downurl)
	if err != nil {
		return fmt.Errorf("获取文件大小错误，%v", err)
	}
	defer resp.Body.Close()

	fileSize := resp.ContentLength
	if fileSize <= 0 {
		return fmt.Errorf("无效的文件。")
	}

	// 获取更新文件
	resp, err = client.Get(downurl)
	if err != nil {
		return fmt.Errorf("下载新版本错误，%v", err)
	}
	defer resp.Body.Close()

	// 创建进度条 - 修正后的版本
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription("下载新版本 ......"),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(1000*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n下载新版本 ...... OK\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	// 包装响应体以跟踪进度
	progressReader := &progressReader{
		reader: resp.Body,
		bar:    bar,
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 404 {
			return fmt.Errorf("文件不存在(%d)", resp.StatusCode)
		}
		return fmt.Errorf("文件下载错误(%d)", resp.StatusCode)
	}

	// 应用更新
	opts := selfupdate.Options{}
	if checksum != "" {
		opts.Checksum, _ = hex.DecodeString(checksum)
	}
	if err := selfupdate.Apply(progressReader, opts); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("无法回滚，请手动删除版本！(%v)", rerr)
		}
		if strings.Contains(err.Error(), "Updated file has wrong checksum") {
			return fmt.Errorf("版本文件已损坏！")
		}
		return err
	}

	return nil
}

// calculateCurrentChecksum 计算当前运行程序的 checksum
func CalcChecksum() (string, error) {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("无法获取可执行文件路径，%v", err)
	}

	// 打开当前可执行文件
	file, err := os.Open(execPath)
	if err != nil {
		return "", fmt.Errorf("无法打开可执行文件，%v", err)
	}
	defer file.Close()

	// 计算 SHA256 checksum
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("计算 checksum 失败，%v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func SimpleCalcChecksum() string {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		logs.Error("无法获取可执行文件路径，%v", err)
		return ""
	}

	// 打开当前可执行文件
	file, err := os.Open(execPath)
	if err != nil {
		logs.Error("无法打开可执行文件，%v", err)
		return ""
	}
	defer file.Close()

	// 计算 SHA256 checksum
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		logs.Error("计算 checksum 失败，%v", err)
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// compareVersions 简单的版本号比较函数
// 返回: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}

	// 这里实现简单的版本比较逻辑
	// 实际项目中可能需要更复杂的版本比较
	// 可以使用第三方库如 hashicorp/go-version
	return strings.Compare(v1, v2)
}

var pskipVersion *bool

var pversion *bool
var pupgrade *bool
var ppublish *bool
var phelp *bool

func init() {
	pskipVersion = flag.BoolP(`skip_version`, ``, false, `是否跳过版本检测。`)
	pversion = flag.BoolP("version", "v", false, "显示版本号。")
	pupgrade = flag.BoolP("upgrade", "", false, "更新版本。")
	ppublish = flag.BoolP("publish", "", false, "发布新版本。")

	phelp = flag.BoolP(`help`, `h`, false, `显示帮助。`)
}

func SetSkipVersion() {
	*pskipVersion = true
}

func StartSelfUpdate() {
	skipVersion := logs.GetParamBool("skip_version", *pskipVersion)

	// 显示版本信息
	if *pversion {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	// 显示帮助信息
	if *phelp {
		flag.Usage()
		os.Exit(0)
	}

	if *ppublish {
		//curl -X POST ^
		//  -F "version=v3.0.7_20250427" ^
		//  -F "verpath=/update/F112" ^
		//  -F "verfile=@C:\DevDisk\Other\MiniXplorer\f112.exe" ^
		//  http://localhost:8080/publish?key=tvQ2YthGoV2wymjWVkyc ^
		//  | jq
		err := PublishSoftware()
		if err != nil {
			fmt.Printf("发布新版本失败，%v", err)
		}

		os.Exit(0)
	}

	if !skipVersion {
		latest, checksum, downurl, err := CheckForUpdate()
		if err != nil {
			fmt.Println(err)

		} else if latest != "" {
			fmt.Println("发现新版本！")
			fmt.Println("=============================================================")
			fmt.Printf("= 最新版本：%-27s SHA256：%s \n", latest, checksum)
			fmt.Printf("= 当前版本：%-27s SHA256：%s Build : %s\n", AppVersion, SimpleCalcChecksum(), BuildTime)
			fmt.Printf("= 下载地址：%s\n", downurl)
			fmt.Println("=============================================================")

			//#region 自动升级
			err := DoUpdateWithProgress(downurl, checksum)
			if err != nil {
				fmt.Printf("新版本升级失败，%s\n", err)

			} else {
				fmt.Println("已经升级到新版本，既将退出程序，请重新启动程序！")
			}
			//#endregion

			os.Exit(0)
			return
		}

		//如果是检测版本，则退出程序
		if *pupgrade {
			if latest == "" {
				fmt.Println("当前已经是新版本！")
				fmt.Printf("= 当前版本：%-27s SHA256: %s\n", AppVersion, SimpleCalcChecksum())
				fmt.Printf("= 版本文件：%s.%s.%s.%s\n", AppName, AppVersion, runtime.GOOS, runtime.GOARCH)
			}
			os.Exit(0)
		}
	}
}
