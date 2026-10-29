package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	pluginUtils "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Menu(ctx context.Context) {
	_ = ctx
	pluginUtils.RegisterMenus(
		system.SysBaseMenu{
			ParentId:  0,
			Path:      "xy",
			Name:      "Xy",
			Component: "view/routerHolder.vue",
			Sort:      2,
			Meta:      system.Meta{Title: "闲鱼助手", Icon: "shop"},
		},
		system.SysBaseMenu{
			Path:      "dashboard",
			Name:      "XyDashboard",
			Component: "plugin/xy/view/dashboard.vue",
			Sort:      1,
			Meta:      system.Meta{Title: "工作台", Icon: "odometer"},
		},
		system.SysBaseMenu{
			Path:      "accounts",
			Name:      "XyAccounts",
			Component: "plugin/xy/view/accounts.vue",
			Sort:      2,
			Meta:      system.Meta{Title: "账号管理", Icon: "user"},
		},
		system.SysBaseMenu{
			Path:      "products",
			Name:      "XyProducts",
			Component: "plugin/xy/view/products.vue",
			Sort:      3,
			Meta:      system.Meta{Title: "商品管理", Icon: "goods"},
		},
		system.SysBaseMenu{
			Path:      "messages",
			Name:      "XyMessages",
			Component: "plugin/xy/view/messages.vue",
			Sort:      4,
			Meta:      system.Meta{Title: "消息记录", Icon: "chat-line-square"},
		},
		system.SysBaseMenu{
			Path:      "orders",
			Name:      "XyOrders",
			Component: "plugin/xy/view/orders.vue",
			Sort:      5,
			Meta:      system.Meta{Title: "订单管理", Icon: "tickets"},
		},
		system.SysBaseMenu{
			Path:      "reply-rules",
			Name:      "XyReplyRules",
			Component: "plugin/xy/view/reply-rules.vue",
			Sort:      6,
			Meta:      system.Meta{Title: "回复规则", Icon: "chat-dot-round"},
		},
		system.SysBaseMenu{
			Path:      "delivery-rules",
			Name:      "XyDeliveryRules",
			Component: "plugin/xy/view/delivery-rules.vue",
			Sort:      7,
			Meta:      system.Meta{Title: "发货规则", Icon: "box"},
		},
		system.SysBaseMenu{
			Path:      "delivery-tasks",
			Name:      "XyDeliveryTasks",
			Component: "plugin/xy/view/delivery-tasks.vue",
			Sort:      8,
			Meta:      system.Meta{Title: "自动发货任务", Icon: "finished"},
		},
		system.SysBaseMenu{Path: "business-events", Name: "XyBusinessEvents", Component: "plugin/xy/view/business-events.vue", Sort: 9, Meta: system.Meta{Title: "业务日志", Icon: "document"}},
		system.SysBaseMenu{
			Path: "telegram-resources", Name: "XyTelegramResources", Component: "view/routerHolder.vue", Sort: 10,
			Meta: system.Meta{Title: "Telegram 资源", Icon: "folder-opened"},
			Children: []system.SysBaseMenu{
				{Path: "overview", Name: "XyTelegramOverview", Component: "plugin/xy/view/telegram-overview.vue", Sort: 1, Meta: system.Meta{Title: "概览与登录", Icon: "data-analysis"}},
				{Path: "dialogs", Name: "XyTelegramDialogs", Component: "plugin/xy/view/telegram-dialogs.vue", Sort: 2, Meta: system.Meta{Title: "群组与频道", Icon: "chat-line-round"}},
				{Path: "tasks", Name: "XyTelegramTasks", Component: "plugin/xy/view/telegram-resource-tasks.vue", Sort: 3, Meta: system.Meta{Title: "资源检索", Icon: "search"}},
				{Path: "files", Name: "XyTelegramFiles", Component: "plugin/xy/view/telegram-resource-files.vue", Sort: 4, Meta: system.Meta{Title: "资源文件", Icon: "document"}},
				{Path: "settings", Name: "XyTelegramSettings", Component: "plugin/xy/view/telegram-settings.vue", Sort: 5, Meta: system.Meta{Title: "服务设置", Icon: "setting"}},
			},
		},
		system.SysBaseMenu{
			ParentId: 0, Path: "ai", Name: "XyAI", Component: "view/routerHolder.vue", Sort: 3,
			Meta: system.Meta{Title: "AI", Icon: "cpu"},
			Children: []system.SysBaseMenu{
				{Path: "agent", Name: "XyAgent", Component: "plugin/xy/view/ai-agent.vue", Sort: 1, Meta: system.Meta{Title: "Agent", Icon: "chat-dot-square"}},
				{Path: "settings", Name: "XyAISettings", Component: "plugin/xy/view/ai-settings.vue", Sort: 2, Meta: system.Meta{Title: "基础配置", Icon: "setting"}},
			},
		},
	)
}
