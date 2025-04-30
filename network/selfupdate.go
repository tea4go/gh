package network

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
	logs "github.com/tea4go/gh/log4go"
)

var AppName string
var AppVersion string

const (
	baseURL = "http://nj.yj2025.icu:23432/update" // 更新服务器基础URL
)

func SetAppVersion(appname, appver string) {
	AppName = appname
	AppVersion = appver
}

// checkForUpdate 检查是否有新版本可用
func CheckForUpdate() (string, string, string, error) {
	if AppName == "" || AppVersion == "" {
		return "", "", "", fmt.Errorf("未设置AppName或AppVersion")
	}

	// 获取版本信息文件
	url := fmt.Sprintf("%s/%s/%s.txt", baseURL, AppName, AppName)
	logs.Debug("检测新版本，版本地址: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", "", fmt.Errorf("无法连接到版本服务器: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("获取版本信息失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 读取版本信息
	versionData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("读取版本信息失败: %v", err)
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
		downurl := fmt.Sprintf("%s/%s/%s.%s.%s.%s", baseURL, AppName, AppName, latestVersion, runtime.GOOS, runtime.GOARCH)
		if runtime.GOOS == "windows" {
			downurl += ".exe"
		}
		logs.Debug("发现新版本，下载地址: %s", downurl)
		return latestVersion, checksum, downurl, nil
	}

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
