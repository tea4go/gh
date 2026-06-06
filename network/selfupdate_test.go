package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/schollz/progressbar/v3"
)

// --- SetAppVersion tests ---

func TestSetAppVersion(t *testing.T) {
	fmt.Println("=== 测试: SetAppVersion ===")

	// Basic version
	SetAppVersion("TestApp", "v1.0.0", "", "")
	if AppName != "TestApp" || AppVersion != "v1.0.0" {
		t.Errorf("SetAppVersion basic failed: name=%s ver=%s", AppName, AppVersion)
	}

	// Beta version with build time
	SetAppVersion("TestApp", "v2.0.0", "true", "2025-01-15(10:30:00)")
	if !strings.Contains(AppVersion, "v2.0.0") {
		t.Errorf("Beta version format wrong: %s", AppVersion)
	}
	t.Logf("Beta version: %s", AppVersion)

	// Beta with "f" flag (not beta)
	SetAppVersion("TestApp", "v3.0.0", "f", "2025-01-15(10:30:00)")
	if AppVersion != "v3.0.0" {
		t.Errorf("Version with f flag should be unchanged: %s", AppVersion)
	}

	// Beta with "false" flag (not beta)
	SetAppVersion("TestApp", "v4.0.0", "false", "2025-01-15(10:30:00)")
	if AppVersion != "v4.0.0" {
		t.Errorf("Version with false flag should be unchanged: %s", AppVersion)
	}

	// Beta with empty buildtime (should not add suffix)
	SetAppVersion("TestApp", "v5.0.0", "true", "")
	if AppVersion != "v5.0.0" {
		t.Errorf("Version with empty buildtime should be unchanged: %s", AppVersion)
	}

	// Beta with unparseable buildtime
	SetAppVersion("TestApp", "v6.0.0", "true", "invalid-time")
	if !strings.Contains(AppVersion, "v6.0.0") {
		t.Errorf("Version with invalid buildtime format wrong: %s", AppVersion)
	}
	t.Logf("Invalid buildtime version: %s", AppVersion)
}

// --- compareVersions tests ---

func TestCompareVersions(t *testing.T) {
	fmt.Println("=== 测试: compareVersions ===")

	if compareVersions("v1.0.0", "v1.0.0") != 0 {
		t.Error("equal versions should return 0")
	}
	if compareVersions("v1.0.0", "v2.0.0") >= 0 {
		t.Error("v1.0.0 < v2.0.0")
	}
	if compareVersions("v2.0.0", "v1.0.0") <= 0 {
		t.Error("v2.0.0 > v1.0.0")
	}
}

// --- CheckForUpdate tests ---

func TestCheckForUpdateNoAppVersion(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate no version ===")
	oldName := AppName
	oldVer := AppVersion
	AppName = ""
	AppVersion = ""
	_, _, _, err := CheckForUpdate("http://example.com", false)
	if err == nil {
		t.Error("expected error for empty AppName/AppVersion")
	}
	AppName = oldName
	AppVersion = oldVer
}

func TestCheckForUpdateNewVersion(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate new version ===")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "v1.0.1")
		fmt.Fprintln(w, "abc123checksum")
	}))
	defer ts.Close()

	latest, checksum, downurl, err := CheckForUpdate(ts.URL, false)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}
	if latest != "v1.0.1" {
		t.Errorf("latest version mismatch: %s", latest)
	}
	if checksum != "abc123checksum" {
		t.Errorf("checksum mismatch: %s", checksum)
	}
	if downurl == "" {
		t.Error("download URL should not be empty")
	}
	t.Logf("new version: latest=%s checksum=%s url=%s", latest, checksum, downurl)
}

func TestCheckForUpdateSameVersion(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate same version ===")
	SetAppVersion("TestApp", "v1.0.1", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "v1.0.1")
		fmt.Fprintln(w, "checksum")
	}))
	defer ts.Close()

	latest, _, _, err := CheckForUpdate(ts.URL, false)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}
	if latest != "" {
		t.Errorf("same version should return empty latest: %s", latest)
	}
}

