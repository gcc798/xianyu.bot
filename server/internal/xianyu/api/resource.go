package api

import (
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type resource struct{}

func (a *resource) List(c *gin.Context) {
	var input xyRequest.ResourceSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	resources, total, err := resourceService.List(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("获取资源列表失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: resources, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

func (a *resource) Get(c *gin.Context) {
	resourceID, ok := pathID(c)
	if !ok {
		return
	}
	resource, err := resourceService.Get(c.Request.Context(), utils.GetUserID(c), resourceID)
	if err != nil {
		failAccount("获取资源详情失败", err, c)
		return
	}
	response.OkWithData(gin.H{"resource": resource}, c)
}

func (a *resource) Download(c *gin.Context) {
	resourceID, ok := pathID(c)
	if !ok {
		return
	}
	downloadURL, expiresAt, err := resourceService.Download(c.Request.Context(), utils.GetUserID(c), resourceID)
	if err != nil {
		failAccount("获取资源下载地址失败", err, c)
		return
	}
	response.OkWithData(gin.H{"downloadUrl": downloadURL, "expiresAt": expiresAt}, c)
}

func (a *resource) GetTask(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	task, err := resourceService.GetTask(c.Request.Context(), utils.GetUserID(c), taskID)
	if err != nil {
		failAccount("获取资源任务失败", err, c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

func (a *resource) CancelTask(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	task, err := resourceService.CancelTask(c.Request.Context(), utils.GetUserID(c), taskID)
	if err != nil {
		failAccount("取消资源任务失败", err, c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

func (a *resource) PresignDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	uploadURL, expiresAt, err := resourceService.PresignDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"uploadUrl": uploadURL, "expiresAt": expiresAt}, c)
}

func (a *resource) CompleteDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.CompleteResourceDelivery
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	resource, err := resourceService.CompleteDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input.Size, input.SHA256)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"resource": resource}, c)
}

func (a *resource) UpdateDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.UpdateResourceDelivery
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := resourceService.UpdateDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

func (a *resource) FailDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.FailResourceDelivery
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := resourceService.FailDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}
