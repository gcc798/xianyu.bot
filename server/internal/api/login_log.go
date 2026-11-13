package api

import (
	"github.com/gcc798/xianyu.bot/internal/model"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoginLogAPI 处理登录审计 HTTP 请求。
type LoginLogAPI struct {
	svc *service.Service // svc 提供业务能力。
}

// NewLoginLogAPI 创建 LoginLogAPI。
func NewLoginLogAPI(svc *service.Service) *LoginLogAPI {
	return &LoginLogAPI{svc: svc}
}

// Delete 处理删除请求。
func (s *LoginLogAPI) Delete(c *gin.Context) {
	var loginLog model.LoginLog
	err := c.ShouldBindJSON(&loginLog)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = s.svc.LoginLog.Delete(loginLog)
	if err != nil {
		zap.L().Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// List 处理列表查询请求。
func (s *LoginLogAPI) List(c *gin.Context) {
	var pageInfo systemReq.LoginLogSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := s.svc.LoginLog.List(pageInfo)
	if err != nil {
		zap.L().Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