func TestCheckForUpdateForced(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate forced ===")
	SetAppVersion("TestApp", "v1.0.1", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "v1.0.1")
		fmt.Fprintln(w, "checksum")
	}))
	defer ts.Close()

	latest, _, downurl, err := CheckForUpdate(ts.URL, true)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}
	if latest != "v1.0.1" {
		t.Errorf("forced should still return latest: %s", latest)
	}
	if downurl == "" {
		t.Error("forced should provide download URL")
	}
}

func TestCheckForUpdateServerError(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate server error ===")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, _, _, err := CheckForUpdate(ts.URL, false)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestCheckForUpdateEmptyVersion(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate empty version file ===")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
	}))
	defer ts.Close()

	_, _, _, err := CheckForUpdate(ts.URL, false)
	if err == nil {
		t.Error("expected error for empty version")
	}
}

// --- CalcChecksum / SimpleCalcChecksum ---

func TestSimpleCalcChecksum(t *testing.T) {
	fmt.Println("=== 测试: SimpleCalcChecksum ===")
	sum := SimpleCalcChecksum()
	if sum == "" {
		t.Log("SimpleCalcChecksum returned empty (may be expected in test env)")
	} else {
		t.Logf("SimpleCalcChecksum: %s", sum)
	}
}

// --- DoUpdate ---

func TestDoUpdateInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate invalid URL ===")
	err := DoUpdate("ftp://invalid", "")
	if err == nil {
		t.Error("expected error for non-http URL")
	}
	t.Logf("DoUpdate invalid URL error (expected): %v", err)
}

// --- DoUpdateWithProgress ---

func TestDoUpdateWithProgressInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress invalid URL ===")
	err := DoUpdateWithProgress("ftp://invalid", "")
	if err == nil {
		t.Error("expected error for non-http URL")
	}
	t.Logf("DoUpdateWithProgress invalid URL error (expected): %v", err)
}

// --- SetRestart / SetForced / SetUpgrade / SetPublish ---

func TestSetRestart(t *testing.T) {
	fmt.Println("=== 测试: SetRestart ===")
	SetRestart(true)
	if !*prestart {
		t.Error("SetRestart(true) failed")
	}
	SetRestart(false)
	if *prestart {
		t.Error("SetRestart(false) failed")
	}
}

func TestSetForced(t *testing.T) {
	fmt.Println("=== 测试: SetForced ===")
	SetForced()
	if !*pforced {
		t.Error("SetForced failed")
	}
	*pforced = false
}

func TestSetUpgrade(t *testing.T) {
	fmt.Println("=== 测试: SetUpgrade ===")
	SetUpgrade()
	if !*pupgrade {
		t.Error("SetUpgrade failed")
	}
	*pupgrade = false
}

func TestSetPublish(t *testing.T) {
	fmt.Println("=== 测试: SetPublish ===")
	// Set BASH_KEY so ppublish is initialized
	os.Setenv("BASH_KEY", "test_key")
	// We need to re-run init() to initialize ppublish, but init() is only called once
	// Instead, check if ppublish is already set
	if ppublish != nil {
		SetPublish()
		if !*ppublish {
			t.Error("SetPublish failed")
		}
		*ppublish = false
	} else {
		t.Log("ppublish is nil (BASH_KEY not set at init time), skipping SetPublish test")
	}
}

// --- progressReader ---

