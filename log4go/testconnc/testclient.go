package main

import (
	"time"

	logs "github.com/tea4go/gh/log4go"
)

func main() {
	log := logs.NewLogger()
	log.SetLogFuncCallDepth(3)
	log.SetLogger("conn", `{"net":"tcp","reconnect":true,"level":7,"addr":"127.0.0.1:6002"}`)

	for i := 0; i < 100; i++ {
		log.Debug("============================")
		time.Sleep(20 * time.Microsecond)
		//log.Emergency("TestLog %04d(Emergency)", i)
		//log.Alert("TestLog %04d(Alert)", i)
		//log.Critical("TestLog %04d(Critical)", i)
		//log.Error("TestLog %04d(Error)", i)
		log.Warning("TestLog %04d(Warning)", i)
		time.Sleep(20 * time.Microsecond)
		log.Notice("TestLog %04d(Notice)", i)
		time.Sleep(20 * time.Microsecond)
		log.Info("TestLog %04d(Info)", i)
		time.Sleep(20 * time.Microsecond)
		log.Debug("TestLog %04d(Debug)", i)
		time.Sleep(1 * time.Second)
	}
}
