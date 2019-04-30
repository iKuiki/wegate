package login

/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
)

// Module 模块实例化
func Module() module.Module {
	m := new(Login)
	listener := new(Listener)
	m.SetListener(listener)
	return m
}

// Login 登陆模块
type Login struct {
	basemodule.BaseModule
}

// GetType 获取模块类型
func (m *Login) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Login"
}

// Version 获取模块Version
func (m *Login) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (m *Login) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)
	m.GetServer().RegisterGO("HD_Login", m.login)   //我们约定所有对客户端的请求都以Handler_开头
	m.GetServer().RegisterGO("HD_Logout", m.logout) //我们约定所有对客户端的请求都以Handler_开头
}

// Run 运行主函数
func (m *Login) Run(closeSig chan bool) {
	for {
		select {
		case <-closeSig:
			return
		}
	}
}

// OnDestroy 析构函数
func (m *Login) OnDestroy() {
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}
