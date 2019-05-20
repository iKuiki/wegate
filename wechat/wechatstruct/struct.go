package wechatstruct

// SendMessageRespond 发送消息的返回
type SendMessageRespond struct {
	LocalID string
	MsgID   string
}

// RevokeMessageRespond 撤回消息的返回
type RevokeMessageRespond struct {
	Introduction string
	SysWording   string
}
