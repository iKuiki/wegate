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
	if session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeUnauthorized,
			Msg: "need login",
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
		loginListenerTopic   = common.ForceString(msg["loginListenerTopic"])
		contactListenerTopic = common.ForceString(msg["contactListenerTopic"])
		msgListenerTopic     = common.ForceString(msg["msgListenerTopic"])
	)
	token := uuid.Rand().Hex()
	session.Set("WechatToken", token)
	plugin := &mqttPlugin{
		name:                 name,
		description:          description,
		loginListenerTopic:   loginListenerTopic,
		contactListenerTopic: contactListenerTopic,
		msgListenerTopic:     msgListenerTopic,
		caller:               session,
	}
	plugin.loginStatus(m.loginStatus)
	m.pluginMap[token] = plugin
	result.Ret = common.RetCodeOK
	result.Msg = token
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
