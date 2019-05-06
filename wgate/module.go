package wgate

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"wegate/common"
)

// Module 模块实例化
func Module() module.Module {
	gate := new(WGate)
	return gate
}

// WGate 网关
type WGate struct {
	basegate.Gate //继承
}

// GetType 返回Type
func (wgt *WGate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Gate"
}

// Version 返回Version
func (wgt *WGate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

//与客户端通信的自定义粘包示例，需要mqant v1.6.4版本以上才能运行
//该示例只用于简单的演示，并没有实现具体的粘包协议
//去掉下面方法的注释就能启用这个自定义的粘包处理了，但也会造成demo都无法正常通行，因为demo都是用的mqtt粘包协议
//func (this *Gate)CreateAgent() gate.Agent{
//	agent:= NewAgent(this)
//	return agent
//}

// OnInit 模块初始化
func (wgt *WGate) OnInit(app module.App, settings *conf.ModuleSettings) {
	//注意这里一定要用 gate.Gate 而不是 module.BaseModule
	wgt.Gate.OnInit(wgt, app, settings)

	wgt.Gate.SetSessionLearner(wgt)
}

// Connect 当连接建立  并且MQTT协议握手成功
func (wgt *WGate) Connect(session gate.Session) {
	log.Info("客户端建立了链接")
}

// DisConnect 当连接关闭	或者客户端主动发送MQTT DisConnect命令 ,这个函数中Session无法再继续后续的设置操作，只能读取部分配置内容了
func (wgt *WGate) DisConnect(session gate.Session) {
	log.Info("客户端断开了链接")
	// 检查此client是否有注册WechatPlugin
	if token := session.Get("WechatPluginToken"); token != "" {
		log.Debug("检测到客户端注册了WechatPlugin，开始卸载")
		result, eStr := wgt.RpcInvoke("Wechat", "Wechat_DisconnectMQTTPlugin", token)
		if eStr != "" {
			log.Error("call Wechat Wechat_DisconnectMQTTPlugin error: %s", eStr)
		}
		r := result.(common.Response)
		if r.Ret != common.RetCodeOK {
			log.Debug("Wechat_DisconnectMQTTPlugin fail(%d): %s", r.Ret, r.Msg)
		}
	}
	// 检查此client是否有注册WechatUploader
	if token := session.Get("WechatUploaderToken"); token != "" {
		log.Debug("检测到客户端注册了WechatUploader，开始卸载")
		result, eStr := wgt.RpcInvoke("Wechat", "Upload_DisconnectMQTTUploader", token)
		if eStr != "" {
			log.Error("call Wechat Upload_DisconnectMQTTUploader error: %s", eStr)
		}
		r := result.(common.Response)
		if r.Ret != common.RetCodeOK {
			log.Debug("Upload_DisconnectMQTTUploader fail(%d): %s", r.Ret, r.Msg)
		}
	}
}
