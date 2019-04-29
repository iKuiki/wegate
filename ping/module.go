package ping

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
)

// Module 模块实例化
func Module() module.Module {
	m := new(Ping)
	return m
}

// Ping ping模块
type Ping struct {
	basemodule.BaseModule
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
}

// Run 运行主函数
func (m *Ping) Run(closeSig chan bool) {
}
