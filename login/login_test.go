package login_test

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/conf"
	"testing"
	"time"
	"wegate/common"
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
	w := Work{}
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
	// 未登陆情况下注销
	// 期望：返回RetCodeUnauthorized
	resp, err := w.Request("Login/HD_Logout", []byte(`{}`))
	if err != nil {
		t.Fatalf("Login/HD_Logout error: %+v", err)
	}
	if resp.Ret != common.RetCodeUnauthorized {
		t.Fatalf("not login yet, func(logout) shoud return RetCodeUnauthorized(401), but now is %v with Msg: %s", resp.Ret, resp.Msg)
	}
	// 用户、密码为nil情况下登陆
	// 期望：返回RetCodeBadRequest
	resp, err = w.Request("Login/HD_Login", []byte(`{}`))
	if err != nil {
		t.Fatalf("Login/HD_Login error: %+v", err)
	}
	if resp.Ret != common.RetCodeBadRequest {
		t.Fatalf("login with empty info, shoud return RetCodeBadRequest(400), but now is %v with Msg: %s", resp.Ret, resp.Msg)
	}
	// 正常登陆
	// 期望：返回RetCodeOK
	resp, err = w.Request("Login/HD_Login", []byte(`{"username":"abc","password":"test"}`))
	if err != nil {
		t.Fatalf("Login/HD_Login error: %+v", err)
	}
	if resp.Ret != common.RetCodeOK {
		t.Fatalf("login with empty info, shoud return RetCodeOK(200), but now is %v with Msg: %s", resp.Ret, resp.Msg)
	}
	// 登陆后再登陆
	// 期望：返回RetCodeBadRequest
	resp, err = w.Request("Login/HD_Login", []byte(`{"username":"abc","password":"test"}`))
	if err != nil {
		t.Fatalf("Login/HD_Login error: %+v", err)
	}
	if resp.Ret != common.RetCodeBadRequest {
		t.Fatalf("login with empty info, shoud return RetCodeBadRequest(400), but now is %v with Msg: %s", resp.Ret, resp.Msg)
	}
	// 正常注销
	// 期望：返回RetCodeOK
	resp, err = w.Request("Login/HD_Logout", []byte(`{}`))
	if err != nil {
		t.Fatalf("Login/HD_Logout error: %+v", err)
	}
	if resp.Ret != common.RetCodeOK {
		t.Fatalf("not login yet, func(logout) shoud return RetCodeOK(200), but now is %v with Msg: %s", resp.Ret, resp.Msg)
	}
	// 已注销的情况下注销
	// 期望：返回RetCodeUnauthorized
	resp, err = w.Request("Login/HD_Logout", []byte(`{}`))
	if err != nil {
		t.Fatalf("Login/HD_Logout error: %+v", err)
	}
	if resp.Ret != common.RetCodeUnauthorized {
		t.Fatalf("not login yet, func(logout) shoud return RetCodeUnauthorized(401), but now is %v with Msg: %s", resp.Ret, resp.Msg)
	}
}
