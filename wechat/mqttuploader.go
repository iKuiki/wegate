package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils/uuid"
	"wegate/common"
)

func (s *mediaStorer) registerMQTTUploader(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeUnauthorized,
			Msg: "need login",
		}
		return
	}
	name, description := common.ForceString(msg["name"]), common.ForceString(msg["description"])
	// 检查mqttUploader是否符合规范
	if name == "" || description == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "module name or description is empty",
		}
		return
	}
	uploadListenerTopic := common.ForceString(msg["uploadListenerTopic"])
	if uploadListenerTopic == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "module name or description is empty",
		}
		return
	}
	log.Info("新MQTT Uploader注册：%s[%s]", name, description)
	token := uuid.Rand().Hex()
	session.Set("WechatToken", token)
	uploader := &mqttUploader{
		name:                name,
		description:         description,
		uploadListenerTopic: uploadListenerTopic,
		caller:              session,
	}
	s.uploaders[token] = uploader
	result.Ret = common.RetCodeOK
	result.Msg = token
	return
}

// mqttUploader 对外公开的rpcUploader插件的地址
type mqttUploader struct {
	name                string
	description         string
	uploadListenerTopic string
	caller              mqttCaller
}

func (u *mqttUploader) getName() string {
	return u.name
}

func (u *mqttUploader) getDescription() string {
	return u.description
}

func (u *mqttUploader) upload(file wwdk.MediaFile) {
	u.caller.Send(u.uploadListenerTopic, file.BinaryContent)
	return
}
