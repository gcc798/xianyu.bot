package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是兼容期内所有 HTTP 接口的统一响应信封。
type Response struct {
	Code int    `json:"code"` // Code 是零成功、非零失败的业务状态码。
	Data any    `json:"data"` // Data 是接口返回的业务数据。
	Msg  string `json:"msg"`  // Msg 是供用户查看的结果消息。
}

const (
	// ERROR 是通用业务失败码。
	ERROR = 7
	// SUCCESS 是业务成功码。
	SUCCESS = 0
)

// Result 使用 HTTP 200 写入统一业务响应；code 表示业务结果。
func Result(code int, data any, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		code,
		data,
		msg,
	})
}

// OkWithMessage 返回不带业务数据的成功消息。
func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]any{}, message, c)
}

// OkWithData 返回业务数据和默认成功消息。
func OkWithData(data any, c *gin.Context) {
	Result(SUCCESS, data, "成功", c)
}

// OkWithDetailed 返回业务数据和调用方提供的成功消息。
func OkWithDetailed(data any, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

// FailWithMessage 返回通用业务失败码和用户消息。
func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]any{}, message, c)
}

// NoAuth 使用 HTTP 401 返回未认证响应。
func NoAuth(message string, c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		7,
		nil,
		message,
	})
}
