package wechat

import (
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/storer"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
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
	wechat      *wwdk.WechatWeb       // 微信sdk本体
	loginStatus wwdk.LoginChannelItem // 当前微信状态
	// 插件Map：
	// 插件模块调用本模块提供的注册方法来注册插件到map中
	// 注册时分配一个uuid作为key，并将此uuid存入mqant的session中（为了断开时反注册
	// 当有对应事件发生时，则遍历插件向其发送事件
	pluginMap map[string]Plugin // 插件注册map
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
	// 实例化WechatWeb对象
	if filename, ok := settings.Settings["LoginStorerFile"].(string); ok && filename != "" {
		loginStorer := storer.MustNewFileStorer(filename)
		m.wechat, err = wwdk.NewWechatWeb(loginStorer)
	} else {
		m.wechat, err = wwdk.NewWechatWeb()
	}
	if err != nil {
		panic("Get new wechatweb client error: " + err.Error())
	}
	m.pluginMap = make(map[string]Plugin)
	m.App.AddRPCSerialize("WechatSerialize", new(wechatSerialize))
	m.GetServer().RegisterGO("RegisterRpcPlugin", m.registerRPCPlugin)
}

// Run 运行主函数
func (m *Wechat) Run(closeSig chan bool) {
	close := make(chan bool)
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
}
