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

func IsAscii(text_str string) bool {
	temp := []rune(text_str)
	return (temp[0] >= '0' && temp[0] <= '9') || (temp[0] >= 'a' && temp[0] <= 'z') || (temp[0] >= 'A' && temp[0] <= 'Z')
}

func IsNumber(text_str string) bool {
	temp := []rune(text_str)
	return unicode.IsDigit(temp[0])
}

func IsHanZi(text_str string) bool {
	temp := []rune(text_str)
	if len(temp) >= 1 {
		return unicode.Is(unicode.Han, temp[0])
	} else {
		return false
	}
}

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

// RealPath converts the given <path> to its absolute path
// and checks if the file path exists.
// If the file does not exist, return an empty string.
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

// Dir returns all but the last element of path, typically the path's directory.
// After dropping the final element, Dir calls Clean on the path and trailing
// slashes are removed.
// If the `path` is empty, Dir returns ".".
// If the `path` is ".", Dir treats the path as current working directory.
// If the `path` consists entirely of separators, Dir returns a single separator.
// The returned path does not end in a separator unless it is the root directory.
func GetFileDir(path string) string {
	if path == "." {
		return filepath.Dir(GetFileRealPath(path))
	}
	return filepath.Dir(path)
}

// Mkdir creates directories recursively with given <path>.
// The parameter <path> is suggested to be an absolute path instead of relative one.
func Mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	return nil
}

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

// Basename returns the last element of path, which contains file extension.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns ".".
// If the path consists entirely of separators, Basename returns a single separator.
// Example:
// /var/www/file.js -> file.js
// file.js          -> file.js
func GetFileName(path string) string {
	return filepath.Base(path)
}

// Name returns the last element of path without file extension.
// Example:
// /var/www/file.js -> file
// file.js          -> file
func GetFileBaseName(path string) string {
	base := filepath.Base(path)
	if i := strings.LastIndexByte(base, '.'); i != -1 {
		return base[:i]
	}
	return base
}

func FileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func GetFileSize(filename string) int64 {
	if fi, err := os.Stat(filename); err == nil {
		return fi.Size()
	}
	return -1
}

func Md5(test_str string) string {
	hash := md5.New()
	hash.Write([]byte(test_str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Base64Encode(test_str string) string {
	encode_str := base64.StdEncoding.EncodeToString([]byte(test_str))
	return encode_str
}

func Base64Decode(test_str string) (string, error) {
	decode_str, err := base64.StdEncoding.DecodeString(test_str)
	if err != nil {
		return "", err
	}
	return string(decode_str), nil
}

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
func GetFloatSize(fBytes float64) string {
	var p TSize
	p = TSize(fBytes)
	return p.String()
}

func GetInt64Size(iBytes int64) string {
	var p TSize
	p = TSize(iBytes)
	return p.String()
}

func GetIntSize(iBytes int) string {
	var p TSize
	p = TSize(iBytes)
	return p.String()
}

func GetUInt64Size(iBytes uint64) string {
	var p TSize
	p = TSize(iBytes)
	return p.String()
}

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

func IIFbyString(flag bool, A, B string) string {
	if flag {
		return A
	}
	return B
}

func IIFByTime(flag bool, A, B time.Time) time.Time {
	if flag {
		return A
	}
	return B
}

func IIFbyInt(flag bool, A, B int) int {
	if flag {
		return A
	}
	return B
}

func IIF(b bool, t, f interface{}) interface{} {
	if b {
		return t
	}
	return f
}

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

// save_object对象一定是指针
func SetJson(json_text string, save_object interface{}) error {
	return json.Unmarshal([]byte(json_text), save_object)
}

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

func GetJson(object interface{}) string {
	s, _ := json.Marshal(object)
	var out bytes.Buffer
	json.Indent(&out, s, "", "  ")
	return string(out.Bytes())
}

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

func GetMapByString(mapi map[string]string, name, default_value string) string {
	if v, ok := mapi[name]; ok {
		return v
	}
	return default_value
}

func GetMapByBool(mapi map[string]string, name string, default_value bool) bool {
	if v, ok := mapi[name]; ok {
		re := strings.ToLower(v)
		return re == "true" || re == "t" || re == "1"
	}
	return default_value
}
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

func GetNow() time.Time {
	location := time.FixedZone("CST", 8*60*60)
	return time.Now().In(location)
}

func GetLastYear() time.Time {
	return time.Now().AddDate(-1, 0, 0)
}

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

func StringToTime(data_str string) (time.Time, error) {
	u, err := time.ParseInLocation("2006-01-02 15:04:05", data_str, TimeLocation)
	if err != nil {
		return time.Time{}, err
	}

	return u, nil
}

func StringToTimeByTemplates(tm, templates string) (time.Time, error) {
	t, err := time.ParseInLocation(templates, tm, TimeLocation)
	if nil == err && !t.IsZero() {
		return t, nil
	}

	return time.Time{}, err
}

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
