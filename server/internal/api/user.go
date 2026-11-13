package api

import (
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/request"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	systemRes "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Login 处理用户登录请求。
func (b *UserAPI) Login(c *gin.Context) {
	var input systemReq.Login
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.Username == "" || input.Password == "" {
		response.FailWithMessage("用户名和密码不能为空", c)
		return
	}
	user, err := b.svc.User.Login(&model.User{Username: input.Username, Password: input.Password})
	if err != nil {
		b.recordLogin(c, input.Username, 0, false, "用户名不存在或者密码错误")
		response.FailWithMessage("用户名不存在或者密码错误", c)
		return
	}
	if user.Enable != 1 {
		b.recordLogin(c, input.Username, user.ID, false, "用户被禁止登录")
		response.FailWithMessage("用户被禁止登录", c)
		return
	}
	b.TokenNext(c, *user)
}

// TokenNext 签发用户登录令牌。
func (b *UserAPI) TokenNext(c *gin.Context, user model.User) {
	token, claims, err := b.jwt.LoginToken(&user)
	if err != nil {
		zap.L().Error("获取 token 失败", zap.Error(err))
		response.FailWithMessage("获取token失败", c)
		return
	}
	b.recordLogin(c, user.Username, user.ID, true, "登录成功")
	security.SetToken(c, token, int(claims.RegisteredClaims.ExpiresAt.Unix()-time.Now().Unix()))
	response.OkWithDetailed(systemRes.LoginResponse{
		User: user, Token: token, ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, "登录成功", c)
}

// recordLogin 记录登录审计信息。
func (b *UserAPI) recordLogin(c *gin.Context, username string, userID uint, status bool, message string) {
	if err := b.svc.LoginLog.Create(model.LoginLog{
		Username: username, IP: c.ClientIP(), Agent: c.Request.UserAgent(), Status: status,
		ErrorMessage: message, UserID: userID,
	}); err != nil {
		zap.L().Warn("记录登录日志失败", zap.Error(err))
	}
}

// Register 处理用户注册请求。
func (b *UserAPI) Register(c *gin.Context) {
	var input systemReq.Register
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.Username == "" || input.Password == "" || input.NickName == "" || input.AuthorityID == 0 {
		response.FailWithMessage("用户名、密码、昵称和角色不能为空", c)
		return
	}
	user, err := b.svc.User.Register(model.User{
		Username: input.Username, Password: input.Password, NickName: input.NickName,
		HeaderImage: input.HeaderImage, AuthorityID: input.AuthorityID, Enable: input.Enable,
		Phone: input.Phone, Email: input.Email,
	})
	if err != nil {
		response.FailWithMessage("注册失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(systemRes.UserResponse{User: user}, "注册成功", c)
}

// ChangePassword 处理修改密码请求。
func (b *UserAPI) ChangePassword(c *gin.Context) {
	var input systemReq.ChangePasswordReq
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.Password == "" || input.NewPassword == "" {
		response.FailWithMessage("原密码和新密码不能为空", c)
		return
	}
	err := b.svc.User.ChangePassword(&model.User{
		Timestamps: model.Timestamps{ID: security.GetUserID(c)}, Password: input.Password,
	}, input.NewPassword)
	if err != nil {
		response.FailWithMessage("修改失败，原密码与当前账户不符", c)
		return
	}
	response.OkWithMessage("修改成功", c)
}

// GetUserList 处理用户列表查询请求。
func (b *UserAPI) GetUserList(c *gin.Context) {
	var page systemReq.UserSearch
	if err := c.ShouldBindJSON(&page); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if page.Page == 0 || page.PageSize == 0 {
		response.FailWithMessage("页码和每页数量不能为空", c)
		return
	}
	list, total, err := b.svc.User.List(page)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: page.Page, PageSize: page.PageSize}, "获取成功", c)
}

// DeleteUser 处理删除用户请求。
func (b *UserAPI) DeleteUser(c *gin.Context) {
	var input request.GetById
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.ID == 0 {
		response.FailWithMessage("用户 ID 不能为空", c)
		return
	}
	if security.GetUserID(c) == uint(input.ID) {
		response.FailWithMessage("删除失败，无法删除自己", c)
		return
	}
	if err := b.svc.User.DeleteUser(input.ID); err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// SetUserInfo 处理管理员更新用户请求。
func (b *UserAPI) SetUserInfo(c *gin.Context) {
	var input systemReq.ChangeUserInfo
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if input.ID == 0 {
		response.FailWithMessage("用户 ID 不能为空", c)
		return
	}
	if err := b.svc.User.SetUserInfo(model.User{
		Timestamps: model.Timestamps{ID: input.ID}, NickName: input.NickName, HeaderImage: input.HeaderImage,
		Phone: input.Phone, Email: input.Email, Enable: input.Enable, AuthorityID: input.AuthorityID,
	}); err != nil {
		response.FailWithMessage("设置失败", c)
		return
	}
	response.OkWithMessage("设置成功", c)
}

// SetSelfInfo 处理当前用户更新资料请求。
func (b *UserAPI) SetSelfInfo(c *gin.Context) {
	var input systemReq.ChangeUserInfo
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := b.svc.User.SetSelfInfo(model.User{
		Timestamps: model.Timestamps{ID: security.GetUserID(c)}, NickName: input.NickName,
		HeaderImage: input.HeaderImage, Phone: input.Phone, Email: input.Email,
	}); err != nil {
		response.FailWithMessage("设置失败", c)
		return
	}
	response.OkWithMessage("设置成功", c)
}

// GetUserInfo 处理当前用户信息查询请求。
func (b *UserAPI) GetUserInfo(c *gin.Context) {
	user, err := b.svc.User.GetUserInfo(security.GetUserUuid(c))
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"userInfo": user}, "获取成功", c)
}

// ResetPassword 处理管理员重置密码请求。
func (b *UserAPI) ResetPassword(c *gin.Context) {
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
	if err := b.svc.User.ResetPassword(input.ID, input.Password); err != nil {
		response.FailWithMessage("重置失败", c)
		return
	}
	response.OkWithMessage("重置成功", c)
}
