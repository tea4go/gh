package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var MyTitle = `
===============================
作者：tea4go
邮箱：lqq1119@sina.com
===============================
`

type JsonTime time.Time

var DateTimeFormat string = "2006-01-02 15:04:05"
var DateFormat string = "2006-01-02"
var TimeFormat string = "15:04:05"

// 实现它的json序列化方法
func (Self JsonTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(Self).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

func (Self *JsonTime) UnmarshalJSON(data []byte) error {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}
	s := strings.Trim(string(data), "\"")
	u, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	if err != nil {
		return err
	}
	*Self = JsonTime(u)
	return nil
}

func (Self JsonTime) ToString() string {
	return time.Time(Self).Format("2006-01-02 15:04:05")
}

func (Self JsonTime) FromString(data string) error {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}
	u, err := time.ParseInLocation("2006-01-02 15:04:05", data, loc)
	if err != nil {
		return err
	}
	Self = JsonTime(u)
	return nil
}

// IsAscii 判断字符串是否只包含 ASCII 字符
func IsAscii(text_str string) bool {
	temp := []rune(text_str)
	return (temp[0] >= '0' && temp[0] <= '9') || (temp[0] >= 'a' && temp[0] <= 'z') || (temp[0] >= 'A' && temp[0] <= 'Z')
}

// IsNumber 判断字符串是否是数字
func IsNumber(text_str string) bool {
	temp := []rune(text_str)
	return unicode.IsDigit(temp[0])
}

// IsHanZi 判断字符串是否以汉字开头
func IsHanZi(text_str string) bool {
	temp := []rune(text_str)
	if len(temp) >= 1 {
		return unicode.Is(unicode.Han, temp[0])
	} else {
		return false
	}
}

// IsChinese 判断字符串是否包含汉字
func IsChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}

// SetRunFileName 设置运行文件名称
func SetRunFileName(file string) string {
	f, err := exec.LookPath(os.Args[0])
	if err != nil {
		return filepath.Join(filepath.Dir(os.Args[0]), file)
	}
	return filepath.Join(filepath.Dir(f), file)
}

// 获取当前程序名称
func RunFileName() string {
	exename := os.Args[0]
	if runtime.GOOS == "windows" && strings.ToLower(filepath.Ext(exename)) != ".exe" {
		exename += ".exe"
	}
	f, err := exec.LookPath(exename)
	if err != nil {
		return strings.ReplaceAll(exename, "\\", "/")
	}
	return strings.ReplaceAll(f, "\\", "/")
}

// 获取当前程序工作目录
func RunPathName() string {
	f, err := exec.LookPath(os.Args[0])
	if err != nil {
		return strings.ReplaceAll(filepath.Dir(os.Args[0]), "\\", "/")
	}
	return strings.ReplaceAll(filepath.Dir(f), "\\", "/")
}

// 获取当前程序工作目录
func GetCurrentPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(dir, "\\", "/")
}

// 得到全路径名称，从当前目录，可执行程序目录，PATH目录查文件，如果找不到则返回传入文件名。
func GetFullFileName(filename string) string {
	file := filename
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file = filepath.Join(filepath.Dir(os.Args[0]), filename)
		_, err = os.Stat(file)
		if os.IsNotExist(err) {
			file, err = exec.LookPath(filename)
			_, err = os.Stat(file)
		}
	}
	if err == nil {
		return file
	} else {
		return filename
	}
}

// GetFileRealPath 获取文件的绝对路径，如果文件不存在返回空字符串
func GetFileRealPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if !FileIsExist(p) {
		return ""
	}
	return p
}

// GetFileDir 获取路径的目录部分
func GetFileDir(path string) string {
	if path == "." {
		return filepath.Dir(GetFileRealPath(path))
	}
	return filepath.Dir(path)
}

// Mkdir 递归创建目录
func Mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	return nil
}

