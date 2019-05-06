package common

import (
	"encoding/json"
	"github.com/pkg/errors"
	"reflect"
)

// ResponseSerializer Response的序列化、反序列化器
type ResponseSerializer struct {
}

const (
	// RPCParamWegateResponseMarshalType Response序列化后的pType常量
	RPCParamWegateResponseMarshalType string = "common.Response"
)

// Serialize 序列化
func (ser *ResponseSerializer) Serialize(param interface{}) (ptype string, p []byte, err error) {
	switch v2 := param.(type) {
	case Response:
		bytes, err := json.Marshal(v2)
		if err != nil {
			return "", nil, errors.Errorf("marshal args to json fail: %v", err)
		}
		return RPCParamWegateResponseMarshalType, bytes, nil
	default:
		return "", nil, errors.Errorf("args [%s] Types not allowed", reflect.TypeOf(param))
	}
}

// Deserialize 反序列化
func (ser *ResponseSerializer) Deserialize(ptype string, b []byte) (param interface{}, err error) {
	switch ptype {
	case RPCParamWegateResponseMarshalType:
		var r Response
		err := json.Unmarshal(b, &r)
		if err != nil {
			return nil, errors.Errorf("unmarshal args from json fail: %v", err)
		}
		return r, nil
	default:
		return nil, errors.Errorf("args [%s] Types not allowed", ptype)
	}
}

// GetTypes 返回可以处理的类型
func (ser *ResponseSerializer) GetTypes() []string {
	return []string{RPCParamWegateResponseMarshalType}
}
