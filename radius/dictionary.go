package radius

import (
	"errors"
	"fmt"
	"sync"

	logs "github.com/tea4go/gh/log4go"
)

var builtinOnce sync.Once

// Builtin is the built-in dictionary. It is initially loaded with the
// attributes defined in RFC 2865 and RFC 2866.
var Builtin *TDictionary

func initDictionary() {
	Builtin = &TDictionary{}
	Builtin.NameItems = make(map[string]*TDictEntry)
}

// 字典编码
type TDictEntry struct {
	Id   byte
	Name string
	Func IAttributeCodec
}

type TDictionary struct {
	IdItems   [1069]*TDictEntry
	NameItems map[string]*TDictEntry
}

func (d *TDictionary) String() string {
	str_text := fmt.Sprintf("字典里共注册%d个属性编码:", len(d.NameItems))
	for _, v := range d.IdItems {
		if v == nil {
			continue
		}
		str_text = str_text + fmt.Sprintf("\n[%03d] %s = %s", v.Id, v.Name, v.Func.GetCodeName())
	}
	return str_text
}

// 注册属性
func (d *TDictionary) MustRegister(name string, t byte, codec IAttributeCodec) {
	if d.IdItems[t] != nil {
		panic(errors.New("属性已经注册过。"))
	}
	entry := &TDictEntry{
		Id:   t,
		Name: name,
		Func: codec,
	}
	d.IdItems[t] = entry
	d.NameItems[name] = entry

	logs.Debug("注册属性 ... [%03d] %-40s (%s)\n", t, name, codec.GetCodeName())

}

func (d *TDictionary) NewAttr(name string, value interface{}) (*TAttribute, error) {
	entry := d.NameItems[name]
	if entry == nil {
		return nil, errors.New("属性没有注册。")
	}
	t := entry.Id
	codec := entry.Func

	if transformer, ok := codec.(IAttributeTransformer); ok {
		transformed, err := transformer.Transform(value)
		if err != nil {
			return nil, errors.New("属性转码出错，" + err.Error())
		}
		return &TAttribute{AttrId: t, AttrValue: transformed}, nil
	}

	return &TAttribute{AttrId: t, AttrValue: value}, nil
}

func (d *TDictionary) GetName(t byte) (string, bool) {
	entry := d.IdItems[t]
	if entry == nil {
		return "", false
	}
	return entry.Name, true
}

func (d *TDictionary) GetIndex(name string) (byte, bool) {
	entry := d.NameItems[name]
	if entry == nil {
		return 0, false
	}
	return entry.Id, true
}

func (d *TDictionary) GetFunc(t byte) IAttributeCodec {
	entry := d.IdItems[t]
	if entry == nil {
		return AttributeUnknown
	}
	return entry.Func
}
