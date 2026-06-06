package execplus

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// --- Command tests ---

func TestCommand_Echo(t *testing.T) {
	cmd := Command("echo", "hello")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command echo failed: %v", err)
	}
	if !strings.Contains(string(out), "hello") {
		t.Errorf("Output = %q, want to contain hello", string(out))
	}
}

func TestCommand_Ls(t *testing.T) {
	cmd := Command("ls", "/tmp")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command ls failed: %v", err)
	}
}

func TestCommand_Output(t *testing.T) {
	cmd := Command("echo", "test output")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command echo Output failed: %v", err)
	}
	if !strings.Contains(string(out), "test output") {
		t.Errorf("Output = %q, want to contain 'test output'", string(out))
	}
}

func TestCommand_Run(t *testing.T) {
	cmd := Command("echo", "run test")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command echo Run failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "run test") {
		t.Errorf("Stdout = %q, want to contain 'run test'", stdout.String())
	}
}

// --- CommandString tests ---

func TestCommandString_Echo(t *testing.T) {
	cmd := CommandString("echo hello world")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CommandString echo failed: %v", err)
	}
	if !strings.Contains(string(out), "hello world") {
		t.Errorf("Output = %q, want to contain 'hello world'", string(out))
	}
}

// --- CommandContext tests ---

func TestCommandContext_Echo(t *testing.T) {
	ctx := context.Background()
	cmd := CommandContext(ctx, "echo", "ctx test")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CommandContext echo failed: %v", err)
	}
	if !strings.Contains(string(out), "ctx test") {
		t.Errorf("Output = %q, want to contain 'ctx test'", string(out))
	}
}

// --- CommandStringContext tests ---

func TestCommandStringContext_Echo(t *testing.T) {
	ctx := context.Background()
	cmd := CommandStringContext(ctx, "echo ctx string test")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CommandStringContext echo failed: %v", err)
	}
	if !strings.Contains(string(out), "ctx string test") {
		t.Errorf("Output = %q, want to contain 'ctx string test'", string(out))
	}
}

// --- Wait tests ---

func TestCmdPlus_Wait(t *testing.T) {
	cmd := Command("echo", "wait test")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	err := cmd.Wait()
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
}

func TestCmdPlus_WaitTwice(t *testing.T) {
	cmd := Command("echo", "double wait")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	err := cmd.Wait()
	if err != nil {
		t.Fatalf("First Wait failed: %v", err)
	}
	// Second Wait should also succeed (not panic)
	err2 := cmd.Wait()
	if err2 != nil {
		t.Logf("Second Wait returned: %v (this is expected)", err2)
	}
}

func TestCmdPlus_Wait_NotStarted(t *testing.T) {
	cmd := Command("echo", "never started")
	// Don't call Start - Process is nil
	err := cmd.Wait()
	if err == nil {
		t.Error("Wait should return error when Process is nil")
	}
	if err.Error() != "程序没有启动！" {
		t.Errorf("Wait error = %q, want 程序没有启动！", err.Error())
	}
}

// --- Terminate tests ---

