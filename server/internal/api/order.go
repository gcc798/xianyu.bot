package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
)

// OrderAPI 处理订单 HTTP 请求。
type OrderAPI struct {
	svc *service.Service // svc 提供业务能力。
}

// NewOrderAPI 创建 OrderAPI。
func NewOrderAPI(svc *service.Service) *OrderAPI { return &OrderAPI{svc: svc} }

// List 分页获取当前用户的闲鱼订单
// @Tags XyOrder
// @Summary 分页获取当前用户的闲鱼订单
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.OrderSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/order/list [get]
func (a *OrderAPI) List(c *gin.Context) {
	var input xyRequest.OrderSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.svc.Order.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取闲鱼订单失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// Detail 获取当前用户的订单详情。
// @Tags XyOrder
// @Summary 获取当前用户的订单详情
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "订单 ID"
// @Success 200 {object} response.Response{data=response.Order,msg=string} "获取成功"
// @Router /xy/order/detail [get]
func (a *OrderAPI) Detail(c *gin.Context) {
	var input xyRequest.GetById
	if err := c.ShouldBindQuery(&input); err != nil || input.ID <= 0 {
		response.FailWithMessage("订单 ID 无效", c)
		return
	}
	item, err := a.svc.Order.Get(c.Request.Context(), security.GetUserID(c), uint(input.ID))
	if err != nil {
		failAccount("获取闲鱼订单详情失败", err, c)
		return
	}
	response.OkWithData(item, c)
}

// SyncPending 从闲鱼同步当前待发货订单
func (a *OrderAPI) SyncPending(c *gin.Context) {
	var input xyRequest.OrderSync
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	count, err := a.svc.Order.SyncPending(c.Request.Context(), security.GetUserID(c), input.AccountID)
	if err != nil {
		failAccount("同步待发货订单失败", err, c)
		return
	}
	response.OkWithData(gin.H{"count": count}, c)
}
