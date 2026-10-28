package initialize

import (
	"context"
	"errors"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
)

const administratorAuthorityID = 888

func Authorization(ctx context.Context) error {
	var menus []system.SysBaseMenu
	if err := global.GVA_DB.WithContext(ctx).
		Where("name IN ?", []string{"Xianyu", "XianyuDashboard", "XianyuAccounts", "XianyuProducts", "XianyuMessages", "XianyuOrders", "XianyuReplyRules", "XianyuDeliveryRules", "XianyuDeliveryTasks", "XianyuBusinessEvents", "XianyuTelegramResources", "XianyuTelegramOverview", "XianyuTelegramDialogs", "XianyuTelegramTasks", "XianyuTelegramFiles", "XianyuTelegramSettings"}).
		Find(&menus).Error; err != nil {
		return err
	}

	authorityID := strconv.Itoa(administratorAuthorityID)
	for _, menu := range menus {
		association := system.SysAuthorityMenu{MenuId: strconv.Itoa(int(menu.ID)), AuthorityId: authorityID}
		if err := global.GVA_DB.WithContext(ctx).Where(association).FirstOrCreate(&association).Error; err != nil {
			return err
		}
	}

	enforcer := utils.GetCasbin()
	if enforcer == nil {
		return errors.New("casbin enforcer is not initialized")
	}
	if _, err := enforcer.RemoveFilteredPolicy(1, legacyTelegramProxyPolicyPath); err != nil {
		return err
	}
	for _, policy := range [][2]string{
		{"/menu/getMenu", "POST"},
		{"/user/getUserInfo", "GET"},
		{"/xianyu/status", "GET"},
		{"/xianyu/account/list", "GET"},
		{"/xianyu/account/create", "POST"},
		{"/xianyu/account/update", "PUT"},
		{"/xianyu/account/delete", "DELETE"},
		{"/xianyu/account/credential", "PUT"},
		{"/xianyu/account/check-login", "POST"},
		{"/xianyu/qr-login/start", "POST"},
		{"/xianyu/qr-login/status", "GET"},
		{"/xianyu/qr-login/cancel", "DELETE"},
		{"/xianyu/websocket/start", "POST"},
		{"/xianyu/websocket/stop", "POST"},
		{"/xianyu/websocket/status", "GET"},
		{"/xianyu/message/list", "GET"},
		{"/xianyu/message/reply", "POST"},
		{"/xianyu/message/takeover", "PUT"},
		{"/xianyu/order/list", "GET"},
		{"/xianyu/order/sync-pending", "POST"},
		{"/xianyu/event/list", "GET"},
		{"/xianyu/product/list", "GET"},
		{"/xianyu/product/sync", "POST"},
		{"/xianyu/product/setting", "PUT"},
		{"/xianyu/reply-rule/list", "GET"},
		{"/xianyu/reply-rule/save", "POST"},
		{"/xianyu/reply-rule/delete", "DELETE"},
		{"/xianyu/delivery-rule/list", "GET"},
		{"/xianyu/delivery-rule/save", "POST"},
		{"/xianyu/delivery-rule/delete", "DELETE"},
		{"/xianyu/delivery-rule/inventory", "GET"},
		{"/xianyu/delivery-rule/resource-import", "POST"},
		{"/xianyu/delivery-rule/task-list", "GET"},
		{"/xianyu/delivery-rule/task-retry", "POST"},
		{telegramProxyPolicyPath, "GET"},
		{telegramProxyPolicyPath, "POST"},
		{telegramProxyPolicyPath, "PUT"},
		{telegramProxyPolicyPath, "PATCH"},
		{telegramProxyPolicyPath, "DELETE"},
	} {
		if _, err := enforcer.AddPolicy(authorityID, policy[0], policy[1]); err != nil {
			return err
		}
	}
	return nil
}
