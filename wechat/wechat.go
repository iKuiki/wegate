package wechat

import (
	"fmt"
	"github.com/ikuiki/wwdk"
	"github.com/liangdas/mqant/log"
)

func (m *Wechat) wechatServe(closeSig <-chan bool) {
	// 创建登陆用channel用于回传登陆信息
	loginChan := make(chan wwdk.LoginChannelItem)
	m.wechat.Login(loginChan)
	// 新建一个func根据channel返回信息进行处理
LOGINLOOP:
	for {
		select {
		case item, ok := <-loginChan:
			if ok {
				m.updateLoginStatus(item)
			} else {
				break LOGINLOOP
			}
		case <-closeSig:
			return
		}
	}
	// 创建同步channel
	syncChannel := make(chan wwdk.SyncChannelItem)
	// 将channel传入startServe方法，开始同步服务并且将新信息通过syncChannel传回
	m.wechat.StartServe(syncChannel)
	// 新建一个func处理syncChannel传回信息
	// 新建一个方法，是为了能够方便的return
	// 之所以用select不用for range，是为了处理closeSig
SYNCLOOP:
	for {
		select {
		case item, ok := <-syncChannel:
			if ok {
				// 在子方法内执行逻辑
				switch item.Code {
				// 联系人变更
				case wwdk.SyncStatusModifyContact:
					// 广播contact
					for _, plugin := range m.pluginMap {
						go func(plugin Plugin) {
							defer func() {
								// 调用外部方法，必须做好recover工作
								if e := recover(); e != nil {
									log.Error("send modify contact panic: %+v", e)
								}
							}()
							plugin.modifyContact(*item.Contact)
						}(plugin)
					}
				// 收到新信息
				case wwdk.SyncStatusNewMessage:
					// 广播message
					for _, plugin := range m.pluginMap {
						go func(plugin Plugin) {
							defer func() {
								// 调用外部方法，必须做好recover工作
								if e := recover(); e != nil {
									log.Error("send new message panic: %+v", e)
								}
							}()
							plugin.newMessage(*item.Message)
						}(plugin)
					}
				case wwdk.SyncStatusPanic:
					// 发生致命错误，sync中断
					panic(fmt.Sprintf("sync panic: %+v", item.Err))
				}
			} else {
				break SYNCLOOP
			}
		case <-closeSig:
			return
		}
	}
}

func (m *Wechat) updateLoginStatus(item wwdk.LoginChannelItem) {
	// 做初步处理
	if item.Code == wwdk.LoginStatusErrorOccurred {
		// 登陆失败
		panic(fmt.Sprintf("WxWeb Login error: %+v", item.Err))
	}
	// 更新到Wechat
	m.loginStatus = item
	// 广播loginStatus
	for _, plugin := range m.pluginMap {
		go func(plugin Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send login status panic: %+v", e)
				}
			}()
			plugin.loginStatus(m.loginStatus)
		}(plugin)
	}
}
