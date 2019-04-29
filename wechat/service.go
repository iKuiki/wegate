package wechat

// 提供微信web登陆同步的维护工作
// 提供其他微信子模块注册维护工作
// 向其他子模块发送消息

// RegisterListener 注册监听器
func (s *Wechat) registerListener(listenerType string, moduleType string, fnName string) {
	uuid := listenerType + moduleType + fnName
	s.statusQueue[uuid] = &rpcStatusListener{
		ModuleType: moduleType,
		FnName:     fnName,
		Caller:     s,
	}
}