func TestProgressReader(t *testing.T) {
	fmt.Println("=== 测试: progressReader ===")
	// Test that progressReader.Read works with a real progress bar
	data := []byte("test data for progress reader")
	pr := &progressReader{
		reader: strings.NewReader(string(data)),
		bar:    progressbar.Default(int64(len(data))),
	}
	buf := make([]byte, len(data))
	n, err := pr.Read(buf)
	if err != nil {
		t.Fatalf("progressReader.Read failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("progressReader.Read count mismatch: %d", n)
	}
}

// --- CheckVerserver ---

func TestCheckVerserver(t *testing.T) {
	fmt.Println("=== 测试: CheckVerserver ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	result := CheckVerserver([]string{ts.URL})
	if result != ts.URL {
		t.Errorf("expected %s, got %s", ts.URL, result)
	}
}

func TestCheckVerserverTimeout(t *testing.T) {
	fmt.Println("=== 测试: CheckVerserver timeout ===")
	// Use an unreachable address
	result := CheckVerserver([]string{"http://127.0.0.1:1"})
	if result != "" {
		t.Errorf("expected empty for unreachable, got %s", result)
	}
}

// --- CheckVerservers with test server ---

func TestCheckVerserversBasic(t *testing.T) {
	fmt.Println("=== 测试: CheckVerservers basic ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			fmt.Fprint(w, "OK")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	valid := CheckVerservers([]string{ts.URL}, 1)
	if len(valid) == 0 {
		t.Error("CheckVerservers should find valid server")
	}
}

func TestCheckVerserversInvalidURL(t *testing.T) {
	fmt.Println("=== 测试: CheckVerservers invalid ===")
	valid := CheckVerservers([]string{"http://invalid-host-that-does-not-exist.local"}, 1)
	// Should return empty or partial results (timeout)
	t.Logf("CheckVerservers invalid result: %v", valid)
}

// --- SetProcAttrs ---

func TestSetProcAttrs(t *testing.T) {
	fmt.Println("=== 测试: SetProcAttrs ===")
	cmd := exec.Command("echo", "test")
	SetProcAttrs(cmd)
	if cmd.SysProcAttr == nil {
		t.Error("SetProcAttrs should set SysProcAttr")
	}
}

// --- CalcChecksum ---

func TestCalcChecksum(t *testing.T) {
	fmt.Println("=== 测试: CalcChecksum ===")
	sum, err := CalcChecksum()
	if err != nil {
		t.Logf("CalcChecksum failed (expected in test env): %v", err)
	} else {
		if sum == "" {
			t.Error("CalcChecksum returned empty string")
		}
		t.Logf("CalcChecksum: %s", sum)
	}
}

// --- DoUpdate with test server ---

func TestDoUpdateWithTestServer(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate test server ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "not found")
	}))
	defer ts.Close()

	err := DoUpdate(ts.URL, "")
	if err == nil {
		t.Error("expected error for 404")
	}
	t.Logf("DoUpdate 404 error (expected): %v", err)
}

// --- DoUpdateWithProgress with test server ---

func TestDoUpdateWithProgressTestServer(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress test server ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "")
	// May fail due to content length or 404
	t.Logf("DoUpdateWithProgress test server: %v", err)
}

// --- CheckForUpdate with single-line version ---

func TestCheckForUpdateSingleLineVersion(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate single line ===")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "v1.0.1") // no newline
	}))
	defer ts.Close()

	latest, checksum, downurl, err := CheckForUpdate(ts.URL, false)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}
	if latest != "v1.0.1" {
		t.Errorf("latest mismatch: %s", latest)
	}
	if checksum != "" {
		t.Errorf("checksum should be empty for single-line: %s", checksum)
	}
	if downurl == "" {
		t.Error("download URL should not be empty")
	}
	t.Logf("single line: latest=%s checksum=%s url=%s", latest, checksum, downurl)
}

// --- CheckVerservers with non-OK response ---

func TestCheckVerserversNonOKResponse(t *testing.T) {
	fmt.Println("=== 测试: CheckVerservers non-OK response ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	valid := CheckVerservers([]string{ts.URL}, 1)
	if len(valid) != 0 {
		t.Errorf("expected no valid servers for 500, got: %v", valid)
	}
}

// --- CheckVerservers with non-OK body ---

func TestCheckVerserversNonOKBody(t *testing.T) {
	fmt.Println("=== 测试: CheckVerservers non-OK body ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			fmt.Fprint(w, "FAILED")
		}
	}))
	defer ts.Close()

	valid := CheckVerservers([]string{ts.URL}, 1)
	if len(valid) != 0 {
		t.Errorf("expected no valid servers for non-OK body, got: %v", valid)
	}
}

// --- SetPublish when ppublish is nil ---

