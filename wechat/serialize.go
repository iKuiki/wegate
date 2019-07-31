package wechat

import (
	"encoding/json"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/pkg/errors"
	"reflect"
	"wegate/wechat/wechatstruct"
)

type wechatSerialize struct{}

func (s *wechatSerialize) Serialize(param interface{}) (ptype string, p []byte, err error) {
	switch param.(type) {
	case wwdk.LoginChannelItem:
		ptype = "wwdk.LoginChannelItem"
	case datastruct.User:
		ptype = "datastruct.User"
	case datastruct.Contact:
		ptype = "datastruct.Contact"
	case datastruct.Message:
		ptype = "datastruct.Message"
	case []datastruct.Contact:
		ptype = "[]datastruct.Contact"
	case wechatstruct.SendMessageRespond:
		ptype = "wechatstruct.SendMessageRespond"
	case wwdk.WechatRunInfo:
		ptype = "wwdk.WechatRunInfo"
	case PluginDesc:
		ptype = "PluginDesc"
	default:
		err = errors.New("unknown param type: " + reflect.TypeOf(param).Name())
		// 此处务必要记得返回呀，不然就出大事了！
		return
	}
	p, err = json.Marshal(param)
	return
}

func (s *wechatSerialize) Deserialize(ptype string, b []byte) (param interface{}, err error) {
	switch ptype {
	case "wwdk.LoginChannelItem":
		var item wwdk.LoginChannelItem
		err = json.Unmarshal(b, &item)
		param = item
	case "datastruct.User":
		var item datastruct.User
		err = json.Unmarshal(b, &item)
		param = item
	case "datastruct.Contact":
		var item datastruct.Contact
		err = json.Unmarshal(b, &item)
		param = item
	case "datastruct.Message":
		var item datastruct.Message
		err = json.Unmarshal(b, &item)
		param = item
	case "[]datastruct.Contact":
		var item []datastruct.Contact
		err = json.Unmarshal(b, &item)
		param = item
	case "wechatstruct.SendMessageRespond":
		var item wechatstruct.SendMessageRespond
		err = json.Unmarshal(b, &item)
		param = item
	case "wwdk.WechatRunInfo":
		var item wwdk.WechatRunInfo
		err = json.Unmarshal(b, &item)
		param = item
	case "PluginDesc":
		var item PluginDesc
		err = json.Unmarshal(b, &item)
		param = item
	default:
		err = errors.New("unknown param type: " + ptype)
	}
	return
}

func (s *wechatSerialize) GetTypes() []string {
	return []string{
		"wwdk.LoginChannelItem",
		"datastruct.User",
		"datastruct.Contact",
		"datastruct.Message",
		"[]datastruct.Contact",
		"wechatstruct.SendMessageRespond",
		"wwdk.WechatRunInfo",
		"PluginDesc",
	}
}
