package login_test

import (
	"encoding/json"
	"github.com/liangdas/armyant/work"
	"github.com/pkg/errors"
	"testing"
	"wegate/common"
)

type Work struct {
	work.MqttWork
}

type rpcResult struct {
	Error  string
	Result common.Response
}

func (w Work) Request(topic string, payload []byte) (resp common.Response, err error) {
	msg, err := w.MqttWork.Request(topic, payload)
	if err != nil {
		return
	}
	var result rpcResult
	err = json.Unmarshal(msg.Payload(), &result)
	if err != nil {
		return
	}
	if result.Error != "" {
		err = errors.New("rpc error occurred: " + result.Error)
		return
	}
	resp = result.Result
	return
}

type TestObjective struct {
	FuncPath    string         // 远程rpc路径
	Payload     string         // 请求载体
	ExpectedRet common.RetCode // 期待返回结果
	Description string         // 测试描述，例如「以xxx条件测试」，当测试结果不正确时会输出
}

func (w Work) CoverageTesting(t *testing.T, objects []TestObjective) {
	for _, object := range objects {
		resp, err := w.Request(object.FuncPath, []byte(object.Payload))
		if err != nil {
			t.Fatalf("%s call error: %+v", object.FuncPath, err)
		}
		if resp.Ret != object.ExpectedRet {
			t.Fatalf("%s, expected RetCode %d, but now is %v with Msg: %s",
				object.Description,
				object.ExpectedRet,
				resp.Ret, resp.Msg)
		}
	}
}
