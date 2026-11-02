package system

import (
	"github.com/gcc798/xianyu.bot/internal/model/common/response"
	"github.com/gcc798/xianyu.bot/internal/model/system"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoginLogAPI struct{ *dependencies }

func (s *LoginLogAPI) Delete(c *gin.Context) {
	var loginLog system.LoginLog
	err := c.ShouldBindJSON(&loginLog)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = s.services.LoginLog.Delete(loginLog)
	if err != nil {
		zap.L().Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (s *LoginLogAPI) List(c *gin.Context) {
	var pageInfo systemReq.LoginLogSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := s.services.LoginLog.List(pageInfo)
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
