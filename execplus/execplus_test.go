package execplus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {

	//测试使用ENV变量与中止运行
	//test_exec_env()

	m.Run()
}

func Test_CombinedOutput(t *testing.T) {
	//cmd := CommandString("/opt/bin/rebootex")
	cmd := CommandString(`ipconfig|findstr IPv4 `)
	//cmd := CommandString("ping -c 5 -i 1 192.168.50.1")
	//cmd := Command("ping", "127.0.0.1", "-c 2", "-i 1")
	//cmd := exec.Command("./cmd_demo/cmd_demo", "err")
	//cmd := exec.Command("./cmd_demo/core_demo", "`err`")
	//cmd_str := "scp -r -p /home/share/ansible root@127.0.0.1:~/"
	//cmd_args := strings.Split(cmd_str, " ")
	//cmd := exec.Command(cmd_args[0], cmd_args[1:]...)
	//cmd := CommandString("scp -r -p /home/share/ansible root@127.0.0.1:~/")
	fmt.Println("==============================")
	fmt.Println(cmd.Args)
	fmt.Println("==============================")

	fmt.Println("通过CombinedOutput执行：")
	fmt.Println("==============================")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("执行命令出错，原因：", err.Error())
		fmt.Println(string(out))
	} else {
		fmt.Printf("%s", ConvertByte2String(out, "GB18030"))
	}
}

func Test_exec_mode(t *testing.T) {
	//常用的exec使用例子
	test_exec("CombinedOutput")
	// test_exec("Output")
	// test_exec("bufio")
	// test_exec("bytes")
	//test_exec_WithCancel()
}
func Test_exec_env(t *testing.T) {
	var err error
	stopChan := make(chan bool)
	w := bytes.NewBuffer(nil)

	//cmd := CommandString("set|grep GOROOT && ping 127.0.0.1 -c 5 -t 1")
	//cmd_str := "echo 将7秒后重启机器！ && ping 127.0.0.1 -n 3 >nvl && echo 开始执行！ && shutdown -r -t 1 && echo 执行成功！"
	cmd_str := "echo 当前IP地址列表 && ipconfig|findstr IPv4"
	cmd_run := CommandString(cmd_str)
	cmd_run.Stderr = os.Stdout
	cmd_run.Stdout = os.Stdout
	fmt.Println("=======================")
	fmt.Println("=", cmd_run.Args)
	fmt.Println("=======================")

	go func() {
		err = cmd_run.Run()
		fmt.Println("end")
		stopChan <- true
	}()

	//执行命令，超时退出（5秒）
	ticker := time.NewTicker(5 * time.Second)
	stop := false
	for stop == false {
		select {
		case <-ticker.C:
			fmt.Println("超时退出")
			cmd_run.Terminate()
			stop = true
			break
		case stop = <-stopChan:
			break
		}
	}
	//判断ssh命令是否执行失败，通过返回码判定。
	if err != nil || (cmd_run.ProcessState != nil && !cmd_run.ProcessState.Success()) {
		fmt.Println(cmd_run.ProcessState.Success())
		fmt.Println("失败", err, string(w.Bytes()))
	} else {
		fmt.Println("成功", string(w.Bytes()))
	}
}

func test_exec(test_mode string) {
	//cmd := CommandString("/opt/bin/rebootex")
	cmd := CommandString(`ipconfig|findstr IPv4 `)
	//cmd := CommandString("ping -c 5 -i 1 192.168.50.1")
	//cmd := Command("ping", "127.0.0.1", "-c 2", "-i 1")
	//cmd := exec.Command("./cmd_demo/cmd_demo", "err")
	//cmd := exec.Command("./cmd_demo/core_demo", "`err`")
	//cmd_str := "scp -r -p /home/share/ansible root@127.0.0.1:~/"
	//cmd_args := strings.Split(cmd_str, " ")
	//cmd := exec.Command(cmd_args[0], cmd_args[1:]...)
	//cmd := CommandString("scp -r -p /home/share/ansible root@127.0.0.1:~/")
	fmt.Println("==============================")
	fmt.Println(cmd.Args)
	fmt.Println("==============================")

	if test_mode == "CombinedOutput" {
		fmt.Println("通过CombinedOutput执行：")
		fmt.Println("==============================")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("执行命令出错，原因：", err.Error())
			fmt.Println(string(out))
		} else {
			fmt.Printf("%s", ConvertByte2String(out, "GB18030"))
		}
	} else if test_mode == "Output" {
		fmt.Println("通过Output执行")
		out, err := cmd.Output()
		if err != nil {
			if nerr, ok := err.(*exec.ExitError); ok {
				fmt.Println("执行命令出错，原因：", string(nerr.Stderr))
			} else {
				fmt.Println("执行命令出错，原因：", err.Error())
			}
		} else {
			fmt.Printf("%s", ConvertByte2String(out, "GB18030"))
		}
	} else if test_mode == "Run" {
		fmt.Println("通过Run运行命令(需要执行完才能得到结果，通过指定Stdout/Stderr写入器[io.Writer])")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout
		err := cmd.Run()
		if err != nil {
			fmt.Printf("执行命令出错，原因：%s\n", err.Error())
			fmt.Printf("%s", ConvertByte2String(stdout.Bytes(), "GB18030"))
		} else {
			fmt.Printf("%s", ConvertByte2String(stdout.Bytes(), "GB18030"))
		}
	} else if test_mode == "bufio" {
		fmt.Println("通过自带的Pipe（bufio）")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println("获取标准控制台失败，原因 ：", err.Error())
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Println("获取错误控制台失败，原因 ：", err.Error())
			return
		}
		if err = cmd.Start(); err != nil {
			fmt.Println("执行命令出错，原因：", err.Error())
			return
		}
		go func() {
			outputBuf := bufio.NewReader(stdout)
			bufs := make([]byte, 1024)
			for {
				n, err := outputBuf.Read(bufs)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("执行报错，原因：%s\n", err.Error())
					}
					break
				}
				fmt.Printf("%s", ConvertByte2String(bufs[:n], "GB18030"))
			}
		}()
		go func() {
			outputBuf := bufio.NewReader(stderr)
			bufs := make([]byte, 1024)
			for {
				n, err := outputBuf.Read(bufs)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("执行报错，原因：%s\n", err.Error())
					}
					break
				}
				fmt.Printf("%s", ConvertByte2String(bufs[:n], "GB18030"))
			}
		}()
		cmd.Wait()
	} else if test_mode == "bytes" {
		fmt.Println("通过自带的Pipe（bytes）")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println("获取标准控制台失败，原因 ：", err.Error())
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Println("获取错误控制台失败，原因 ：", err.Error())
			return
		}
		if err = cmd.Start(); err != nil {
			fmt.Println("执行命令出错，原因：", err.Error())
			return
		}
		go func() {
			bufs := make([]byte, 1024)
			for {
				n, err := stdout.Read(bufs)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("执行报错，原因：%s\n", err.Error())
					}
					break
				}
				fmt.Printf("%s", ConvertByte2String(bufs[:n], "GB18030"))
			}
		}()
		go func() {
			bufs := make([]byte, 1024)
			for {
				n, err := stderr.Read(bufs)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("执行报错，原因：%s\n", err.Error())
					}
					break
				}
				fmt.Printf("%s", ConvertByte2String(bufs[:n], "GB18030"))
			}
		}()
		cmd.Wait()
	}
	time.Sleep(1 * time.Second)
}
