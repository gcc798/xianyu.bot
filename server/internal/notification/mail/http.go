package mail

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type message struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func RegisterRoutes(private *gin.RouterGroup) {
	routes := private.Group("email").Use(middleware.OperationRecord())
	routes.POST("emailTest", test)
	routes.POST("sendEmail", sendMessage)
}

// test godoc
// @Tags System
// @Summary 发送测试邮件
// @Security ApiKeyAuth
// @Router /email/emailTest [post]
func test(c *gin.Context) {
	if err := SendConfigured("test", "test"); err != nil {
		global.GVA_LOG.Error("发送测试邮件失败", zap.Error(err))
		response.FailWithMessage("发送失败", c)
		return
	}
	response.OkWithMessage("发送成功", c)
}

// sendMessage godoc
// @Tags System
// @Summary 发送邮件
// @Security ApiKeyAuth
// @Router /email/sendEmail [post]
func sendMessage(c *gin.Context) {
	var input message
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := Send(input.To, input.Subject, input.Body); err != nil {
		global.GVA_LOG.Error("发送邮件失败", zap.Error(err))
		response.FailWithMessage("发送失败", c)
		return
	}
	response.OkWithMessage("发送成功", c)
}
