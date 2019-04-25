package wechat

import (
	"fmt"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/mdp/qrterminal"
	"os"
)

func (m *Wechat) wechatServe() {
	// 实例化WechatWeb对象
	wx, err := wwdk.NewWechatWeb()
	if err != nil {
		panic("Get new wechatweb client error: " + err.Error())
	}
	// 创建登陆用channel用于回传登陆信息
	loginChan := make(chan wwdk.LoginChannelItem)
	wx.Login(loginChan)
	// 根据channel返回信息进行处理
	for item := range loginChan {
		switch item.Code {
		case wwdk.LoginStatusWaitForScan:
			// 返回了登陆二维码链接，输出到屏幕
			qrterminal.Generate(item.Msg, qrterminal.L, os.Stdout)
		case wwdk.LoginStatusErrorOccurred:
			// 登陆失败
			panic(fmt.Sprintf("WxWeb Login error: %+v", item.Err))
		}
	}
	// 创建同步channel
	syncChannel := make(chan wwdk.SyncChannelItem)
	// 将channel传入startServe方法，开始同步服务并且将新信息通过syncChannel传回
	wx.StartServe(syncChannel)
	// 处理syncChannel传回信息
	for item := range syncChannel {
		// 在子方法内执行逻辑
		switch item.Code {
		// 收到新信息
		case wwdk.SyncStatusNewMessage:
			// 根据收到的信息类型分别处理
			msg := item.Message
			switch msg.MsgType {
			case datastruct.TextMsg:
				// 处理文字信息
				processTextMessage(wx, msg)
			}
		case wwdk.SyncStatusPanic:
			// 发生致命错误，sync中断
			fmt.Printf("sync panic: %+v\n", err)
			break
		}
	}
}

func processTextMessage(app *wwdk.WechatWeb, msg *datastruct.Message) {
	from, err := app.GetContact(msg.FromUserName)
	if err != nil {
		fmt.Println("getContact error: " + err.Error())
		return
	}
	fmt.Printf("Recived a text msg from %s: %s", from.NickName, msg.Content)
}
