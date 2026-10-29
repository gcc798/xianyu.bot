package response

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"

type AgentSessionDetail struct {
	model.AgentSession
	Messages []model.AgentMessage `json:"messages"`
}