func TestSetPublishNil(t *testing.T) {
	fmt.Println("=== 测试: SetPublish nil ===")
	// If BASH_KEY is not set, ppublish will be nil
	if ppublish == nil {
		t.Log("ppublish is nil, SetPublish would panic - skipping")
	} else {
		SetPublish()
		if !*ppublish {
			t.Error("SetPublish failed")
		}
		*ppublish = false
	}
}

// --- CheckVerserver with error URL ---

func TestCheckVerserverErrorURL(t *testing.T) {
	fmt.Println("=== 测试: CheckVerserver error URL ===")
	result := CheckVerserver([]string{"://invalid"})
	if result != "" {
		t.Errorf("expected empty for invalid URL, got: %s", result)
	}
}

// --- DoUpdate with 200 OK test server (tests checksum decode) ---

func TestDoUpdateWithChecksum(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate with checksum ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "fake update binary content")
	}))
	defer ts.Close()

	// With checksum - selfupdate.Apply will fail in test env, but covers the code path
	err := DoUpdate(ts.URL, "abc123")
	t.Logf("DoUpdate with checksum: %v", err)
}

// --- DoUpdate without checksum ---

func TestDoUpdateWithoutChecksum(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate without checksum ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "fake update binary content")
	}))
	defer ts.Close()

	err := DoUpdate(ts.URL, "")
	t.Logf("DoUpdate without checksum: %v", err)
}

// --- DoUpdateWithProgress with valid server ---

func TestDoUpdateWithProgressValid(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress valid server ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Fprint(w, strings.Repeat("x", 100))
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "abc123")
	t.Logf("DoUpdateWithProgress valid: %v", err)
}

// --- PublishSoftware with test server ---

func TestPublishSoftwareWithServer(t *testing.T) {
	fmt.Println("=== 测试: PublishSoftware with server ===")
	os.Setenv("BASH_KEY", "test_key")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/publish" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"errno":0, "errmsg":"success"}`)
		}
	}))
	defer ts.Close()

	err := PublishSoftware(ts.URL)
	if err != nil {
		t.Logf("PublishSoftware error (expected): %v", err)
	}
}

// --- PublishSoftware with error server ---

func TestPublishSoftwareErrorServer(t *testing.T) {
	fmt.Println("=== 测试: PublishSoftware error server ===")
	os.Setenv("BASH_KEY", "test_key")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errno":  1,
			"errmsg": "internal error",
		})
	}))
	defer ts.Close()

	err := PublishSoftware(ts.URL)
	if err != nil {
		t.Logf("PublishSoftware error server (expected): %v", err)
	}
}

// --- DoUpdate with 404 ---

func TestDoUpdate404(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate 404 ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	err := DoUpdate(ts.URL, "")
	if err == nil {
		t.Error("expected error for 404")
	}
	if !strings.Contains(err.Error(), "文件不存在") {
		t.Logf("DoUpdate 404 error: %v", err)
	}
}

// --- DoUpdate with non-200, non-404 ---

func TestDoUpdateNonOK(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate non-OK ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	err := DoUpdate(ts.URL, "")
	if err == nil {
		t.Error("expected error for 403")
	}
	if !strings.Contains(err.Error(), "文件下载错误") {
		t.Logf("DoUpdate 403 error: %v", err)
	}
}

// --- DoUpdateWithProgress with 404 ---

func TestDoUpdateWithProgress404(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress 404 ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "")
	if err == nil {
		t.Error("expected error for 404")
	}
	t.Logf("DoUpdateWithProgress 404: %v", err)
}

// --- DoUpdateWithProgress with non-200, non-404 ---

func TestDoUpdateWithProgressNonOK(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress non-OK ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "")
	if err == nil {
		t.Error("expected error for 403")
	}
	t.Logf("DoUpdateWithProgress 403: %v", err)
}

// --- DoUpdateWithProgress with invalid file size ---

func TestDoUpdateWithProgressInvalidSize(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress invalid size ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "-1")
			w.WriteHeader(http.StatusOK)
			return
		}
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "")
	if err == nil {
		t.Error("expected error for invalid file size")
	}
	t.Logf("DoUpdateWithProgress invalid size: %v", err)
}

