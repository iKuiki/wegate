package wechat

import (
	"encoding/json"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils/uuid"
	"wegate/common"
)

// 提供mqtt client微信子模块注册维护工作

type mqttCaller interface {
	Send(topic string, payload []byte) string
}

func (m *Wechat) registerMQTTPlugin(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	name, description := msg["name"].(string), msg["description"].(string)
	// 检查mqttPlugin是否符合规范
	if name == "" || description == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "module name or description is empty",
		}
		return
	}
	log.Info("新MQTT Plugin注册：%s[%s]", name, description)
	var (
		loginListenerTopic   string
		contactListenerTopic string
		msgListenerTopic     string
	)
	plugin := &mqttPlugin{
		name:                 name,
		description:          description,
		loginListenerTopic:   loginListenerTopic,
		contactListenerTopic: contactListenerTopic,
		msgListenerTopic:     msgListenerTopic,
		caller:               session,
	}
	token := uuid.Rand().Hex()
	plugin.loginStatus(m.loginStatus)
	m.pluginMap[token] = plugin
	result.Ret = common.RetCodeOK
	result.Msg = token
	return
}

// mqttPlugin 对外公开的rpcPlugin插件的地址
type mqttPlugin struct {
	name                 string
	description          string
	loginListenerTopic   string
	contactListenerTopic string
	msgListenerTopic     string
	caller               mqttCaller
}

func (p *mqttPlugin) getName() string {
	return p.name
}

func (p *mqttPlugin) getDescription() string {
	return p.description
}

func (p *mqttPlugin) loginStatus(loginStatus wwdk.LoginChannelItem) {
	if p.loginListenerTopic != "" {
		// 如果监听，则发送消息
		payload, err := json.Marshal(loginStatus)
		if err != nil {
			log.Error("marshal loginStatus to json error: %v", err)
		} else {
			p.caller.Send(p.loginListenerTopic, payload)
		}
	}
	return
}

func (p *mqttPlugin) modifyContact(contact datastruct.Contact) {
	if p.contactListenerTopic != "" {
		// 如果监听，则发送消息
		payload, err := json.Marshal(contact)
		if err != nil {
			log.Error("marshal contact to json error: %v", err)
		} else {
			p.caller.Send(p.contactListenerTopic, payload)
		}
	}
	return
}

func (p *mqttPlugin) newMessage(msg datastruct.Message) {
	if p.msgListenerTopic != "" {
		// 如果监听，则发送消息
		payload, err := json.Marshal(msg)
		if err != nil {
			log.Error("marshal msg to json error: %v", err)
		} else {
			p.caller.Send(p.msgListenerTopic, payload)
		}
	}
	return
}
