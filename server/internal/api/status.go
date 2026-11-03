package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StatusAPI struct{ service *service.Service }

// GetStatus 获取闲鱼助手基础服务状态
// @Tags Xy
// @Summary 获取闲鱼助手基础服务状态
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=xyResponse.Status,msg=string} "获取成功"
// @Router /xy/status [get]
func (a *StatusAPI) GetStatus(c *gin.Context) {
	var status xyResponse.Status
	status, err := a.service.Status.Get(c.Request.Context(), security.GetUserID(c))
	if err != nil {
		zap.L().Error("获取闲鱼助手状态失败", zap.Error(err))
		response.FailWithMessage("获取状态失败", c)
		return
	}
	response.OkWithData(status, c)
}

// GetOverview 获取闲鱼助手工作台聚合数据
// @Tags Xy
// @Summary 获取闲鱼助手工作台聚合数据
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.DashboardOverview true "账号和日期范围"
// @Success 200 {object} response.Response{data=xyResponse.DashboardOverview,msg=string} "获取成功"
// @Router /xy/dashboard/overview [get]
func (a *StatusAPI) GetOverview(c *gin.Context) {
	var input xyRequest.DashboardOverview
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	overview, err := a.service.Status.Overview(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		zap.L().Error("获取闲鱼助手工作台失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(overview, c)
}
