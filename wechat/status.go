package wechat

// WxServiceStatus 微信服务状态常量
type WxServiceStatus int32

const (
	// WxServiceStatusNeedLogin 需要登陆
	WxServiceStatusNeedLogin WxServiceStatus = 10
)

// WxServiceStatusItem 微信状态数据
type WxServiceStatusItem struct {
	// Status 状态值
	Status WxServiceStatus
	// Msg 附加信息
	Msg string
}
