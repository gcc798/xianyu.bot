package api

import (
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type product struct{}

// List 分页获取当前用户的商品
// @Tags XyProduct
// @Summary 分页获取当前用户的商品
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.ProductSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/product/list [get]
func (a *product) List(c *gin.Context) {
	var input xyRequest.ProductSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := productService.List(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("获取闲鱼商品失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// Sync 从闲鱼同步商品
// @Tags XyProduct
// @Summary 从闲鱼同步商品
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.ProductSync true "闲鱼账号"
// @Success 200 {object} response.Response{msg=string} "同步成功"
// @Router /xy/product/sync [post]
func (a *product) Sync(c *gin.Context) {
	var input xyRequest.ProductSync
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	count, err := productService.Sync(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("同步闲鱼商品失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"count": count}, "同步成功", c)
}

// UpdateSetting 更新商品业务开关
// @Tags XyProduct
// @Summary 更新商品业务开关
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.UpdateProductSetting true "商品设置"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /xy/product/setting [put]
func (a *product) UpdateSetting(c *gin.Context) {
	var input xyRequest.UpdateProductSetting
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := productService.UpdateSetting(c.Request.Context(), utils.GetUserID(c), input); err != nil {
		failAccount("更新闲鱼商品设置失败："+err.Error(), err, c)
		return
	}
	response.OkWithMessage("更新成功", c)
}
