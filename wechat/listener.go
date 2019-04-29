package wechat

import (
	"github.com/ikuiki/wwdk/datastruct"
)

/*
监听器接口
定义了3种监听器
当内部／外部服务注册时，注册方法负责构造一个实现了这些方法的对象
*/

type statusListener interface {
	// NewStatus 推送新的Status
	NewStatus(statusItem WxServiceStatusItem)
}

type msgListener interface {
	// NewMessage 接受到新消息时的推送
	NewMessage(msg datastruct.Message)
}

type contactListener interface {
	// ModifyContact 联系人发生修改时的推送
	ModifyContact(contact datastruct.Contact)
}

type rpcCaller interface {
	RpcInvokeNR(moduleType string, _func string, params ...interface{}) (err error)
}

type rpcStatusListener struct {
	ModuleType string
	FnName     string
	Caller     rpcCaller
}

func (l *rpcStatusListener) NewStatus(statusItem WxServiceStatusItem) {
	l.Caller.RpcInvokeNR(l.ModuleType, l.FnName, statusItem)
}