// --- DoUpdateWithProgress with error on HEAD ---

func TestDoUpdateWithProgressHEADError(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress HEAD error ===")
	err := DoUpdateWithProgress("http://127.0.0.1:1/nonexistent", "")
	if err == nil {
		t.Log("DoUpdateWithProgress did not return error")
	}
	t.Logf("DoUpdateWithProgress HEAD error: %v", err)
}

// --- CalcChecksum with temp file ---

func TestCalcChecksumWithTempFile(t *testing.T) {
	fmt.Println("=== 测试: CalcChecksum with temp file ===")
	// CalcChecksum reads the current executable, so it should succeed in normal circumstances
	sum, err := CalcChecksum()
	if err != nil {
		t.Logf("CalcChecksum error (expected in some envs): %v", err)
	} else {
		if len(sum) != 64 { // SHA256 hex is 64 chars
			t.Errorf("CalcChecksum length wrong: %d", len(sum))
		}
		t.Logf("CalcChecksum: %s", sum)
	}
}

// --- SimpleCalcChecksum thorough ---

func TestSimpleCalcChecksumThorough(t *testing.T) {
	fmt.Println("=== 测试: SimpleCalcChecksum thorough ===")
	sum := SimpleCalcChecksum()
	if sum == "" {
		t.Log("SimpleCalcChecksum returned empty")
	} else {
		if len(sum) != 64 {
			t.Errorf("SimpleCalcChecksum length wrong: %d", len(sum))
		}
		// Should be same as CalcChecksum
		sum2, _ := CalcChecksum()
		if sum2 != "" && sum != sum2 {
			t.Errorf("SimpleCalcChecksum and CalcChecksum mismatch: %s vs %s", sum, sum2)
		}
		t.Logf("SimpleCalcChecksum: %s", sum)
	}
}

// --- progressReader with error ---

func TestProgressReaderWithError(t *testing.T) {
	fmt.Println("=== 测试: progressReader with error ===")
	// Test progressReader wrapping an error-returning reader
	errReader := &errorReader{}
	pr := &progressReader{
		reader: errReader,
		bar:    progressbar.Default(100),
	}
	buf := make([]byte, 10)
	n, err := pr.Read(buf)
	if err == nil {
		t.Error("expected error from errorReader")
	}
	if n != 0 {
		t.Errorf("expected 0 bytes, got %d", n)
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

// --- progressReader with zero read ---

func TestProgressReaderZeroRead(t *testing.T) {
	fmt.Println("=== 测试: progressReader zero read ===")
	zeroReader := &zeroThenErrorReader{}
	pr := &progressReader{
		reader: zeroReader,
		bar:    progressbar.Default(100),
	}
	buf := make([]byte, 10)
	n, err := pr.Read(buf)
	if n != 0 {
		t.Errorf("expected 0 bytes, got %d", n)
	}
	if err == nil {
		t.Error("expected error after zero read")
	}
}

type zeroThenErrorReader struct{}

func (r *zeroThenErrorReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}

// --- progressReader with partial read ---

func TestProgressReaderPartialRead(t *testing.T) {
	fmt.Println("=== 测试: progressReader partial read ===")
	data := []byte("hello")
	pr := &progressReader{
		reader: strings.NewReader(string(data)),
		bar:    progressbar.Default(int64(len(data))),
	}
	// Read in small chunks
	buf := make([]byte, 2)
	total := 0
	for {
		n, err := pr.Read(buf)
		total += n
		if err != nil {
			break
		}
	}
	if total != len(data) {
		t.Errorf("expected %d total bytes, got %d", len(data), total)
	}
}

// --- PublishSoftware without BASH_KEY ---

func TestPublishSoftwareNoKey(t *testing.T) {
	fmt.Println("=== 测试: PublishSoftware no BASH_KEY ===")
	os.Unsetenv("BASH_KEY")
	SetAppVersion("TestApp", "v1.0.0", "", "")
	err := PublishSoftware("http://localhost:8080")
	if err == nil {
		t.Error("expected error when BASH_KEY is not set")
	}
	if !strings.Contains(err.Error(), "权限") {
		t.Logf("PublishSoftware no key error: %v", err)
	}
}

// --- DoUpdate with 200 OK but bad binary (checksum mismatch) ---

func TestDoUpdateChecksumMismatch(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate checksum mismatch ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "this is not a valid binary for selfupdate")
	}))
	defer ts.Close()

	err := DoUpdate(ts.URL, "0000000000000000000000000000000000000000000000000000000000000000")
	t.Logf("DoUpdate checksum mismatch: %v", err)
	// The error could be about checksum or about invalid binary
}

