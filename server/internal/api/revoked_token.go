package api

import (
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JWTAPI 处理令牌 HTTP 请求。
type JWTAPI struct {
	service *service.Service // service 提供业务能力。
}

// NewJWTAPI 创建 JWTAPI。
func NewJWTAPI(services *service.Service) *JWTAPI { return &JWTAPI{service: services} }

// Revoke
// @Tags      Jwt
// @Summary   jwt加入黑名单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{msg=string}  "jwt加入黑名单"
// @Router    /jwt/jsonInBlacklist [post]
func (j *JWTAPI) Revoke(c *gin.Context) {
	token := security.GetToken(c)
	err := j.service.RevokedTokens.Revoke(model.RevokedToken{Token: token})
	if err != nil {
		zap.L().Error("jwt作废失败!", zap.Error(err))
		response.FailWithMessage("jwt作废失败", c)
		return
	}
	security.ClearToken(c)
	response.OkWithMessage("jwt作废成功", c)
}
