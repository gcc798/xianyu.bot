package api

import (
	"github.com/gcc798/xianyu.bot/internal/agent"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
)

// ResourceAPI 处理资源 HTTP 请求。
type ResourceAPI struct {
	resources *agent.ResourceCatalog // resources 提供资源目录能力。
}

// NewResourceAPI 创建 ResourceAPI。
func NewResourceAPI(resources *agent.ResourceCatalog) *ResourceAPI {
	return &ResourceAPI{resources: resources}
}

// List 处理列表查询请求。
func (a *ResourceAPI) List(c *gin.Context) {
	var input xyRequest.ResourceSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	resources, total, err := a.resources.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取资源列表失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: resources, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// Get 处理详情查询请求。
func (a *ResourceAPI) Get(c *gin.Context) {
	resourceID, ok := pathID(c)
	if !ok {
		return
	}
	resource, err := a.resources.Get(c.Request.Context(), security.GetUserID(c), resourceID)
	if err != nil {
		failAccount("获取资源详情失败", err, c)
		return
	}
	response.OkWithData(gin.H{"resource": resource}, c)
}

// Download 处理资源下载请求。
func (a *ResourceAPI) Download(c *gin.Context) {
	resourceID, ok := pathID(c)
	if !ok {
		return
	}
	downloadURL, expiresAt, err := a.resources.Download(c.Request.Context(), security.GetUserID(c), resourceID)
	if err != nil {
		failAccount("获取资源下载地址失败", err, c)
		return
	}
	response.OkWithData(gin.H{"downloadUrl": downloadURL, "expiresAt": expiresAt}, c)
}

// GetTask 处理资源任务查询请求。
func (a *ResourceAPI) GetTask(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	task, err := a.resources.GetTask(c.Request.Context(), security.GetUserID(c), taskID)
	if err != nil {
		failAccount("获取资源任务失败", err, c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

// CancelTask 处理资源任务取消请求。
func (a *ResourceAPI) CancelTask(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	task, err := a.resources.CancelTask(c.Request.Context(), security.GetUserID(c), taskID)
	if err != nil {
		failAccount("取消资源任务失败", err, c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

// PresignDelivery 处理资源交付预签名请求。
func (a *ResourceAPI) PresignDelivery(c *gin.Context) {
	taskID, ok := pathID(c)
	if !ok {
		return
	}
	uploadURL, expiresAt, err := a.resources.PresignDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"uploadUrl": uploadURL, "expiresAt": expiresAt}, c)
}

// CompleteDelivery 处理资源交付完成请求。
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
	resource, err := a.resources.CompleteDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input.Size, input.SHA256)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"resource": resource}, c)
}

// UpdateDelivery 处理资源交付进度更新请求。
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
	task, err := a.resources.UpdateDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}

// FailDelivery 处理资源交付失败请求。
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
	task, err := a.resources.FailDelivery(c.Request.Context(), taskID, c.GetHeader("X-Delivery-Token"), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"task": task}, c)
}
