package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
)

// BusinessEventAPI 处理业务事件 HTTP 请求。
type BusinessEventAPI struct {
	svc *service.Service // svc 提供业务能力。
}

// NewBusinessEventAPI 创建 BusinessEventAPI。
func NewBusinessEventAPI(svc *service.Service) *BusinessEventAPI {
	return &BusinessEventAPI{svc: svc}
}

// List 处理列表查询请求。
func (a *BusinessEventAPI) List(c *gin.Context) {
	var input xyRequest.BusinessEventSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.svc.BusinessEvent.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取业务日志失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}
