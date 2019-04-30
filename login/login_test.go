package login_test

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/conf"
	"testing"
	"time"
	"wegate/common"
	"wegate/common/test"
	"wegate/login"
	"wegate/wgate"
)

func TestLogin(t *testing.T) {
	// 先创建一个服务器
	conf.LoadConfig("../bin/conf/server.json")
	app := mqant.CreateApp(true, conf.Conf) //只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
	app.AddRPCSerialize(common.RPCParamWegateResponseMarshalType, new(common.ResponseSerializer))
	//app.Route("Chat",ChatRoute)
	go app.Run(
		wgate.Module(), //这是默认网关模块,是必须的支持 TCP,websocket,MQTT协议
		login.Module(), //这是用户登录验证模块
	)
	time.Sleep(time.Second) // 小睡1秒等待mqant启动完成
	// -------------- 开始测试逻辑 --------------
	w := commontest.Work{}
	opts := w.GetDefaultOptions("tcp://127.0.0.1:3563")
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		fmt.Println("ConnectionLost", err.Error())
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		fmt.Println("OnConnectHandler")
	})
	err := w.Connect(opts)
	if err != nil {
		panic(err)
	}
	pass := conf.Conf.Module["Login"][0].Settings["Password"].(string) + time.Now().Format(time.RFC3339)
	objects := []commontest.TestObjective{
		commontest.TestObjective{
			FuncPath:    "Login/HD_Logout",
			Payload:     `{}`,
			ExpectedRet: common.RetCodeUnauthorized,
			Description: "未登陆情况下注销",
		},
		commontest.TestObjective{
			FuncPath:    "Login/HD_Login",
			Payload:     `{}`,
			ExpectedRet: common.RetCodeBadRequest,
			Description: "用户、密码为nil情况下登陆",
		},
		commontest.TestObjective{
			FuncPath:    "Login/HD_Login",
			Payload:     `{"username":"abc","password":"` + pass + `"}`,
			ExpectedRet: common.RetCodeOK,
			Description: "正常登陆",
		},
		commontest.TestObjective{
			FuncPath:    "Login/HD_Login",
			Payload:     `{"username":"abc","password":"` + pass + `"}`,
			ExpectedRet: common.RetCodeBadRequest,
			Description: "登陆后再登陆",
		},
		commontest.TestObjective{
			FuncPath:    "Login/HD_Logout",
			Payload:     `{}`,
			ExpectedRet: common.RetCodeOK,
			Description: "正常注销",
		},
		commontest.TestObjective{
			FuncPath:    "Login/HD_Logout",
			Payload:     `{}`,
			ExpectedRet: common.RetCodeUnauthorized,
			Description: "已注销的情况下注销",
		},
	}
	w.CoverageTesting(t, objects)
}
