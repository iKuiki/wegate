package wechat_test

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
	"wegate/wechat"
	"wegate/wgate"
)

func TestMQTT(t *testing.T) {
	// 先创建一个服务器
	conf.LoadConfig("../bin/conf/server.json")
	// 不设置自动登陆
	delete(conf.Conf.Module["Wechat"][0].Settings, "LoginStorerFile")
	app := mqant.CreateApp(true, conf.Conf) //只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
	app.AddRPCSerialize(common.RPCParamWegateResponseMarshalType, new(common.ResponseSerializer))
	//app.Route("Chat",ChatRoute)
	go app.Run(
		wgate.Module(), //这是默认网关模块,是必须的支持 TCP,websocket,MQTT协议
		login.Module(), //这是用户登录验证模块
		wechat.Module(),
	)
	time.Sleep(2 * time.Second) // 小睡2秒等待mqant启动完成
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
	loginStatusChannel := make(chan bool)
	w.On("LoginStatus", func(client MQTT.Client, msg MQTT.Message) {
		t.Log("LoginStatus: " + string(msg.Payload()))
		loginStatusChannel <- true
	})
	pass := conf.Conf.Module["Login"][0].Settings["Password"].(string) + time.Now().Format(time.RFC3339)
	resp, _ := w.Request("Login/HD_Login", []byte(`{"username":"abc","password":"`+pass+`"}`))
	if resp.Ret != common.RetCodeOK {
		t.Fatalf("登录失败: %s", resp.Msg)
	}
	token := resp.Msg
	resp, _ = w.Request("Wechat/HD_Wechat_RegisterMQTTPlugin", []byte(`{"token":"`+token+`","name":"testPlugin","description":"测试模块","loginListenerTopic":"LoginStatus"}`))
	if resp.Ret != common.RetCodeOK {
		t.Fatalf("注册plugin失败: %s", resp.Msg)
	}
	select {
	case <-loginStatusChannel:
		t.Log("login status channel recive")
	case <-time.After(5 * time.Second):
		t.Fatal("waiting login status channel timeout")
	}
	// TODO: 测试调用wechat方法
}
