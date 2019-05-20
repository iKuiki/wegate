package ping

import (
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"sync"
)

/*
Ping模块简介
Ping模块用来测试服务是否正常
当收到星标联系人发送的ping信息时，返回pong
*/

// Module 模块实例化
func Module() module.Module {
	m := new(Ping)
	return m
}

// Ping ping模块
type Ping struct {
	basemodule.BaseModule
	token        string   // 用来调用wechat模块的token
	starContacts []string // 星标的联系人的username数组
	lock         sync.Mutex
}

// GetType 获取模块类型
func (m *Ping) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Ping"
}

// Version 获取模块Version
func (m *Ping) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (m *Ping) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)
	m.GetServer().RegisterGO("RegisterContacts", m.registerContacts)
	m.GetServer().RegisterGO("ContactModifyHandle", m.contactModifyHandle)
	m.GetServer().RegisterGO("MsgHandle", m.msgHandle)
}

// Run 运行主函数
func (m *Ping) Run(closeSig chan bool) {
	token, err := m.RpcInvoke("Wechat", "RegisterRpcPlugin",
		"Ping",                // Name
		"接收到ping时返回响应",        // Description
		m.GetType(),           // ModuleType
		"RegisterContacts",    // loginListenerFunc
		"ContactModifyHandle", // contactListenerFunc
		"MsgHandle",           // msgListenerFunc
		"",                    // addPluginListenerFunc
		"",                    // removePluginListenerFunc
	)
	if err != "" {
		log.Error("RegisterRpcPlugin error: %s", err)
	} else {
		m.token = token.(string)
	}
	log.Debug("ping模块注册完成，token: %s", token)
	// 关闭信号
	<-closeSig
}

func (m *Ping) registerContacts(loginItem wwdk.LoginChannelItem) (result, err string) {
	if loginItem.Code == wwdk.LoginStatusGotBatchContact {
		log.Debug("检测到登陆成功(%d)开始获取星标联系人", loginItem.Code)
		resp, err := m.RpcInvoke("Wechat", "Wechat_GetContactList", m.token)
		if err != "" {
			log.Error("Wechat getContactList error: %s", err)
		}
		log.Debug("rpc结束")
		if contacts, ok := resp.([]datastruct.Contact); ok {
			for _, contact := range contacts {
				if contact.IsStar() {
					m.starContacts = append(m.starContacts, contact.UserName)
				}
			}
		}
		m.starContacts = language.ArrayUnique(m.starContacts).([]string)
		log.Debug("共找到%d位星标联系人", len(m.starContacts))
	}
	result = "success"
	return
}

func (m *Ping) contactModifyHandle(contact datastruct.Contact) (result, err string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if contact.IsStar() {
		if language.ArrayIn(m.starContacts, contact.UserName) == -1 {
			// 找到新的星标联系人
			log.Debug("发现新的星标联系人：%s", contact.NickName)
			m.starContacts = append(m.starContacts, contact.UserName)
		}
	} else {
		if language.ArrayIn(m.starContacts, contact.UserName) != -1 {
			// 发现已经移除的星标联系人
			log.Debug("发现已经移除的星标联系人：%s", contact.NickName)
			olds := m.starContacts
			m.starContacts = []string{}
			for _, old := range olds {
				if old != contact.UserName {
					m.starContacts = append(m.starContacts, old)
				}
			}
		}
	}
	return
}

func (m *Ping) msgHandle(msg datastruct.Message) (result, err string) {
	if msg.Content == "ping" {
		if m.token != "" {
			// 有token说明初始化了，才可以处理
			if language.ArrayIn(m.starContacts, msg.FromUserName) != -1 {
				log.Debug("收到ping消息，开始回复")
				// 验证此用户是否在星标列表中
				m.RpcInvoke("Wechat", "Wechat_SendTextMessage",
					m.token,
					msg.FromUserName,
					"pong",
				)
			}
		}
	}
	return
}
