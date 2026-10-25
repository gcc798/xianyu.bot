package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	xianyuRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type order struct{}

// List 分页获取当前用户的闲鱼订单
// @Tags XianyuOrder
// @Summary 分页获取当前用户的闲鱼订单
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xianyuRequest.OrderSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xianyu/order/list [get]
func (a *order) List(c *gin.Context) {
	var input xianyuRequest.OrderSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := orderService.List(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("获取闲鱼订单失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// SyncPending 从闲鱼同步当前待发货订单
func (a *order) SyncPending(c *gin.Context) {
	var input xianyuRequest.OrderSync
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	count, err := orderService.SyncPending(c.Request.Context(), utils.GetUserID(c), input.AccountID)
	if err != nil {
		failAccount("同步待发货订单失败", err, c)
		return
	}
	response.OkWithData(gin.H{"count": count}, c)
}
