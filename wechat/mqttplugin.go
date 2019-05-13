package wechat

import (
	"encoding/json"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils/uuid"
	"time"
	"wegate/common"
)

// 提供mqtt client微信子模块注册维护工作

type mqttCaller interface {
	Send(topic string, payload []byte) string
}

func (m *Wechat) registerMQTTPlugin(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeUnauthorized,
			Msg: "need login",
		}
		return
	}
	// 检查此client是否有注册WechatUploader
	if token := session.Get("WechatPluginToken"); token != "" {
		log.Debug("检测到session尝试重复注册Plugin")
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "duplicate registered",
		}
		return
	}
	name, description := common.ForceString(msg["name"]), common.ForceString(msg["description"])
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
		loginListenerTopic        = common.ForceString(msg["loginListenerTopic"])
		contactListenerTopic      = common.ForceString(msg["contactListenerTopic"])
		msgListenerTopic          = common.ForceString(msg["msgListenerTopic"])
		addPluginListenerTopic    = common.ForceString(msg["addPluginListenerTopic"])
		removePluginListenerTopic = common.ForceString(msg["removePluginListenerTopic"])
	)
	token := uuid.Rand().Hex()
	session.Set("WechatPluginToken", token)
	eStr := session.Push()
	if eStr != "" {
		log.Error("推送session失败: %s", eStr)
		result = common.Response{
			Ret: common.RetCodeServerError,
			Msg: "push session fail",
		}
		return
	}
	plugin := &mqttPlugin{
		name:                      name,
		description:               description,
		loginListenerTopic:        loginListenerTopic,
		contactListenerTopic:      contactListenerTopic,
		msgListenerTopic:          msgListenerTopic,
		addPluginListenerTopic:    addPluginListenerTopic,
		removePluginListenerTopic: removePluginListenerTopic,
		caller:                    session,
		createdAt:                 time.Now(),
	}
	// 广播新插件注册消息
	pMap := m.pluginMap
	for _, p := range pMap {
		go func(p Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send add plugin message panic: %+v", e)
				}
			}()
			pDesc := PluginDesc{
				Name:        plugin.getName(),
				Description: plugin.getDescription(),
				PluginType:  plugin.getPluginType(),
				CreatedAt:   plugin.getCreatedAt(),
			}
			p.addPlugin(pDesc)
		}(p)
	}
	plugin.loginStatus(m.loginStatus)
	m.pluginMap[token] = plugin
	result.Ret = common.RetCodeOK
	result.Msg = token
	return
}

func (m *Wechat) disconnectMQTTPlugin(token string) (result common.Response, err string) {
	plugin, ok := m.pluginMap[token]
	if !ok {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "plugin not found",
		}
		return
	}
	delete(m.pluginMap, token)
	log.Debug("已卸载Plugin[%s]: %s", plugin.getName(), plugin.getDescription())
	// 广播插件卸载消息
	pMap := m.pluginMap
	for _, p := range pMap {
		go func(p Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send add plugin message panic: %+v", e)
				}
			}()
			pDesc := PluginDesc{
				Name:        plugin.getName(),
				Description: plugin.getDescription(),
				PluginType:  plugin.getPluginType(),
				CreatedAt:   plugin.getCreatedAt(),
			}
			p.removePlugin(pDesc)
		}(p)
	}
	result = common.Response{
		Ret: common.RetCodeOK,
	}
	return
}

