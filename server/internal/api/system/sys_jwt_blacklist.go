package system

import (
	"github.com/gcc798/xianyu.bot/internal/model/common/response"
	"github.com/gcc798/xianyu.bot/internal/model/system"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type JwtApi struct{ *dependencies }

// JsonInBlacklist
// @Tags      Jwt
// @Summary   jwt加入黑名单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{msg=string}  "jwt加入黑名单"
// @Router    /jwt/jsonInBlacklist [post]
func (j *JwtApi) JsonInBlacklist(c *gin.Context) {
	token := security.GetToken(c)
	jwt := system.JwtBlacklist{Jwt: token}
	err := j.services.JWT.JsonInBlacklist(jwt)
	if err != nil {
		zap.L().Error("jwt作废失败!", zap.Error(err))
		response.FailWithMessage("jwt作废失败", c)
		return
	}
	security.ClearToken(c)
	response.OkWithMessage("jwt作废成功", c)
}
