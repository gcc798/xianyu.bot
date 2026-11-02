package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
)

type businessEvent struct{ *dependencies }

func (a *businessEvent) List(c *gin.Context) {
	var input xyRequest.BusinessEventSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.services.BusinessEvent.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取业务日志失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}
