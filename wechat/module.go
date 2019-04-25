package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
)

// Module 模块实例化
func Module() module.Module {
	wechat := new(Wechat)
	return wechat
}

// Wechat 微信模块
type Wechat struct {
	basemodule.BaseModule
	wechat *wwdk.WechatWeb
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
