package wechat

import (
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/pkg/errors"
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
	statusListenerFunc  string
	contactListenerFunc string
	msgListenerFunc     string
	caller              rpcCaller
}

// RegisterListener 注册监听器
func (s *Wechat) registerListener(
	name,
	description,
	moduleType,
	statusListenerFunc,
	contactListenerFunc,
	msgListenerFunc string) (token string, err error) {
	// ------------------ func start ------------------
	// 检查rpcPlugin是否符合规范
	if name == "" || description == "" || moduleType == "" {
		err = errors.New("module name or description or moduleType is empty")
		return
	}
	plugin := &rpcPlugin{
		name:                name,
		description:         description,
		moduleType:          moduleType,
		statusListenerFunc:  statusListenerFunc,
		contactListenerFunc: contactListenerFunc,
		msgListenerFunc:     msgListenerFunc,
		caller:              s,
	}
	token = name
	s.pluginMap[token] = plugin
	return
}

func (p *rpcPlugin) getName() string {
	return p.name
}

func (p *rpcPlugin) getDescription() string {
	return p.description
}

func (p *rpcPlugin) newStatus(statusItem WxServiceStatusItem) {
	if p.statusListenerFunc != "" {
		p.caller.RpcInvokeNR(p.moduleType, p.statusListenerFunc, statusItem)
	}
}

func (p *rpcPlugin) modifyContact(contact datastruct.Contact) {
	if p.contactListenerFunc != "" {
		p.caller.RpcInvokeNR(p.moduleType, p.contactListenerFunc, contact)
	}
}

func (p *rpcPlugin) newMessage(msg datastruct.Message) {
	if p.msgListenerFunc != "" {
		p.caller.RpcInvokeNR(p.moduleType, p.msgListenerFunc, msg)
	}
}
