/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/
package login

import (
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"time"
	"wegate/common"
)

// Module 模块实例化
func Module() module.Module {
	gate := new(Login)
	return gate
}

// Login 登陆模块
type Login struct {
	basemodule.BaseModule
	currentTime string // 测试用，试好删
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
	m.currentTime = time.Now().Format("15:04:05")   // 测试用，试好删
}

// Run 运行主函数
func (m *Login) Run(closeSig chan bool) {
	for {
		select {
		case <-closeSig:
			break
		case now := <-time.Tick(time.Second):
			m.currentTime = now.Format("15:04:05")
		}
	}
}

// OnDestroy 析构函数
func (m *Login) OnDestroy() {
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}

func (m *Login) login(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if !session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "already login",
		}
		return
	}
	if username, ok := msg["username"]; !ok || username == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "username cannot be empty",
		}
		return
	}
	if password, ok := msg["password"]; !ok || password == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "password cannot be empty",
		}
		return
	}
	username := msg["username"].(string)
	err = session.Bind(username)
	if err != "" {
		return
	}
	session.Set("login", "true")
	session.Push() //推送到网关
	return common.Response{Ret: common.RetCodeOK, Msg: fmt.Sprintf("login success %s", username)}, ""
}

func (m *Login) logout(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeUnauthorized,
			Msg: "is guest, need login",
		}
		return
	}
	session.Remove("login")
	err = session.UnBind()
	if err != "" {
		return
	}
	err = session.Push()
	if err != "" {
		return
	}
	return common.Response{Ret: common.RetCodeOK, Msg: "logout success"}, ""
}