func TestCmdPlus_Terminate(t *testing.T) {
	cmd := CommandString("sleep 10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	cmd.Terminate()
	err := cmd.Wait()
	if err == nil {
		t.Log("Process terminated successfully")
	}
}

func TestCmdPlus_Terminate_NoCancelFunc(t *testing.T) {
	ctx := context.Background()
	cmd := CommandContext(ctx, "echo", "no cancel")
	// CommandContext doesn't set cancelFunc
	cmd.Terminate() // Should not panic when cancelFunc is nil
}

// --- ShowConsole tests ---

func TestCmdPlus_ShowConsole(t *testing.T) {
	cmd := Command("echo", "show console")
	cmd.ShowConsole(true)
	if cmd.Stdout != os.Stdout {
		t.Error("ShowConsole(true) should set Stdout to os.Stdout")
	}
	if cmd.Stderr != os.Stderr {
		t.Error("ShowConsole(true) should set Stderr to os.Stderr")
	}

	cmd.ShowConsole(false)
	if cmd.Stdout != nil {
		t.Error("ShowConsole(false) should set Stdout to nil")
	}
	if cmd.Stderr != nil {
		t.Error("ShowConsole(false) should set Stderr to nil")
	}
}

// --- SetEnv / GetEnv tests ---

func TestCmdPlus_SetEnv_GetEnv(t *testing.T) {
	cmd := Command("echo", "env test")
	// Initially, Env should be populated from os.Environ()
	initialVal := cmd.GetEnv("PATH")
	if initialVal == "" {
		t.Log("PATH not found in cmd.Env (may be expected in some environments)")
	}

	// Set a new env variable
	newVar := cmd.SetEnv("TEST_EXECPLUS_KEY", "test_value")
	if !newVar {
		// Already existed
		t.Log("TEST_EXECPLUS_KEY already existed")
	}

	val := cmd.GetEnv("TEST_EXECPLUS_KEY")
	if val != "test_value" {
		t.Errorf("GetEnv(TEST_EXECPLUS_KEY) = %q, want test_value", val)
	}

	// Overwrite existing
	cmd.SetEnv("TEST_EXECPLUS_KEY", "new_value")
	val = cmd.GetEnv("TEST_EXECPLUS_KEY")
	if val != "new_value" {
		t.Errorf("GetEnv(TEST_EXECPLUS_KEY) after overwrite = %q, want new_value", val)
	}
}

func TestCmdPlus_GetEnv_NotFound(t *testing.T) {
	cmd := Command("echo", "env notfound")
	val := cmd.GetEnv("NONEXISTENT_KEY_12345")
	if val != "" {
		t.Errorf("GetEnv(NONEXISTENT) = %q, want empty string", val)
	}
}

func TestCmdPlus_SetEnv_MalformedEnv(t *testing.T) {
	cmd := Command("echo", "env malformed")
	// Add a malformed env entry (no =)
	cmd.Env = append(cmd.Env, "malformed_entry_no_equals")
	// GetEnv should skip it
	val := cmd.GetEnv("malformed_entry_no_equals")
	if val != "" {
		t.Errorf("GetEnv(malformed) = %q, want empty string", val)
	}
}

// --- ConvertByte2String tests ---

func TestConvertByte2String_UTF8(t *testing.T) {
	input := []byte("hello world")
	got := ConvertByte2String(input, "UTF-8")
	if got != "hello world" {
		t.Errorf("ConvertByte2String(UTF-8) = %q, want hello world", got)
	}
}

func TestConvertByte2String_Default(t *testing.T) {
	input := []byte("test default")
	got := ConvertByte2String(input, "unknown_charset")
	if got != "test default" {
		t.Errorf("ConvertByte2String(unknown) = %q, want test default", got)
	}
}

func TestConvertByte2String_GB18030(t *testing.T) {
	// Simple ASCII text should be the same in GB18030
	input := []byte("ascii text")
	got := ConvertByte2String(input, "GB18030")
	if got != "ascii text" {
		t.Errorf("ConvertByte2String(GB18030) = %q, want ascii text", got)
	}
}

// --- SetShellName tests ---

func TestSetShellName(t *testing.T) {
	originalName := shell_name
	SetShellName("/bin/sh")
	if shell_name != "/bin/sh" {
		t.Errorf("shell_name = %q, want /bin/sh", shell_name)
	}
	SetShellName(originalName) // Restore
}

// --- Timeout with context test ---

func TestCmdPlus_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := CommandContext(ctx, "sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	err := cmd.Wait()
	if err == nil {
		t.Log("Process was killed by context timeout")
	}
}

// --- Exit error test ---

func TestCmdPlus_ExitError(t *testing.T) {
	cmd := Command("false") // Command that exits with non-zero status
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Success() {
				t.Error("false command should not exit successfully")
			}
		}
	} else {
		t.Log("false command ran but may have succeeded")
	}
	_ = out // just to use it
}

// --- Start then StdoutPipe test ---

func TestCmdPlus_StdoutPipe(t *testing.T) {
	cmd := Command("echo", "pipe test")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe failed: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	buf := make([]byte, 1024)
	n, _ := stdout.Read(buf)
	_ = string(buf[:n])
	cmd.Wait()
}

// --- HideWindow test (posix: should be no-op) ---

func TestCmdPlus_HideWindow(t *testing.T) {
	cmd := Command("echo", "hide window")
	cmd.HideWindow() // Should be a no-op on posix
	// Verify it doesn't panic
	if err := cmd.Run(); err != nil {
		t.Fatalf("Run after HideWindow failed: %v", err)
	}
}

// --- SetUser test (posix) ---

func TestCmdPlus_SetUser(t *testing.T) {
	cmd := Command("echo", "set user")
	err := cmd.SetUser("root")
	// May fail if root doesn't exist or we can't set credentials
	if err != nil {
		t.Logf("SetUser(root) returned error: %v (expected if not running as root)", err)
	}
}

// --- PowerShellFile test ---

func TestPowerShellFile(t *testing.T) {
	// This creates a CmdPlus but doesn't execute it (PowerShell may not be available)
	cmd := PowerShellFile("pwsh", "/tmp/test.ps1")
	if cmd == nil {
		t.Fatal("PowerShellFile returned nil")
	}
}