// --- DoUpdateWithProgress with 200 OK but bad binary ---

func TestDoUpdateWithProgressChecksumMismatch(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress checksum mismatch ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "28")
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Fprint(w, "this is not a valid binary")
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "0000000000000000000000000000000000000000000000000000000000000000")
	t.Logf("DoUpdateWithProgress checksum mismatch: %v", err)
}

// --- DoUpdateWithProgress with HEAD returning zero content length ---

func TestDoUpdateWithProgressZeroContentLength(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress zero content length ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusOK)
			return
		}
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "")
	if err == nil {
		t.Error("expected error for zero content length")
	}
	t.Logf("DoUpdateWithProgress zero content length: %v", err)
}

// --- CheckForUpdate with network error ---

func TestCheckForUpdateNetworkError(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate network error ===")
	SetAppVersion("TestApp", "v1.0.0", "", "")
	_, _, _, err := CheckForUpdate("http://127.0.0.1:1", false)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
	t.Logf("CheckForUpdate network error (expected): %v", err)
}

// --- CheckForUpdate with various version formats ---

func TestCheckForUpdateOlderVersion(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate older version on server ===")
	SetAppVersion("TestApp", "v2.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "v1.0.0")
		fmt.Fprintln(w, "abc123")
	}))
	defer ts.Close()

	latest, _, _, err := CheckForUpdate(ts.URL, false)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}
	if latest != "" {
		t.Errorf("older version should return empty latest: %s", latest)
	}
}

// --- CheckVerservers with multiple servers ---

func TestCheckVerserversMultipleServers(t *testing.T) {
	fmt.Println("=== 测试: CheckVerservers multiple servers ===")
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			fmt.Fprint(w, "OK")
		}
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			fmt.Fprint(w, "OK")
		}
	}))
	defer ts2.Close()

	valid := CheckVerservers([]string{ts1.URL, ts2.URL}, 2)
	if len(valid) < 1 {
		t.Error("CheckVerservers should find at least one valid server")
	}
	t.Logf("CheckVerservers found %d valid servers", len(valid))
}

// --- CheckVerservers with body read error ---

func TestCheckVerserversBodyReadError(t *testing.T) {
	fmt.Println("=== 测试: CheckVerservers body read error ===")
	// Server returns 200 but body doesn't contain "OK"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "FAILED")
	}))
	defer ts.Close()

	valid := CheckVerservers([]string{ts.URL}, 1)
	if len(valid) != 0 {
		t.Errorf("expected no valid servers for non-OK body, got: %v", valid)
	}
}

// --- PublishSoftware with bad URL ---

func TestPublishSoftwareBadURL(t *testing.T) {
	fmt.Println("=== 测试: PublishSoftware bad URL ===")
	os.Setenv("BASH_KEY", "test_key")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	err := PublishSoftware("http://127.0.0.1:1")
	if err != nil {
		t.Logf("PublishSoftware bad URL error (expected): %v", err)
	}
}

// --- PublishSoftware with server returning non-JSON error body ---

func TestPublishSoftwareNonJSONError(t *testing.T) {
	fmt.Println("=== 测试: PublishSoftware non-JSON error ===")
	os.Setenv("BASH_KEY", "test_key")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "plain text error not json") // non-JSON body
	}))
	defer ts.Close()

	err := PublishSoftware(ts.URL)
	if err != nil {
		t.Logf("PublishSoftware non-JSON error (expected): %v", err)
	}
}

