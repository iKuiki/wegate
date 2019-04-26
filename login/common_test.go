package login_test

import (
	"encoding/json"
	"github.com/liangdas/armyant/work"
	"github.com/pkg/errors"
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
