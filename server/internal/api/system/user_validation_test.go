package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	systemService "github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserAPIRejectsMissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := New(systemService.New(t.Context(), nil, nil), nil).User
	tests := []struct {
		name    string
		body    string
		handle  func(*gin.Context)
		message string
	}{
		{name: "login", body: `{"username":"admin"}`, handle: api.Login, message: "用户名和密码不能为空"},
		{name: "register", body: `{"userName":"user","passWord":"secret","nickName":"User"}`, handle: api.Register, message: "用户名、密码、昵称和角色不能为空"},
		{name: "change password", body: `{"password":"old"}`, handle: api.ChangePassword, message: "原密码和新密码不能为空"},
		{name: "user list", body: `{}`, handle: api.GetUserList, message: "页码和每页数量不能为空"},
		{name: "delete user", body: `{}`, handle: api.DeleteUser, message: "用户 ID 不能为空"},
		{name: "update user", body: `{}`, handle: api.SetUserInfo, message: "用户 ID 不能为空"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			ctx.Request.Header.Set("Content-Type", "application/json")

			test.handle(ctx)

			var response struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			require.Equal(t, 7, response.Code)
			require.Equal(t, test.message, response.Msg)
		})
	}
}
