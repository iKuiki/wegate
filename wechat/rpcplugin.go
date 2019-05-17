package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/log"
	"time"
)

// 提供其他微信子模块注册维护工作

type rpcCaller interface {
	RpcInvokeNR(moduleType string, _func string, params ...interface{}) (err error)
}

// rpcPlugin 对外公开的rpcPlugin插件的地址
type rpcPlugin struct {
	name                     string
	description              string
	moduleType               string
	loginListenerFunc        string
	contactListenerFunc      string
	msgListenerFunc          string
	addPluginListenerFunc    string
	removePluginListenerFunc string
	caller                   rpcCaller
	createdAt                time.Time
}

// registerRPCPlugin 注册监听器
func (m *Wechat) registerRPCPlugin(
	name,
	description,
	moduleType,
	loginListenerFunc,
	contactListenerFunc,
	msgListenerFunc,
	addPluginListenerFunc,
	removePluginListenerFunc string) (token string, err string) {
	// ------------------ func start ------------------
	// 检查rpcPlugin是否符合规范
	if name == "" || description == "" || moduleType == "" {
		err = "module name or description or moduleType is empty"
		return
	}
	log.Info("新RPC Plugin注册：%s[%s](%s)", name, description, moduleType)
	plugin := &rpcPlugin{
		name:                     name,
		description:              description,
		moduleType:               moduleType,
		loginListenerFunc:        loginListenerFunc,
		contactListenerFunc:      contactListenerFunc,
		msgListenerFunc:          msgListenerFunc,
		addPluginListenerFunc:    addPluginListenerFunc,
		removePluginListenerFunc: removePluginListenerFunc,
		caller:                   m,
		createdAt:                time.Now(),
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
			p.addPluginEvent(pDesc)
		}(p)
	}
	token = name
	plugin.loginStatusEvent(m.loginStatus)
	m.pluginMap[token] = plugin
	return
}

func (p *rpcPlugin) getName() string {
	return p.name
}

func (p *rpcPlugin) getDescription() string {
	return p.description
}

func (p *rpcPlugin) getPluginType() PluginType {
	return PluginTypeRPCPlugin
}

func (p *rpcPlugin) getCreatedAt() time.Time {
	return p.createdAt
}

func (p *rpcPlugin) loginStatusEvent(loginStatus wwdk.LoginChannelItem) {
	if p.loginListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.loginListenerFunc, loginStatus)
		if e != nil {
			log.Info("推送登陆消息%s(%s)失败: %v", p.moduleType, p.loginListenerFunc, e)
		}
	}
}

func (p *rpcPlugin) modifyContactEvent(contact datastruct.Contact) {
	if p.contactListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.contactListenerFunc, contact)
		if e != nil {
			log.Info("推送修改联系人消息%s(%s)失败: %v", p.moduleType, p.contactListenerFunc, e)
		}
	}
}

func (p *rpcPlugin) newMessageEvent(msg datastruct.Message) {
	if p.msgListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.msgListenerFunc, msg)
		if e != nil {
			log.Info("推送新消息%s(%s)失败: %v", p.moduleType, p.msgListenerFunc, e)
		}
	}
}

func (p *rpcPlugin) addPluginEvent(pluginDesc PluginDesc) {
	if p.addPluginListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.addPluginListenerFunc, pluginDesc)
		if e != nil {
			log.Info("推送插件注册消息%s(%s)失败: %v", p.moduleType, p.addPluginListenerFunc, e)
		}
	}
}
func (p *rpcPlugin) removePluginEvent(pluginDesc PluginDesc) {
	if p.removePluginListenerFunc != "" {
		e := p.caller.RpcInvokeNR(p.moduleType, p.removePluginListenerFunc, pluginDesc)
		if e != nil {
			log.Info("推送插件卸载消息%s(%s)失败: %v", p.moduleType, p.removePluginListenerFunc, e)
		}
	}
}
