package wgate

import (
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
)

// Module 模块实例化
func Module() module.Module {
	gate := new(WGate)
	return gate
}

// WGate 网关
type WGate struct {
	basegate.Gate //继承
}

// GetType 返回Type
func (wgt *WGate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Gate"
}

// Version 返回Version
func (wgt *WGate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

//与客户端通信的自定义粘包示例，需要mqant v1.6.4版本以上才能运行
//该示例只用于简单的演示，并没有实现具体的粘包协议
//去掉下面方法的注释就能启用这个自定义的粘包处理了，但也会造成demo都无法正常通行，因为demo都是用的mqtt粘包协议
//func (this *Gate)CreateAgent() gate.Agent{
//	agent:= NewAgent(this)
//	return agent
//}

// OnInit 模块初始化
func (wgt *WGate) OnInit(app module.App, settings *conf.ModuleSettings) {
	//注意这里一定要用 gate.Gate 而不是 module.BaseModule
	wgt.Gate.OnInit(wgt, app, settings)

	wgt.Gate.SetSessionLearner(wgt)
	wgt.Gate.SetStorageHandler(wgt) //设置持久化处理器
	wgt.Gate.SetTracingHandler(wgt) //设置分布式跟踪系统处理器
}

// Connect 当连接建立  并且MQTT协议握手成功
func (wgt *WGate) Connect(session gate.Session) {
	log.Info("客户端建立了链接")
}

// DisConnect 当连接关闭	或者客户端主动发送MQTT DisConnect命令 ,这个函数中Session无法再继续后续的设置操作，只能读取部分配置内容了
func (wgt *WGate) DisConnect(session gate.Session) {
	log.Info("客户端断开了链接")
}

// OnRequestTracing 是否需要对本次客户端请求进行跟踪
func (wgt *WGate) OnRequestTracing(session gate.Session, topic string, msg []byte) bool {
	if session.GetUserId() == "" {
		//没有登陆的用户不跟踪
		return false
	}
	//if session.GetUserid()!="liangdas"{
	//	//userId 不等于liangdas 的请求不跟踪
	//	return false
	//}
	return true
}

// Storage 存储用户的Session信息
// Session Bind Userid以后每次设置 settings都会调用一次Storage
func (wgt *WGate) Storage(Userid string, session gate.Session) (err error) {
	log.Info("需要处理对Session的持久化")
	return nil
}

// Delete 强制删除Session信息
func (wgt *WGate) Delete(Userid string) (err error) {
	log.Info("需要删除Session持久化数据")
	return nil
}

// Query 获取用户Session信息
// 用户登录以后会调用Query获取最新信息
func (wgt *WGate) Query(Userid string) ([]byte, error) {
	log.Info("查询Session持久化数据")
	return nil, fmt.Errorf("no redis")
}

// Heartbeat 用户心跳,一般用户在线时60s发送一次
// 可以用来延长Session信息过期时间
func (wgt *WGate) Heartbeat(Userid string) {
	log.Info("用户在线的心跳包")
}
