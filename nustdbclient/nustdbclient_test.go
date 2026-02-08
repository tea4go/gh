package nustdbclient

import (
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

func PrintData(title string, items []*TNustDBField) {
	fmt.Println("------------------------", title, "-------------------------")
	for _, v := range items {
		fmt.Println(v.Key, "=>", v.Value)
	}
}

func TestLPush(t *testing.T) {
	inst, err := InitInstance("default", "./nustdb", true)
	if err != nil {
		t.Fatal(err)
	}
	inst.SetHead("TEST")
	inst.LSetMaxSize(3)
	err = inst.LPushByBucket("List", "List01", "123123")
	fmt.Println(err)
}

func TestInitInstance(t *testing.T) {
	inst, err := InitInstance("default", "./nustdb", false)
	if err != nil {
		t.Fatal(err)
	}
	inst.SetHead("TEST")

	items, err := inst.GetAllValue("")
	if err != nil {
		t.Error(err)
	} else {
		PrintData("磁盘的数据", items)
	}

	_ = inst.SetValue("NameA01", "Value01")
	_ = inst.SetValue("NameA02", "Value02")
	_ = inst.SetValue("NameA03", "Value03")
	_ = inst.SetValue("NameA04", "Value04")
	_ = inst.SetValue("NameA05", "Value05")

	items, err = inst.GetAllValue("")
	if err != nil {
		t.Error(err)
	} else {
		PrintData("初始化的数据", items)
	}
}
