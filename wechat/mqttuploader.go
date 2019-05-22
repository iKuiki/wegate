package wechat

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
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
	// 检查此client是否有注册WechatUploader
	if token := session.Get("WechatUploaderToken"); token != "" {
		log.Debug("检测到session尝试重复注册Uploader")
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "duplicate registered",
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
			Msg: "upload listener topic is empty",
		}
		return
	}
	log.Info("新MQTT Uploader注册：%s[%s]", name, description)
	token := uuid.New().String()
	session.Set("WechatUploaderToken", token)
	eStr := session.Push()
	if eStr != "" {
		log.Error("推送session失败: %s", eStr)
		result = common.Response{
			Ret: common.RetCodeServerError,
			Msg: "push session fail",
		}
		return
	}
	uploader := &mqttUploader{
		name:                name,
		description:         description,
		uploadListenerTopic: uploadListenerTopic,
		caller:              session,
	}
	s.uploaders[token] = uploader
	s.moduleControlSigChan <- controlSigUploadContactImg
	result.Ret = common.RetCodeOK
	result.Msg = token
	return
}

func (s *mediaStorer) disconnectMQTTUploader(token string) (result common.Response, err string) {
	u, ok := s.uploaders[token]
	if !ok {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "uploader not found",
		}
		return
	}
	delete(s.uploaders, token)
	log.Debug("已卸载Uploader[%s]: %s", u.getName(), u.getDescription())
	result = common.Response{
		Ret: common.RetCodeOK,
	}
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

func (u *mqttUploader) upload(file MediaFile) {
	b, _ := json.Marshal(file)
	u.caller.Send(u.uploadListenerTopic, b)
	return
}
