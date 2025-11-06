package response

import "net/http"

// 这个文件 1.用于定义允许返回的HTTP状态码和错误码
//         2.用于定义业务的错误码常量

// 允许使用的HTTP状态码白名单
var allowedHTTPStatus = map[int]struct{}{
	// 2xx 成功
	http.StatusOK: {}, // 200
	// 4xx 客户端错误
	http.StatusBadRequest:       {}, //请求失败
	http.StatusUnauthorized:     {}, //未授权
	http.StatusForbidden:        {}, //禁止访问
	http.StatusNotFound:         {}, //资源未找到
	http.StatusMethodNotAllowed: {}, //方法不允许
	// 5xx 服务器错误
	http.StatusInternalServerError: {}, //服务器内部错误
	http.StatusServiceUnavailable:  {}, //服务不可用
}

//http状态码 5开头表示服务器端错误。4开头表示客户端错误

//Code 代码从 100001 开始,1000 以下为 保留 code
// 10 00 01表示3个部分， 10表示服务，00表示模块，01表示具体错误

// 服务	模块	说明（服务 - 模块）
// 10	0	通用 - 基本错误
// 10	1	通用 - 数据库类错误
// 10	2	通用 - 认证授权类错误
// 10	3	通用 - 加解码类错误
// 11	0	其他服务 - 用户模块错误
// 11	1	其他服务  - 密钥模块错误
// 11	2	其他服务  - 策略模块错误
//通用：说明所有服务都适用的错误，提高复用性，避免重复造轮子。

// 基础错误
// code must start with 1xxxxx
const (
	ErrSuccess int = iota + 100001
	ErrUnknown
	ErrBind
	ErrValidation   //validation failed
	ErrTokenInvalid //token invalid
	ErrNotFound
)

// 数据库类错误
const (
	ErrDatabase int = iota + 100101
	ErrDatabaseNotFind
)

// 认证授权类错误
const (
	ErrEncrypt int = iota + 100201
	ErrSignatureInvalid
	ErrExpired
	ErrorLoginFaild
	ErrInvalidAuthHeader
	ErrMissingHeader //The `Authorization` header was empty.
	ErrPasswordIncorrect
	ErrPermissionDenied //Permission denied.
)

// 编解码类错误
const (
	// ErrEncodingFailed - 500: Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301
	ErrDecodingFailed
	ErrInvalidJSON
	ErrEncodingJSON
	ErrDecodingJSON
	// ErrInvalidYaml - 500: Data is not valid Yaml.
	ErrInvalidYaml
	ErrEncodingYaml
	ErrDecodingYaml
)
