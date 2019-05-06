package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/log"
)

// 提供其他微信子模块注册维护工作

type rpcCaller interface {
	RpcInvokeNR(moduleType string, _func string, params ...interface{}) (err error)
}

// rpcPlugin 对外公开的rpcPlugin插件的地址
type rpcPlugin struct {
	name                string
	description         string
	moduleType          string
	loginListenerFunc   string
	contactListenerFunc string
	msgListenerFunc     string
	caller              rpcCaller
}

// registerRPCPlugin 注册监听器
func (m *Wechat) registerRPCPlugin(
	name,
	description,
	moduleType,
	loginListenerFunc,
	contactListenerFunc,
	msgListenerFunc string) (token string, err string) {
	// ------------------ func start ------------------
	// 检查rpcPlugin是否符合规范
	if name == "" || description == "" || moduleType == "" {
		err = "module name or description or moduleType is empty"
		return
	}
	log.Info("新RPC Plugin注册：%s[%s](%s)", name, description, moduleType)
	plugin := &rpcPlugin{
		name:                name,
		description:         description,
		moduleType:          moduleType,
		loginListenerFunc:   loginListenerFunc,
		contactListenerFunc: contactListenerFunc,
		msgListenerFunc:     msgListenerFunc,
		caller:              m,
	}
	token = name
	plugin.loginStatus(m.loginStatus)
	m.pluginMap[token] = plugin
	return
}

func (p *rpcPlugin) getName() string {
	return p.name
}

func (p *rpcPlugin) getDescription() string {
	return p.description
}

func (p *rpcPlugin) loginStatus(loginStatus wwdk.LoginChannelItem) {
	if p.loginListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.loginListenerFunc, loginStatus)
		if e != nil {
			log.Info("推送登陆消息%s(%s)失败: %v", e)
		}
	}
}

func (p *rpcPlugin) modifyContact(contact datastruct.Contact) {
	if p.contactListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.contactListenerFunc, contact)
		if e != nil {
			log.Info("推送修改联系人消息%s(%s)失败: %v", e)
		}
	}
}

func (p *rpcPlugin) newMessage(msg datastruct.Message) {
	if p.msgListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.msgListenerFunc, msg)
		if e != nil {
			log.Info("推送新消息%s(%s)失败: %v", e)
		}
	}
}
