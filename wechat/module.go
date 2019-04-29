package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
)

// Module 模块实例化
func Module() module.Module {
	m := new(Wechat)
	return m
}

// Wechat 微信模块
type Wechat struct {
	basemodule.BaseModule
	// 微信相关对象
	wechat *wwdk.WechatWeb // 微信sdk本体
	// 监听器队列：
	// 监听器队列调用本模块提供的注册方法来注册监听器到queue中
	// 注册时分配一个uuid作为key，并将此uuid存入mqant的session中（为了断开时反注册
	// 当有对应事件发生时，则遍历监听器向其发送事件
	statusQueue  map[string]statusListener  // 状态监听器队列
	messageQueue map[string]msgListener     // 信息监听器队列
	contactQueue map[string]contactListener // 联系人监听器队列
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
}

// Run 运行主函数
func (m *Wechat) Run(closeSig chan bool) {
}
