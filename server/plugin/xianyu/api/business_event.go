package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	xianyuRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type businessEvent struct{}

func (a *businessEvent) List(c *gin.Context) {
	var input xianyuRequest.BusinessEventSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := businessEventService.List(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("获取业务日志失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}