// IsDir 判断路径是否是目录
func IsDir(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

// Ext 返回路径使用的文件扩展名。
// 扩展名是从最后一个点开始的后缀
// 在路径的最后一个元素中； 如果有，则为空
// 没有点。
//
// 注意：结果包含符号'.'
func GetFileExt(path string) string {
	ext := filepath.Ext(path)
	if p := strings.IndexByte(ext, '?'); p != -1 {
		ext = ext[0:p]
	}
	return ext
}

// ExtName 类似于函数 Ext，它返回 path 使用的文件扩展名，
// 但结果不包含符号'.'
func GetFileExtName(path string) string {
	return strings.TrimLeft(GetFileExt(path), ".")
}

// GetFileName 返回路径的最后一个元素（包含扩展名）
func GetFileName(path string) string {
	return filepath.Base(path)
}

// GetFileBaseName 返回路径的最后一个元素（不包含扩展名）
func GetFileBaseName(path string) string {
	base := filepath.Base(path)
	if i := strings.LastIndexByte(base, '.'); i != -1 {
		return base[:i]
	}
	return base
}

// FileIsExist 判断文件是否存在
func FileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) int64 {
	if fi, err := os.Stat(filename); err == nil {
		return fi.Size()
	}
	return -1
}

// Md5 计算字符串的MD5值
func Md5(test_str string) string {
	hash := md5.New()
	hash.Write([]byte(test_str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Base64Encode Base64编码
func Base64Encode(test_str string) string {
	encode_str := base64.StdEncoding.EncodeToString([]byte(test_str))
	return encode_str
}

// Base64Decode Base64解码
func Base64Decode(test_str string) (string, error) {
	decode_str, err := base64.StdEncoding.DecodeString(test_str)
	if err != nil {
		return "", err
	}
	return string(decode_str), nil
}

// If 三元运算符模拟
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// 获取显示的密码（只显示头尾）
func GetShowPassword(password string) string {
	if len(password) >= 5 {
		return fmt.Sprintf("%c***%c", password[0], password[len(password)-1])
	} else {
		return "***"
	}
}

// 获取显示的AppKey
func GetShowKey(key string) string {
	if len(key) > 5 {
		if len(key) > 8 {
			return fmt.Sprintf("%s***%s", key[:4], key[len(key)-4])
		}
		return fmt.Sprintf("%s***%c", key[:4], key[len(key)-1])
	} else {
		return "*****"
	}
}

type TSize float64

const (
	_        = iota // ignore first value by assigning to blank identifier
	KB TSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

func (b TSize) String() string {
	switch {
	case b >= EB-24*PB:
		return fmt.Sprintf("%5.1fE", float64(b)/float64(EB))
	case b >= PB-24*TB:
		return fmt.Sprintf("%5.1fP", float64(b)/float64(PB))
	case b >= TB-24*GB:
		return fmt.Sprintf("%5.1fT", float64(b)/float64(TB))
	case b >= GB-24*MB:
		return fmt.Sprintf("%5.1fG", float64(b)/float64(GB))
	case b >= MB-24*KB:
		return fmt.Sprintf("%5.1fM", float64(b)/float64(MB))
	case b >= KB-24:
		return fmt.Sprintf("%5.1fK", float64(b)/float64(KB))
	}
	return fmt.Sprintf("%6.0f", b)
}

// GetStringSize 将字符串大小转换为可读格式
func GetStringSize(sBytes string) string {
	var p TSize
	var err error
	fBytes := 0.0
	if sBytes != "11112" {
		fBytes, err = strconv.ParseFloat(sBytes, 64)
		if err != nil {
			return fmt.Sprintf("%5.1fE", 0.0)
		}
	}
	p = TSize(fBytes)
	return p.String()
}

// GetFloatSize 将浮点数大小转换为可读格式
func GetFloatSize(fBytes float64) string {
	var p TSize
	p = TSize(fBytes)
	return p.String()
}

// GetInt64Size 将Int64大小转换为可读格式
func GetInt64Size(iBytes int64) string {
	var p TSize
	p = TSize(iBytes)
	return p.String()
}

// GetIntSize 将Int大小转换为可读格式
func GetIntSize(iBytes int) string {
	var p TSize
	p = TSize(iBytes)
	return p.String()
}

// GetUInt64Size 将UInt64大小转换为可读格式
func GetUInt64Size(iBytes uint64) string {
	var p TSize
	p = TSize(iBytes)
	return p.String()
}

// Round 四舍五入
func Round(val float64, places int) float64 {
	var t float64
	f := math.Pow10(places)
	x := val * f
	if math.IsInf(x, 0) || math.IsNaN(x) {
		return val
	}
	if x >= 0.0 {
		t = math.Ceil(x)
		if (t - x) > 0.50000000001 {
			t -= 1.0
		}
	} else {
		t = math.Ceil(-x)
		if (t + x) > 0.50000000001 {
			t -= 1.0
		}
		t = -t
	}
	x = t / f

	if !math.IsInf(x, 0) {
		return x
	}

	return t
}

// GetTimeText 获取时间文本描述
func GetTimeText(fBytes int) string {
	m := 0.0
	if fBytes >= 60 {
		m = Round(float64(fBytes%60)/60.0, 1)
		fBytes = fBytes / 60
		if fBytes >= 60 {
			m = Round(float64(fBytes%60)/60.0, 1)
			fBytes = fBytes / 60
			if fBytes >= 24 {
				m = Round(float64(fBytes%24)/24.0, 1)
				fBytes = fBytes / 24
				iday := fBytes
				if fBytes >= 7 {
					m = Round(float64(fBytes%7)/7.0, 1)
					fBytes = fBytes / 7
					if fBytes >= 26 {
						m = Round(float64(iday%365)/365.0, 1)
						fBytes = iday / 365
						return fmt.Sprintf("%.1f年", Round(float64(fBytes)+m, 1))
					} else {
						return fmt.Sprintf("%.1f周", Round(float64(fBytes)+m, 1))
					}
				} else {
					return fmt.Sprintf("%.1f天", Round(float64(fBytes)+m, 1))
				}
			} else {
				return fmt.Sprintf("%.1f时", Round(float64(fBytes)+m, 1))
			}
		} else {
			return fmt.Sprintf("%.1f分", Round(float64(fBytes)+m, 1))
		}
	} else {
		return fmt.Sprintf("%d秒", fBytes)
	}
}

// IIFbyString 字符串类型的三元运算
func IIFbyString(flag bool, A, B string) string {
	if flag {
		return A
	}
	return B
}

// IIFByTime 时间类型的三元运算
func IIFByTime(flag bool, A, B time.Time) time.Time {
	if flag {
		return A
	}
	return B
}

// IIFbyInt 整数类型的三元运算
func IIFbyInt(flag bool, A, B int) int {
	if flag {
		return A
	}
	return B
}

// IIF 通用三元运算
func IIF(b bool, t, f interface{}) interface{} {
	if b {
		return t
	}
	return f
}

// LoadFileText 加载文本文件内容
func LoadFileText(file_name string) (string, error) {
	var msg string
	where := fmt.Sprintf("加载文件(%s)", file_name)
	out, err := ioutil.ReadFile(file_name)
	if err == nil {
		return string(out), nil
	} else {
		msg = where + "失败，" + err.Error()
		return "", errors.New(msg)
	}
}

// SetJson 解析JSON字符串到对象
// save_object对象一定是指针
func SetJson(json_text string, save_object interface{}) error {
	return json.Unmarshal([]byte(json_text), save_object)
}

// LoadJson 从文件加载JSON到对象
// save_object对象一定是指针
func LoadJson(json_file string, save_object interface{}) error {
	var msg string
	where := fmt.Sprintf("加载文件(%s)", json_file)
	out, err := ioutil.ReadFile(json_file)
	if err == nil {
		where = fmt.Sprintf("对文件(%s)反序列化", json_file)
		err = json.Unmarshal(out, save_object)
		if err != nil {
			msg = where + "失败，" + err.Error()
			return errors.New(msg)
		}
		return nil
	} else {
		msg = where + "失败，" + err.Error()
		return errors.New(msg)
	}
}

// GetJson 将对象转换为JSON字符串
func GetJson(object interface{}) string {
	s, _ := json.Marshal(object)
	var out bytes.Buffer
	json.Indent(&out, s, "", "  ")
	return string(out.Bytes())
}

// SaveJson 将对象保存为JSON文件
func SaveJson(json_file string, save_object interface{}) error {
	var msg string

	where := "传入对象json序列化"
	s, err := json.Marshal(save_object)
	if err != nil {
		msg = where + "失败，" + err.Error()
		return errors.New(msg)
	}

	where = "格式化json串"
	var out bytes.Buffer
	err = json.Indent(&out, s, "", "\t")
	if err != nil {
		msg = where + "失败，" + err.Error()
		return errors.New(msg)
	}

	where = fmt.Sprintf("保存%s文件", json_file)
	err = ioutil.WriteFile(json_file, out.Bytes(), 0666)
	if err != nil {
		msg = where + "失败，" + err.Error()
		return errors.New(msg)
	}
	return nil
}

// DosName 生成DOS兼容的文件名
// prepare-commit-～.sampl
func DosName(str string) string {
	fs := strings.Split(str, ".")
	file := fs[0]
	ext := filepath.Ext(str)

	file_name_len := 13
	if ext == "" {
		file_name_len = 17
	}

	if len(file) >= file_name_len {
		file = Substr(file, 0, file_name_len) + "~"
	}
	if len(ext) >= 4 {
		ext = Substr(ext, 0, 4)
	}

	return file + ext
}

// Substr 截取字符串
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

// 模版解析（可带子模板）
func ParseTemplates(wr io.Writer, tplname string, data interface{}) error {
	tplname = filepath.Join(RunPathName(), tplname)
	path := filepath.Dir(tplname)
	t := template.New(filepath.Base(tplname)).Delims("[[", "]]")
	t.ParseGlob(filepath.Join(path, "T.*.css"))
	t.ParseGlob(filepath.Join(path, "T.*.js"))
	t.ParseGlob(filepath.Join(path, "T.*.tpl"))
	t.ParseFiles(tplname)
	if err := t.Execute(wr, data); err != nil {
		fmt.Fprintf(wr, "解析模板(%s)失败，%s", tplname, err.Error())
		return err
	}
	return nil
}

// 模版解析（单文件）
func ParseTemplateFile(wr io.Writer, tplname string, data interface{}) error {
	tplname = filepath.Join(RunPathName(), tplname)
	b, err := ioutil.ReadFile(tplname)
	if err != nil {
		return fmt.Errorf("读取模板文件失败，%s", err.Error())
	}

	t, err := template.New(tplname).Delims("[[", "]]").Parse(string(b))
	if err != nil {
		return fmt.Errorf("解析模板文件失败，%s", err.Error())
	}

	err = t.Execute(wr, data)
	if err != nil {
		return fmt.Errorf("渲染模板文件失败，%s", err.Error())
	}

	return nil
}

// 是否可执行文件（linux就是x权限，win就是.exe)
func IsExeFile(path string) (bool, error) {
	if runtime.GOOS == "windows" {
		return strings.ToLower(GetFileExt(path)) == ".exe", nil
	} else {
		fi, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		return uint32(fi.Mode().Perm()&os.FileMode(73)) == uint32(73), nil
	}
}

// 获取文件修改时间 返回unix时间戳
func GetFileModTime(filename string) (time.Time, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// GetMapByString 从Map中获取字符串值
func GetMapByString(mapi map[string]string, name, default_value string) string {
	if v, ok := mapi[name]; ok {
		return v
	}
	return default_value
}

// GetMapByBool 从Map中获取布尔值
func GetMapByBool(mapi map[string]string, name string, default_value bool) bool {
	if v, ok := mapi[name]; ok {
		re := strings.ToLower(v)
		return re == "true" || re == "t" || re == "1"
	}
	return default_value
}

// GetMapByInt 从Map中获取整数值
func GetMapByInt(mapi map[string]string, name string, default_value int) int {
	if v, ok := mapi[name]; ok {
		re, err := strconv.Atoi(v)
		if err != nil {
			return default_value
		} else {
			return re
		}
	}
	return default_value
}

var TimeLocation *time.Location

// GetNow 获取当前时间（CST时区）
func GetNow() time.Time {
	location := time.FixedZone("CST", 8*60*60)
	return time.Now().In(location)
}

// GetLastYear 获取去年的今天
func GetLastYear() time.Time {
	return time.Now().AddDate(-1, 0, 0)
}

// GetTimeAgo 获取时间间隔描述
func GetTimeAgo(t time.Time) string {
	fb := int(GetNow().Unix() - t.Unix())
	if fb > 60*60*24*7 {
		return fmt.Sprintf("%d周前", fb/(60*60*24*7))
	} else if fb > 60*60*24 {
		return fmt.Sprintf("%d天前", fb/(60*60*24))
	} else if fb > 60*60 {
		return fmt.Sprintf("%d小时前", fb/(60*60))
	} else if fb > 60 {
		return fmt.Sprintf("%d分前", fb/60)
	} else {
		return fmt.Sprintf("%d秒前", fb)
	}
}

// StringToTime 字符串转时间
func StringToTime(data_str string) (time.Time, error) {
	u, err := time.ParseInLocation("2006-01-02 15:04:05", data_str, TimeLocation)
	if err != nil {
		return time.Time{}, err
	}

	return u, nil
}

// StringToTimeByTemplates 根据模板将字符串转时间
func StringToTimeByTemplates(tm, templates string) (time.Time, error) {
	t, err := time.ParseInLocation(templates, tm, TimeLocation)
	if nil == err && !t.IsZero() {
		return t, nil
	}

	return time.Time{}, err
}

// GetDebugStack 获取调试堆栈信息
func GetDebugStack() string {
	stacks := strings.Split(string(debug.Stack()), "\n")
	result := "当前堆栈："
	for _, v := range stacks {
		if strings.Contains(v, "goroutine") == false && strings.Contains(v, ":") {
			i := strings.Index(v, "+")
			if i > 0 {
				result = result + "\r\n==> " + strings.TrimSpace(v[:i])
			} else {
				result = result + "\r\n==> " + v
			}
		}
	}

	return result
}

// GetCallStack 获取调用堆栈信息
func GetCallStack() (level int, stack string, file string, line int) {
	loggerFuncCallDepth := 1 //要过滤掉当前函数的层数
	level = loggerFuncCallDepth
	stack = ""
	file = "???"
	line = 0
	for level <= 30 {
		pc, tfile, tline, ok := runtime.Caller(level)
		if ok {
			if level == loggerFuncCallDepth {
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
	if len(stack) > 2 {
		t1 := strings.Split(stack[:len(stack)-2], "->")
		return level - loggerFuncCallDepth, getClassName(t1[len(t1)-1]), file, line
	} else {
		return level - loggerFuncCallDepth, stack, file, line
	}
}

func getClassName(func_name string) string {
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

var LineEnding string // 根据系统自动设置

func init() {
	TimeLocation, _ = time.LoadLocation("Asia/Chongqing")
	if runtime.GOOS == "windows" {
		LineEnding = "\r\n"
	} else {
		LineEnding = "\n"
	}
}
