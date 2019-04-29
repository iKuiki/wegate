package wechat

// LoginStatus 微信服务状态常量
type LoginStatus int32

const (
	// LoginStatusWaitForScan 等待扫码
	// 返回Msg: 待扫码url
	LoginStatusWaitForScan LoginStatus = 1
	// LoginStatusScanedWaitForLogin 用户已经扫码
	// 返回Msg: 用户头像的base64
	LoginStatusScanedWaitForLogin LoginStatus = 2
	// LoginStatusScanedFinish 用户已同意登陆
	LoginStatusScanedFinish LoginStatus = 3
	// LoginStatusGotCookie 已获取到Cookie
	LoginStatusGotCookie LoginStatus = 4
	// LoginStatusInitFinish 登陆初始化完成
	LoginStatusInitFinish LoginStatus = 5
	// LoginStatusGotContact 已获取到联系人
	LoginStatusGotContact LoginStatus = 6
	// LoginStatusGotBatchContact 已获取到群聊成员
	LoginStatusGotBatchContact LoginStatus = 7
)

// LoginStatusItem 微信登陆状态数据
type LoginStatusItem struct {
	// Status 状态值
	Status LoginStatus
	// Msg 附加信息
	Msg string
}
