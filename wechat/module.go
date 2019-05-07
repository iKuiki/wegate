package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/ikuiki/wwdk/storer"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"strings"
)

// Module 模块实例化
func Module() module.Module {
	m := new(Wechat)
	return m
}

// 控制信号
type controlSig int32

const (
	controlSigUploadContactImg controlSig = 100
)

// 用户信息，包括当前登陆用户与联系人列表
type userInfo struct {
	user        *datastruct.User
	contactList map[string]datastruct.Contact
}

// Wechat 微信模块
type Wechat struct {
	basemodule.BaseModule
	// 微信相关对象
	wechat      *wwdk.WechatWeb       // 微信sdk本体
	userInfo    userInfo              // 用户信息，包括当前登陆用户与联系人列表
	loginStatus wwdk.LoginChannelItem // 当前微信状态
	// 插件Map：
	// 插件模块调用本模块提供的注册方法来注册插件到map中
	// 注册时分配一个uuid作为key，并将此uuid存入mqant的session中（为了断开时反注册
	// 当有对应事件发生时，则遍历插件向其发送事件
	pluginMap      map[string]Plugin // 插件注册map
	controlSigChan chan controlSig   // 控制信号发送通道
}

// GetType 获取模块类型
func (m *Wechat) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Wechat"
}

// Version 获取模块Version
func (m *Wechat) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (m *Wechat) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)
	var err error
	m.controlSigChan = make(chan controlSig)
	mediaStorer := newMediaStorer(m.controlSigChan)
	wxConfigs := []interface{}{mediaStorer}
	// 实例化WechatWeb对象
	if filename, ok := settings.Settings["LoginStorerFile"].(string); ok && filename != "" {
		wxConfigs = append(wxConfigs, storer.MustNewFileStorer(filename))
	}
	m.wechat, err = wwdk.NewWechatWeb(wxConfigs...)
	if err != nil {
		panic("Get new wechatweb client error: " + err.Error())
	}
	m.pluginMap = make(map[string]Plugin)
	m.userInfo.contactList = make(map[string]datastruct.Contact)
	m.App.AddRPCSerialize("WechatSerialize", new(wechatSerialize))
	m.GetServer().RegisterGO("RegisterRpcPlugin", m.registerRPCPlugin)
	// 针对wwdk操作的方法都以Wechat开头
	m.GetServer().RegisterGO("Wechat_SendTextMessage", m.sendTextMessage)
	m.GetServer().RegisterGO("Wechat_RevokeMessage", m.revokeMessage)
	m.GetServer().RegisterGO("Wechat_GetUser", m.getUser)
	m.GetServer().RegisterGO("Wechat_GetContactList", m.getContactList)
	m.GetServer().RegisterGO("Wechat_GetContactByUserName", m.getContactByUserName)
	m.GetServer().RegisterGO("Wechat_GetContactByAlias", m.getContactByAlias)
	m.GetServer().RegisterGO("Wechat_GetContactByNickname", m.getContactByNickname)
	m.GetServer().RegisterGO("Wechat_GetContactByRemarkName", m.getContactByRemarkName)
	m.GetServer().RegisterGO("Wechat_ModifyUserRemarkName", m.modifyUserRemarkName)
	m.GetServer().RegisterGO("Wechat_ModifyChatRoomTopic", m.modifyChatRoomTopic)
	m.GetServer().RegisterGO("Wechat_GetRunInfo", m.getRunInfo)
	// ------------------ 客户端 ------------------
	m.GetServer().RegisterGO("HD_Wechat_RegisterMQTTPlugin", m.registerMQTTPlugin)
	m.GetServer().RegisterGO("HD_Wechat_CallWechat", m.callWechat)
	m.GetServer().RegisterGO("Wechat_DisconnectMQTTPlugin", m.disconnectMQTTPlugin)
	// ------------------ 媒体容器客户端 ------------------
	m.GetServer().RegisterGO("HD_Upload_RegisterMQTTUploader", mediaStorer.registerMQTTUploader)
	m.GetServer().RegisterGO("HD_Upload_MQTTUploadFinish", mediaStorer.mqttUploadFinish)
	m.GetServer().RegisterGO("Upload_DisconnectMQTTUploader", mediaStorer.disconnectMQTTUploader)
}

// Run 运行主函数
func (m *Wechat) Run(closeSig chan bool) {
	close, controlClose := make(chan bool), make(chan bool)
	// 执行控制信号
	go func() {
		for {
			select {
			case sig := <-m.controlSigChan:
				switch sig {
				case controlSigUploadContactImg:
					// 检查用户是否上传了头像
					if strings.HasPrefix(m.userInfo.user.HeadImgURL, "/cgi-bin/mmwebwx-bin/") {
						m.syncUser()
					}
					// 统计未完成上传头像的联系人
					var originImgContact []datastruct.Contact
					for _, contact := range m.userInfo.contactList {
						if strings.HasPrefix(contact.HeadImgURL, "/cgi-bin/mmwebwx-bin/") {
							originImgContact = append(originImgContact, contact)
						}
					}
					m.syncContact(originImgContact)
				}
			case <-controlClose:
				return
			}
		}
	}()
	closed := false
	go func() {
		// 内嵌一层函数以异步
		for !closed {
			func() {
				// 在子方法中运行，发生panic后可以及时恢复
				defer func() {
					if e := recover(); e != nil {
						log.Error("wechat module run panic: %+v", e)
					}
				}()
				// 开始执行服务块
				// TODO: 应当把closeSig传入wechatServe中，在其逻辑内合理停止
				m.wechatServe(close)
			}()
		}
	}()
	// 等待到关闭信号，关闭wechat循环与wechat服务器
	<-closeSig
	closed = true
	close <- true
	controlClose <- true
}
