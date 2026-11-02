package api

import (
	"github.com/gcc798/xianyu.bot/internal/infra"
	"github.com/gcc798/xianyu.bot/internal/model/common/response"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/xianyu/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type status struct{ *dependencies }

// GetStatus 获取闲鱼助手基础服务状态
// @Tags Xy
// @Summary 获取闲鱼助手基础服务状态
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=xyResponse.Status,msg=string} "获取成功"
// @Router /xy/status [get]
func (a *status) GetStatus(c *gin.Context) {
	var status xyResponse.Status
	status, err := a.services.Status.Get(c.Request.Context(), security.GetUserID(c))
	if err != nil {
		infra.Log.Error("获取闲鱼助手状态失败", zap.Error(err))
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
func (a *status) GetOverview(c *gin.Context) {
	var input xyRequest.DashboardOverview
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	overview, err := a.services.Status.Overview(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		infra.Log.Error("获取闲鱼助手工作台失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(overview, c)
}