func (m *Wechat) callWechat(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeUnauthorized,
			Msg: "need login",
		}
		return
	}
	// token后续会检查，所以此处可以暂不检查
	fnName := common.ForceString(msg["fnName"])
	var (
		resp interface{}
		eStr string
	)
	switch fnName {
	case "SendTextMessage":
		// sendTextMessage(token string, toUserName, content string) (result wechatstruct.SendMessageRespond, err string)
		resp, eStr = m.sendTextMessage(
			common.ForceString(msg["token"]),
			common.ForceString(msg["toUserName"]),
			common.ForceString(msg["content"]),
		)
	case "RevokeMessage":
		// revokeMessage(token string, srvMsgID, localMsgID, toUserName string) (result wechatstruct.RevokeMessageRespond, err string)
		resp, eStr = m.revokeMessage(
			common.ForceString(msg["token"]),
			common.ForceString(msg["srvMsgID"]),
			common.ForceString(msg["localMsgID"]),
			common.ForceString(msg["toUserName"]),
		)
	case "GetUser":
		// getContactList(token string) (result []datastruct.Contact, err string)
		resp, eStr = m.getUser(
			common.ForceString(msg["token"]),
		)
	case "GetContactList":
		// getContactList(token string) (result []datastruct.Contact, err string)
		resp, eStr = m.getContactList(
			common.ForceString(msg["token"]),
		)
	case "GetContactByUserName":
		// getContactByUserName(token string, userName string) (result datastruct.Contact, err string)
		resp, eStr = m.getContactByUserName(
			common.ForceString(msg["token"]),
			common.ForceString(msg["userName"]),
		)
	case "GetContactByAlias":
		// getContactByAlias(token string, alias string) (result datastruct.Contact, err string)
		resp, eStr = m.getContactByAlias(
			common.ForceString(msg["token"]),
			common.ForceString(msg["alias"]),
		)
	case "GetContactByNickname":
		// getContactByNickname(token string, nickname string) (result datastruct.Contact, err string)
		resp, eStr = m.getContactByNickname(
			common.ForceString(msg["token"]),
			common.ForceString(msg["nickname"]),
		)
	case "GetContactByRemarkName":
		// getContactByRemarkName(token string, remarkName string) (result datastruct.Contact, err string)
		resp, eStr = m.getContactByRemarkName(
			common.ForceString(msg["token"]),
			common.ForceString(msg["remarkName"]),
		)
	case "ModifyUserRemarkName":
		// modifyUserRemarkName(token string, userName, remarkName string) (result, err string)
		resp, eStr = m.modifyUserRemarkName(
			common.ForceString(msg["token"]),
			common.ForceString(msg["userName"]),
			common.ForceString(msg["remarkName"]),
		)
	case "ModifyChatRoomTopic":
		// modifyChatRoomTopic(token string, userName, newTopic string) (result, err string)
		resp, eStr = m.modifyChatRoomTopic(
			common.ForceString(msg["token"]),
			common.ForceString(msg["userName"]),
			common.ForceString(msg["newTopic"]),
		)
	case "GetRunInfo":
		// getRunInfo(token string) (result wwdk.WechatRunInfo, err string)
		resp, eStr = m.getRunInfo(
			common.ForceString(msg["token"]),
		)
	default:
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "func not found",
		}
		return
	}
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

// mqttPlugin 对外公开的rpcPlugin插件的地址
type mqttPlugin struct {
	name                      string
	description               string
	loginListenerTopic        string
	contactListenerTopic      string
	msgListenerTopic          string
	addPluginListenerTopic    string
	removePluginListenerTopic string
	caller                    mqttCaller
	createdAt                 time.Time
}

func (p *mqttPlugin) getName() string {
	return p.name
}

func (p *mqttPlugin) getDescription() string {
	return p.description
}

func (p *mqttPlugin) getPluginType() PluginType {
	return PluginTypeMQTTPlugin
}

func (p *mqttPlugin) getCreatedAt() time.Time {
	return p.createdAt
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

func (p *mqttPlugin) addPlugin(pluginDesc PluginDesc) {
	if p.addPluginListenerTopic != "" {
		// 如果监听，则发送消息
		payload, err := json.Marshal(pluginDesc)
		if err != nil {
			log.Error("marshal pluginDesc to json error: %v", err)
		} else {
			p.caller.Send(p.addPluginListenerTopic, payload)
		}
	}
	return
}

func (p *mqttPlugin) removePlugin(pluginDesc PluginDesc) {
	if p.removePluginListenerTopic != "" {
		// 如果监听，则发送消息
		payload, err := json.Marshal(pluginDesc)
		if err != nil {
			log.Error("marshal pluginDesc to json error: %v", err)
		} else {
			p.caller.Send(p.removePluginListenerTopic, payload)
		}
	}
	return
}