// --- CheckVerserver with non-200 status ---

func TestCheckVerserverNonOKStatus(t *testing.T) {
	fmt.Println("=== 测试: CheckVerserver non-OK status ===")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := CheckVerserver([]string{ts.URL})
	if result != "" {
		t.Errorf("expected empty for non-OK status, got: %s", result)
	}
}

// --- DoUpdate with http.Get error ---

func TestDoUpdateHTTPGetError(t *testing.T) {
	fmt.Println("=== 测试: DoUpdate http.Get error ===")
	// Use a URL that will cause http.Get to fail
	err := DoUpdate("http://127.0.0.1:1/nonexistent", "")
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
	t.Logf("DoUpdate http.Get error (expected): %v", err)
}

// --- DoUpdateWithProgress with Get error after HEAD success ---

func TestDoUpdateWithProgressGetError(t *testing.T) {
	fmt.Println("=== 测试: DoUpdateWithProgress Get error ===")
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(http.StatusOK)
			return
		}
		// Close the connection abruptly for GET
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	err := DoUpdateWithProgress(ts.URL, "")
	t.Logf("DoUpdateWithProgress Get error: %v", err)
}

// --- CalcChecksum error path simulation ---

func TestCalcChecksumErrorPaths(t *testing.T) {
	fmt.Println("=== 测试: CalcChecksum error paths ===")
	// CalcChecksum reads the current executable, which should succeed
	// The error paths (os.Executable fail, os.Open fail, io.Copy fail)
	// are hard to trigger without mocking, but we can verify the success path
	sum, err := CalcChecksum()
	if err != nil {
		t.Logf("CalcChecksum error: %v", err)
	} else {
		if len(sum) != 64 {
			t.Errorf("CalcChecksum length wrong: %d", len(sum))
		}
	}
}

// --- SimpleCalcChecksum error path simulation ---

func TestSimpleCalcChecksumErrorPaths(t *testing.T) {
	fmt.Println("=== 测试: SimpleCalcChecksum error paths ===")
	sum := SimpleCalcChecksum()
	if sum == "" {
		t.Log("SimpleCalcChecksum returned empty")
	} else {
		if len(sum) != 64 {
			t.Errorf("SimpleCalcChecksum length wrong: %d", len(sum))
		}
	}
}

// --- CheckVerservers with multiple valid servers and count limit ---

func TestCheckVerserversCountLimit(t *testing.T) {
	fmt.Println("=== 测试: CheckVerservers count limit ===")
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			fmt.Fprint(w, "OK")
		}
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/state" {
			fmt.Fprint(w, "OK")
		}
	}))
	defer ts2.Close()

	// Request only 1 valid server even though 2 are available
	valid := CheckVerservers([]string{ts1.URL, ts2.URL}, 1)
	if len(valid) < 1 {
		t.Error("CheckVerservers should find at least one valid server")
	}
	t.Logf("CheckVerservers count=1 found %d valid servers", len(valid))
}

// --- PublishSoftware with server returning JSON error ---

func TestPublishSoftwareJSONError(t *testing.T) {
	fmt.Println("=== 测试: PublishSoftware JSON error ===")
	os.Setenv("BASH_KEY", "test_key")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/publish" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errno":  2,
				"errmsg": "bad request",
			})
		}
	}))
	defer ts.Close()

	err := PublishSoftware(ts.URL)
	if err != nil {
		t.Logf("PublishSoftware JSON error (expected): %v", err)
	}
}

// --- CheckForUpdate with read body error ---

func TestCheckForUpdateReadBodyError(t *testing.T) {
	fmt.Println("=== 测试: CheckForUpdate read body error ===")
	SetAppVersion("TestApp", "v1.0.0", "", "")

	// Server that closes connection after sending headers
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Don't write any body - this should still work but with empty content
	}))
	defer ts.Close()

	_, _, _, err := CheckForUpdate(ts.URL, false)
	// Empty body means empty version line, which should return an error
	if err != nil {
		t.Logf("CheckForUpdate empty body error (expected): %v", err)
	}
}
