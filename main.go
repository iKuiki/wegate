package main

import (
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/module/modules"
	"wegate/login"
	"wegate/ping"
	"wegate/qrterminal"
	"wegate/wechat"
	"wegate/wgate"
)

//func ChatRoute( app module.App,Type string,hash string) (*module.ServerSession){
//	//演示多个服务路由 默认使用第一个Server
//	log.Debug("Hash:%s 将要调用 type : %s",hash,Type)
//	servers:=app.GetServersByType(Type)
//	if len(servers)==0{
//		return nil
//	}
//	return servers[0]
//}

func main() {
	app := mqant.CreateApp(true) // 只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
	//app.Route("Chat",ChatRoute)
	app.Run(
		modules.MasterModule(),
		wgate.Module(), //这是默认网关模块,是必须的支持 TCP,websocket,MQTT协议
		login.Module(), //这是用户登录验证模块
		wechat.Module(),
		qrterminal.Module(),
		ping.Module(),
	)
}
