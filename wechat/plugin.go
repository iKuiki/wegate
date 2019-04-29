package wechat

import (
	"github.com/ikuiki/wwdk/datastruct"
)

/*
Wechat Plugin插件
每个插件代表一个Wechat的功能模块
插件有自己的名称，有自己的描述
插件可以附带监听器来获取wechat的新消息
插件注册时会生成一个token，在主动调用Wechat（发信息等）时需要附带token以验证
*/

// Plugin 插件
type Plugin interface {
	getName() string
	getDescription() string
	// loginStatus 登陆状态变化时
	loginStatus(loginStatus LoginStatusItem)
	// modifyContact 联系人发生修改时的推送
	modifyContact(contact datastruct.Contact)
	// newMessage 接受到新消息时的推送
	newMessage(msg datastruct.Message)
}
