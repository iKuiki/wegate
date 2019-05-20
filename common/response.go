package common

// Response 调用返回
type Response struct {
	Ret RetCode // 返回值
	Msg string  // 返回信息
}

// RetCode 返回值
type RetCode int32

const (
	// RetCodeOK 请求成功
	RetCodeOK RetCode = 200
	// RetCodeCreated 请求已经被实现
	RetCodeCreated RetCode = 201
	// RetCodeAccepted 请求已经被接受但尚未实现（或者不会再被实现
	RetCodeAccepted RetCode = 202
	// RetCodeBadRequest 请求错误（格式、语法等错误
	RetCodeBadRequest RetCode = 400
	// RetCodeUnauthorized 未认证
	RetCodeUnauthorized RetCode = 401
	// RetCodeServerError 服务器端出错
	RetCodeServerError RetCode = 500
)
