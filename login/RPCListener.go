package login

import (
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/pkg/errors"
)

// Listener login模块自定义rpc监听器
type Listener struct {
	module.RPCModule
}

// BeforeHandle 请求前处理器
func (l *Listener) BeforeHandle(fn string, callInfo *mqrpc.CallInfo) error {
	return nil
}

// NoFoundFunction 请求未找到的自定义处理器
func (l *Listener) NoFoundFunction(fn string) (*mqrpc.FunctionInfo, error) {
	return nil, errors.Errorf("Remote function(%s) not found", fn)
}

// OnComplete 方法完成时处理器
func (l *Listener) OnComplete(fn string, callInfo *mqrpc.CallInfo, result *rpcpb.ResultInfo, execTime int64) {
	log.Info("请求(%s) 执行时间为:[%d 微妙]!", fn, execTime/1000)
}

// OnError 异常时监听器
func (l *Listener) OnError(fn string, callInfo *mqrpc.CallInfo, err error) {
	log.Error("请求(%s)出现异常 error(%s)!", fn, err.Error())
}

// OnTimeOut rpc超时监听器
func (l *Listener) OnTimeOut(fn string, Expired int64) {
	log.Error("请求(%s)超时了!", fn)
}
