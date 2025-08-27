package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	logs "github.com/tea4go/gh/log4go"
)

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	// 创建一个接收信号的channel
	sigs := make(chan os.Signal, 1)
	// 使用signal包中的Notify函数来监听SIGINT信号
	signal.Notify(sigs, syscall.SIGINT)

	debug := flag.Bool("debug", false, "是否显示调试信息")
	loglevel := flag.Int("l", logs.LevelNotice, "日志级别")
	host := flag.String("host", "127.0.0.1", "设置日志服务器地址。")
	port := flag.Int("port", 9514, "设置日志服务器端口。")
	thread := flag.Int("t", 1, "设置多少线程数")
	batch := flag.Int("b", 1, "设置写多少批次日志。")
	flag.Usage = func() {
		fmt.Printf("使用方法： %s [参数 ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// 创建一个goroutine来处理信号
	go func() {
		<-sigs
		fmt.Println("-= 退出 =-")
		os.Exit(0)
	}()

	logs.SetFDebug(*debug)
	log := logs.NewLogger()
	log.SetLogFuncCallDepth(3)
	log.SetLogger("console")
	log.SetLogger("conn", fmt.Sprintf(`{"addr":"%s:%d","level":5,"name":"testlog"}`, *host, *port))
	//log.SetLevel(*loglevel, "console")
	log.SetLevel(*loglevel)

	fmt.Printf("Start Log4go Client ...... (%s:%d)\n", *host, *port)
	fmt.Printf("= 开启 %d 个线程，每个线程执行 %d 批次日志。\n", *thread, *batch)
	fmt.Println("...... 请按 Ctrl+C 结果 ......")
	fmt.Println("====================================================================")

	var wg sync.WaitGroup
	wg.Add(*thread * *batch)
	for b := 1; b <= *thread; b++ {
		go func(b int) {
			for i := 1; i <= *batch; i++ {
				fmt.Printf("= 第%d线程，第%d批测试\n", b, i)
				log.Emergency("TestLog %04d-%04d(Emergency)", b, i)
				log.Alert("TestLog %04d-%04d(Alert)", b, i)
				log.Critical("TestLog %04d-%04d(Critical)", b, i)
				log.Error("TestLog %04d-%04d(Error)", b, i)
				log.Warning("TestLog %04d-%04d(Warning)", b, i)
				log.Notice("TestLog %04d-%04d(Notice)", b, i)
				log.Info("TestLog %04d-%04d(Info)", b, i)
				log.Debug("TestLog %04d-%04d(Debug)", b, i)
				log.Notice("--------------------------")
				wg.Done()
				time.Sleep(time.Duration(rand.Intn(1000)+1000) * time.Millisecond)
			}
		}(b)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
	wg.Wait()
	log.Close()
}
