package main

import (
	logs "github.com/tea4go/gh/log4go"
)

func main() {
	logs.SetFDebug(true)
	log := logs.NewLogger()
	log.SetLogFuncCallDepth(3)
	log.SetLogger("console", `{"color":true,"level":7}`)
	log.SetLevel(logs.LevelDebug)

	log.Emergency("TestLog (Emergency)")
	log.Alert("TestLog (Alert)")
	log.Critical("TestLog (Critical)")
	log.Error("TestLog (Error)")
	log.Warning("TestLog (Warning)")
	log.Notice("TestLog (Notice)")
	log.Info("TestLog (Info)")
	log.Debug("TestLog (Debug)")
	log.Notice("--------------------------")

}
