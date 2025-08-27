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
	inst := InitInstance("default", "./nustdb")

	inst.SetHead("TEST")
	inst.SetBucketName("List", 3)
	fmt.Println(inst.LPushByBucketName("List", "List01", "123123"))
}

func TestInitInstance(t *testing.T) {
	inst := InitInstance("default", "./nustdb")

	inst.SetHead("TEST")

	items, err := inst.GetAll("")
	if err != nil {
		t.Error(err)
	} else {
		PrintData("磁盘的数据", items)
	}

	inst.Set("NameA01", "Value01")
	inst.Set("NameA02", "Value02")
	inst.Set("NameA03", "Value03")
	inst.Set("NameA04", "Value04")
	inst.Set("NameA05", "Value05")

	items, err = inst.GetAll("")
	if err != nil {
		t.Error(err)
	} else {
		PrintData("初始化的数据", items)
	}

	inst.DelAll("")

	items, err = inst.GetAll("")
	if err != nil {
		t.Error(err)
	} else {
		PrintData("删除后的数据", items)
	}
}
