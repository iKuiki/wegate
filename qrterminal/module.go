package qrterminal

import (
	"github.com/ikuiki/wwdk"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/mdp/qrterminal"
	"os"
)

// Module 模块实例化
func Module() module.Module {
	m := new(QrTerminal)
	return m
}

// QrTerminal QrTerminal模块
type QrTerminal struct {
	basemodule.BaseModule
}

// GetType 获取模块类型
func (m *QrTerminal) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "QrTerminal"
}

// Version 获取模块Version
func (m *QrTerminal) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (m *QrTerminal) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)
	m.GetServer().RegisterGO("ShowLoginQrCode", m.showLoginQrCode)
}

// Run 运行主函数
func (m *QrTerminal) Run(closeSig chan bool) {
	token, err := m.RpcInvoke("Wechat", "RegisterRpcPlugin",
		"QrTerminal",      // Name
		"显示扫码的二维码",        // Description
		m.GetType(),       // ModuleType
		"ShowLoginQrCode", // loginListenerFunc
		"",                // contactListenerFunc
		"")                // msgListenerFunc
	if err != "" {
		log.Error("RegisterRpcPlugin error: %s", err)
	}
	log.Info("注册完成，token: %s", token)
	// 关闭信号
	<-closeSig
}

func (m *QrTerminal) showLoginQrCode(loginItem wwdk.LoginChannelItem) (result, err string) {
	log.Debug("收到新的登陆消息(%d)", loginItem.Code)
	if loginItem.Code == wwdk.LoginStatusWaitForScan {
		qrterminal.Generate(loginItem.Msg, qrterminal.L, os.Stdout)
	}
	result = "success"
	return
}
