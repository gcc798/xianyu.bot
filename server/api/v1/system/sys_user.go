package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	systemRes "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (b *BaseApi) Login(c *gin.Context) {
	var input systemReq.Login
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.Username == "" || input.Password == "" {
		response.FailWithMessage("用户名和密码不能为空", c)
		return
	}
	user, err := userService.Login(&system.SysUser{Username: input.Username, Password: input.Password})
	if err != nil {
		recordLogin(c, input.Username, 0, false, "用户名不存在或者密码错误")
		response.FailWithMessage("用户名不存在或者密码错误", c)
		return
	}
	if user.Enable != 1 {
		recordLogin(c, input.Username, user.ID, false, "用户被禁止登录")
		response.FailWithMessage("用户被禁止登录", c)
		return
	}
	b.TokenNext(c, *user)
}

func (b *BaseApi) TokenNext(c *gin.Context, user system.SysUser) {
	token, claims, err := utils.LoginToken(&user)
	if err != nil {
		global.GVA_LOG.Error("获取 token 失败", zap.Error(err))
		response.FailWithMessage("获取token失败", c)
		return
	}
	recordLogin(c, user.Username, user.ID, true, "登录成功")
	utils.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
	response.OkWithDetailed(systemRes.LoginResponse{
		User: user, Token: token, ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, "登录成功", c)
}

func recordLogin(c *gin.Context, username string, userID uint, status bool, message string) {
	if err := loginLogService.CreateLoginLog(system.SysLoginLog{
		Username: username, Ip: c.ClientIP(), Agent: c.Request.UserAgent(), Status: status,
		ErrorMessage: message, UserID: userID,
	}); err != nil {
		global.GVA_LOG.Warn("记录登录日志失败", zap.Error(err))
	}
}

func (b *BaseApi) Register(c *gin.Context) {
	var input systemReq.Register
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.Username == "" || input.Password == "" || input.NickName == "" || input.AuthorityId == 0 {
		response.FailWithMessage("用户名、密码、昵称和角色不能为空", c)
		return
	}
	user, err := userService.Register(system.SysUser{
		Username: input.Username, Password: input.Password, NickName: input.NickName,
		HeaderImg: input.HeaderImg, AuthorityId: input.AuthorityId, Enable: input.Enable,
		Phone: input.Phone, Email: input.Email,
	})
	if err != nil {
		response.FailWithMessage("注册失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(systemRes.SysUserResponse{User: user}, "注册成功", c)
}

func (b *BaseApi) ChangePassword(c *gin.Context) {
	var input systemReq.ChangePasswordReq
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.Password == "" || input.NewPassword == "" {
		response.FailWithMessage("原密码和新密码不能为空", c)
		return
	}
	err := userService.ChangePassword(&system.SysUser{
		GVA_MODEL: global.GVA_MODEL{ID: utils.GetUserID(c)}, Password: input.Password,
	}, input.NewPassword)
	if err != nil {
		response.FailWithMessage("修改失败，原密码与当前账户不符", c)
		return
	}
	response.OkWithMessage("修改成功", c)
}

func (b *BaseApi) GetUserList(c *gin.Context) {
	var page systemReq.GetUserList
	if err := c.ShouldBindJSON(&page); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if page.Page == 0 || page.PageSize == 0 {
		response.FailWithMessage("页码和每页数量不能为空", c)
		return
	}
	list, total, err := userService.GetUserInfoList(page)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: page.Page, PageSize: page.PageSize}, "获取成功", c)
}

func (b *BaseApi) DeleteUser(c *gin.Context) {
	var input request.GetById
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.ID == 0 {
		response.FailWithMessage("用户 ID 不能为空", c)
		return
	}
	if utils.GetUserID(c) == uint(input.ID) {
		response.FailWithMessage("删除失败，无法删除自己", c)
		return
	}
	if err := userService.DeleteUser(input.ID); err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (b *BaseApi) SetUserInfo(c *gin.Context) {
	var input systemReq.ChangeUserInfo
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.ID == 0 {
		response.FailWithMessage("用户 ID 不能为空", c)
		return
	}
	if err := userService.SetUserInfo(system.SysUser{
		GVA_MODEL: global.GVA_MODEL{ID: input.ID}, NickName: input.NickName, HeaderImg: input.HeaderImg,
		Phone: input.Phone, Email: input.Email, Enable: input.Enable, AuthorityId: input.AuthorityId,
	}); err != nil {
		response.FailWithMessage("设置失败", c)
		return
	}
	response.OkWithMessage("设置成功", c)
}

func (b *BaseApi) SetSelfInfo(c *gin.Context) {
	var input systemReq.ChangeUserInfo
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := userService.SetSelfInfo(system.SysUser{
		GVA_MODEL: global.GVA_MODEL{ID: utils.GetUserID(c)}, NickName: input.NickName,
		HeaderImg: input.HeaderImg, Phone: input.Phone, Email: input.Email,
	}); err != nil {
		response.FailWithMessage("设置失败", c)
		return
	}
	response.OkWithMessage("设置成功", c)
}

func (b *BaseApi) GetUserInfo(c *gin.Context) {
	user, err := userService.GetUserInfo(utils.GetUserUuid(c))
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"userInfo": user}, "获取成功", c)
}

func (b *BaseApi) ResetPassword(c *gin.Context) {
	var input systemReq.ResetPassword
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.ID == 0 {
		response.FailWithMessage("用户 ID 不能为空", c)
		return
	}
	if len(input.Password) < 6 {
		response.FailWithMessage("密码至少 6 位", c)
		return
	}
	if err := userService.ResetPassword(input.ID, input.Password); err != nil {
		response.FailWithMessage("重置失败", c)
		return
	}
	response.OkWithMessage("重置成功", c)
}
