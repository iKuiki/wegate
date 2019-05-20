package wechat

import (
	"encoding/json"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/gate"
	"time"
	"wegate/common"
)

/*
Wechat Plugin插件
每个插件代表一个Wechat的功能模块
插件有自己的名称，有自己的描述
插件可以附带监听器来获取wechat的新消息
插件注册时会生成一个token，在主动调用Wechat（发信息等）时需要附带token以验证
*/

// Plugin 插件
type Plugin interface {
	getName() string
	getDescription() string
	getPluginType() PluginType
	getCreatedAt() time.Time
	// loginStatusEvent 登陆状态变化时的事件推送
	loginStatusEvent(loginStatus wwdk.LoginChannelItem)
	// modifyContactEvent 联系人发生修改时的事件推送
	modifyContactEvent(contact datastruct.Contact)
	// newMessageEvent 接受到新消息时的事件推送
	newMessageEvent(msg datastruct.Message)
	// 本来插件变更应该是使用同一条通道的，但是那样的话需要在传输的信息中再标注是添加还是移除
	// 为了减少struct数量，所以把插件的变更分为添加和移除2条通道
	// addPluginEvent 新插件注册时的事件推送
	addPluginEvent(pluginDesc PluginDesc)
	// removePluginEvent 插件发生卸载时的事件推送
	removePluginEvent(pluginDesc PluginDesc)
}

// PluginType 插件类型
type PluginType int32

const (
	// PluginTypeRPCPlugin RPC插件
	PluginTypeRPCPlugin PluginType = 1
	// PluginTypeMQTTPlugin MQTT插件
	PluginTypeMQTTPlugin PluginType = 2
)

// PluginDesc 插件描述
type PluginDesc struct {
	Name        string
	Description string
	PluginType  PluginType
	CreatedAt   time.Time
}

// mqttGetPluginList mqtt客户端获取已注册的插件列表
func (m *Wechat) mqttGetPluginList(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeUnauthorized,
			Msg: "need login",
		}
		return
	}
	resp, eStr := m.getPluginList(
		common.ForceString(msg["token"]),
	)
	if eStr != "" {
		result = common.Response{
			Ret: common.RetCodeServerError,
			Msg: eStr,
		}
		return
	}
	payload, e := json.Marshal(resp)
	if e != nil {
		result = common.Response{
			Ret: common.RetCodeServerError,
			Msg: e.Error(),
		}
		return
	}
	result = common.Response{
		Ret: common.RetCodeOK,
		Msg: string(payload),
	}
	return
}

// getPluginList 获取已注册的插件列表
func (m *Wechat) getPluginList(token string) (list []PluginDesc, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	// 复制一份pluginMap防止发生同时读写map的异常
	pMap := m.pluginMap
	for _, p := range pMap {
		list = append(list, PluginDesc{
			Name:        p.getName(),
			Description: p.getDescription(),
			PluginType:  p.getPluginType(),
			CreatedAt:   p.getCreatedAt(),
		})
	}
	return
}
