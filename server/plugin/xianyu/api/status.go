package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	xianyuResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type status struct{}

// GetStatus 获取闲鱼助手基础服务状态
// @Tags Xianyu
// @Summary 获取闲鱼助手基础服务状态
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=xianyuResponse.Status,msg=string} "获取成功"
// @Router /xianyu/status [get]
func (a *status) GetStatus(c *gin.Context) {
	var status xianyuResponse.Status
	status, err := statusService.Get(c.Request.Context(), utils.GetUserID(c))
	if err != nil {
		global.GVA_LOG.Error("获取闲鱼助手状态失败", zap.Error(err))
		response.FailWithMessage("获取状态失败", c)
		return
	}
	response.OkWithData(status, c)
}
