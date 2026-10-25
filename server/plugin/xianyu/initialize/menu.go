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
			Path:      "xianyu",
			Name:      "Xianyu",
			Component: "view/routerHolder.vue",
			Sort:      2,
			Meta:      system.Meta{Title: "闲鱼助手", Icon: "shop"},
		},
		system.SysBaseMenu{
			Path:      "dashboard",
			Name:      "XianyuDashboard",
			Component: "plugin/xianyu/view/dashboard.vue",
			Sort:      1,
			Meta:      system.Meta{Title: "工作台", Icon: "odometer"},
		},
		system.SysBaseMenu{
			Path:      "accounts",
			Name:      "XianyuAccounts",
			Component: "plugin/xianyu/view/accounts.vue",
			Sort:      2,
			Meta:      system.Meta{Title: "账号管理", Icon: "user"},
		},
		system.SysBaseMenu{
			Path:      "products",
			Name:      "XianyuProducts",
			Component: "plugin/xianyu/view/products.vue",
			Sort:      3,
			Meta:      system.Meta{Title: "商品管理", Icon: "goods"},
		},
		system.SysBaseMenu{
			Path:      "messages",
			Name:      "XianyuMessages",
			Component: "plugin/xianyu/view/messages.vue",
			Sort:      4,
			Meta:      system.Meta{Title: "消息记录", Icon: "chat-line-square"},
		},
		system.SysBaseMenu{
			Path:      "orders",
			Name:      "XianyuOrders",
			Component: "plugin/xianyu/view/orders.vue",
			Sort:      5,
			Meta:      system.Meta{Title: "订单管理", Icon: "tickets"},
		},
		system.SysBaseMenu{
			Path:      "reply-rules",
			Name:      "XianyuReplyRules",
			Component: "plugin/xianyu/view/reply-rules.vue",
			Sort:      6,
			Meta:      system.Meta{Title: "回复规则", Icon: "chat-dot-round"},
		},
		system.SysBaseMenu{
			Path:      "delivery-rules",
			Name:      "XianyuDeliveryRules",
			Component: "plugin/xianyu/view/delivery-rules.vue",
			Sort:      7,
			Meta:      system.Meta{Title: "发货规则", Icon: "box"},
		},
		system.SysBaseMenu{
			Path:      "delivery-tasks",
			Name:      "XianyuDeliveryTasks",
			Component: "plugin/xianyu/view/delivery-tasks.vue",
			Sort:      8,
			Meta:      system.Meta{Title: "自动发货任务", Icon: "finished"},
		},
		system.SysBaseMenu{Path: "business-events", Name: "XianyuBusinessEvents", Component: "plugin/xianyu/view/business-events.vue", Sort: 9, Meta: system.Meta{Title: "业务日志", Icon: "document"}},
	)
}
