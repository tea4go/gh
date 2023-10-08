// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logs

import (
	"testing"
	"time"
)

// Try each log level in decreasing order of priority.
func testConsoleCalls(bl *TLogger) {
	bl.Emergency("emergency")
	time.Sleep(20 * time.Microsecond)
	bl.Alert("alert")
	time.Sleep(20 * time.Microsecond)
	bl.Critical("critical")
	time.Sleep(20 * time.Microsecond)
	bl.Error("error")
	time.Sleep(20 * time.Microsecond)
	bl.Warning("warning")
	time.Sleep(20 * time.Microsecond)
	bl.Notice("notice")
	time.Sleep(20 * time.Microsecond)
	bl.Info("Info")
	time.Sleep(20 * time.Microsecond)
	bl.Debug("debug")
	time.Sleep(20 * time.Microsecond)
}

func TestFormatLog(t *testing.T) {
	v := formatLog("test", "123123")
	t.Log(v)
	v = formatLog("test", "123123")
	t.Log(v)
	v = formatLog("test", 123, 123)
	t.Log(v)
}
