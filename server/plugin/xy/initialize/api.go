package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	pluginUtils "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"go.uber.org/zap"
)

const (
	telegramProxyPolicyPath = "/xy/proxy/telegram/*"
)

var xyAPIs = []system.SysApi{
	{Path: "/xy/status", Description: "获取闲鱼助手基础服务状态", ApiGroup: "闲鱼助手", Method: "GET"},
	{Path: "/xy/dashboard/overview", Description: "获取闲鱼助手工作台聚合数据", ApiGroup: "闲鱼助手", Method: "GET"},
	{Path: "/xy/account/list", Description: "获取闲鱼账号列表", ApiGroup: "闲鱼账号", Method: "GET"},
	{Path: "/xy/account/create", Description: "创建闲鱼账号", ApiGroup: "闲鱼账号", Method: "POST"},
	{Path: "/xy/account/update", Description: "更新闲鱼账号", ApiGroup: "闲鱼账号", Method: "PUT"},
	{Path: "/xy/account/delete", Description: "删除闲鱼账号", ApiGroup: "闲鱼账号", Method: "DELETE"},
	{Path: "/xy/account/credential", Description: "加密保存闲鱼账号凭证", ApiGroup: "闲鱼账号", Method: "PUT"},
	{Path: "/xy/account/check-login", Description: "检查闲鱼账号登录态", ApiGroup: "闲鱼账号", Method: "POST"},
	{Path: "/xy/qr-login/start", Description: "创建闲鱼二维码登录会话", ApiGroup: "闲鱼登录", Method: "POST"},
	{Path: "/xy/qr-login/status", Description: "获取闲鱼二维码登录状态", ApiGroup: "闲鱼登录", Method: "GET"},
	{Path: "/xy/qr-login/cancel", Description: "取消闲鱼二维码登录会话", ApiGroup: "闲鱼登录", Method: "DELETE"},
	{Path: "/xy/websocket/start", Description: "启动闲鱼消息连接", ApiGroup: "闲鱼连接", Method: "POST"},
	{Path: "/xy/websocket/stop", Description: "停止闲鱼消息连接", ApiGroup: "闲鱼连接", Method: "POST"},
	{Path: "/xy/websocket/status", Description: "获取闲鱼消息连接状态", ApiGroup: "闲鱼连接", Method: "GET"},
	{Path: "/xy/message/list", Description: "获取闲鱼消息列表", ApiGroup: "闲鱼消息", Method: "GET"},
	{Path: "/xy/message/reply", Description: "发送闲鱼人工回复", ApiGroup: "闲鱼消息", Method: "POST"},
	{Path: "/xy/message/takeover", Description: "更新会话人工接管状态", ApiGroup: "闲鱼消息", Method: "PUT"},
	{Path: "/xy/order/list", Description: "获取闲鱼订单列表", ApiGroup: "闲鱼订单", Method: "GET"},
	{Path: "/xy/order/sync-pending", Description: "同步闲鱼待发货订单", ApiGroup: "闲鱼订单", Method: "POST"},
	{Path: "/xy/event/list", Description: "获取闲鱼业务日志", ApiGroup: "闲鱼日志", Method: "GET"},
	{Path: "/xy/product/list", Description: "获取闲鱼商品列表", ApiGroup: "闲鱼商品", Method: "GET"},
	{Path: "/xy/product/sync", Description: "同步闲鱼商品", ApiGroup: "闲鱼商品", Method: "POST"},
	{Path: "/xy/product/setting", Description: "更新闲鱼商品设置", ApiGroup: "闲鱼商品", Method: "PUT"},
	{Path: "/xy/reply-rule/list", Description: "获取回复规则列表", ApiGroup: "闲鱼回复", Method: "GET"},
	{Path: "/xy/reply-rule/save", Description: "保存回复规则", ApiGroup: "闲鱼回复", Method: "POST"},
	{Path: "/xy/reply-rule/delete", Description: "删除回复规则", ApiGroup: "闲鱼回复", Method: "DELETE"},
	{Path: "/xy/delivery-rule/list", Description: "获取发货规则列表", ApiGroup: "闲鱼发货", Method: "GET"},
	{Path: "/xy/delivery-rule/save", Description: "保存发货规则", ApiGroup: "闲鱼发货", Method: "POST"},
	{Path: "/xy/delivery-rule/delete", Description: "删除发货规则", ApiGroup: "闲鱼发货", Method: "DELETE"},
	{Path: "/xy/delivery-rule/inventory", Description: "获取卡密库存", ApiGroup: "闲鱼发货", Method: "GET"},
	{Path: "/xy/delivery-rule/resource-import", Description: "导入卡密库存", ApiGroup: "闲鱼发货", Method: "POST"},
	{Path: "/xy/delivery-rule/task-list", Description: "获取自动发货任务", ApiGroup: "闲鱼发货", Method: "GET"},
	{Path: "/xy/delivery-rule/task-retry", Description: "重试自动发货任务", ApiGroup: "闲鱼发货", Method: "POST"},
	{Path: telegramProxyPolicyPath, Description: "代理读取 Telegram 服务", ApiGroup: "Telegram 服务代理", Method: "GET"},
	{Path: telegramProxyPolicyPath, Description: "代理写入 Telegram 服务", ApiGroup: "Telegram 服务代理", Method: "POST"},
	{Path: telegramProxyPolicyPath, Description: "代理更新 Telegram 服务", ApiGroup: "Telegram 服务代理", Method: "PUT"},
	{Path: telegramProxyPolicyPath, Description: "代理部分更新 Telegram 服务", ApiGroup: "Telegram 服务代理", Method: "PATCH"},
	{Path: telegramProxyPolicyPath, Description: "代理删除 Telegram 服务资源", ApiGroup: "Telegram 服务代理", Method: "DELETE"},
	{Path: "/xy/ai/models", Description: "获取 AI 模型配置", ApiGroup: "AI 基础配置", Method: "GET"},
	{Path: "/xy/ai/models", Description: "创建 AI 模型配置", ApiGroup: "AI 基础配置", Method: "POST"},
	{Path: "/xy/ai/models/:id", Description: "更新 AI 模型配置", ApiGroup: "AI 基础配置", Method: "PUT"},
	{Path: "/xy/ai/models/:id", Description: "删除 AI 模型配置", ApiGroup: "AI 基础配置", Method: "DELETE"},
	{Path: "/xy/ai/models/:id/default", Description: "设置默认 AI 模型", ApiGroup: "AI 基础配置", Method: "PUT"},
	{Path: "/xy/ai/mcp-servers", Description: "获取 MCP 服务配置", ApiGroup: "AI 基础配置", Method: "GET"},
	{Path: "/xy/ai/mcp-servers", Description: "创建 MCP 服务配置", ApiGroup: "AI 基础配置", Method: "POST"},
	{Path: "/xy/ai/mcp-servers/:id", Description: "更新 MCP 服务配置", ApiGroup: "AI 基础配置", Method: "PUT"},
	{Path: "/xy/ai/mcp-servers/:id", Description: "删除 MCP 服务配置", ApiGroup: "AI 基础配置", Method: "DELETE"},
	{Path: "/xy/ai/agent-config", Description: "获取 Agent 运行配置", ApiGroup: "AI 基础配置", Method: "GET"},
	{Path: "/xy/ai/agent-config", Description: "更新 Agent 运行配置", ApiGroup: "AI 基础配置", Method: "PUT"},
	{Path: "/xy/resources", Description: "获取当前用户资源列表", ApiGroup: "AI 资源", Method: "GET"},
	{Path: "/xy/resources/:id", Description: "获取当前用户资源详情", ApiGroup: "AI 资源", Method: "GET"},
	{Path: "/xy/resource-tasks/:id", Description: "获取当前用户资源任务", ApiGroup: "AI 资源", Method: "GET"},
}

func Api(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).Model(&system.SysApi{}).
		Where("path = ?", "/xy/proxy/telegram/*path").
		Update("path", telegramProxyPolicyPath).Error; err != nil {
		zap.L().Error("迁移 Telegram 代理 API 失败", zap.Error(err))
	}
	pluginUtils.RegisterApis(xyAPIs...)
}
