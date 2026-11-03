package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
)

type ResourceAPI struct{ service *service.Service }

func (a *ResourceAPI) List(c *gin.Context) {
	var input xyRequest.ResourceSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	resources, total, err := a.service.Resource.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取资源列表失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: resources, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

func (a *ResourceAPI) Get(c *gin.Context) {
	resourceID, ok := pathID(c)
	if !ok {
		return
	}
	resource, err := a.service.Resource.Get(c.Request.Context(), security.GetUserID(c), resourceID)
	if err != nil {
		failAccount("获取资源详情失败", err, c)
		return
	}
	response.OkWithData(gin.H{"resource": resource}, c)
}

func (a *ResourceAPI) Download(c *gin.Context) {
	resourceID, ok := pathID(c)
	if !ok {
		return
	}
	downloadURL, expiresAt, err := a.service.Resource.Download(c.Request.Context(), security.GetUserID(c), resourceID)
	if err != nil {
		failAccount("获取资源下载地址失败", err, c)
		return
	}
	response.OkWithData(gin.H{"downloadUrl": downloadURL, "expiresAt": expiresAt}, c)
}

func (a *ResourceAPI) GetTask(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	task, err := a.service.Resource.GetTask(c.Request.Context(), security.GetUserID(c), taskID)
	if err != nil {
		failAccount("获取资源任务失败", err, c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

func (a *ResourceAPI) CancelTask(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	task, err := a.service.Resource.CancelTask(c.Request.Context(), security.GetUserID(c), taskID)
	if err != nil {
		failAccount("取消资源任务失败", err, c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

func (a *ResourceAPI) PresignDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	uploadURL, expiresAt, err := a.service.Resource.PresignDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"uploadUrl": uploadURL, "expiresAt": expiresAt}, c)
}

func (a *ResourceAPI) CompleteDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.CompleteResourceDelivery
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	resource, err := a.service.Resource.CompleteDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input.Size, input.SHA256)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"resource": resource}, c)
}

func (a *ResourceAPI) UpdateDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.UpdateResourceDelivery
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := a.service.Resource.UpdateDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

func (a *ResourceAPI) FailDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.FailResourceDelivery
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := a.service.Resource.FailDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}
