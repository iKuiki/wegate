package wechat

import (
	"encoding/json"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/pkg/errors"
	"reflect"
)

type wechatSerialize struct{}

func (s *wechatSerialize) Serialize(param interface{}) (ptype string, p []byte, err error) {
	switch param.(type) {
	case wwdk.LoginChannelItem:
		ptype = "wwdk.LoginChannelItem"
	case datastruct.Contact:
		ptype = "datastruct.Contact"
	case datastruct.Message:
		ptype = "datastruct.Message"
	default:
		err = errors.New("unknown param type: " + reflect.TypeOf(param).Name())
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
	case "datastruct.Contact":
		var item datastruct.Contact
		err = json.Unmarshal(b, &item)
		param = item
	case "datastruct.Message":
		var item datastruct.Message
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
		"datastruct.Contact",
		"datastruct.Message",
	}
}
