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
	for _, policy := range administratorPolicies() {
		if _, err := enforcer.AddPolicy(authorityID, policy[0], policy[1]); err != nil {
			return err
		}
	}
	return nil
}

func administratorPolicies() [][2]string {
	policies := make([][2]string, 0, len(xianyuAPIs)+2)
	policies = append(policies,
		[2]string{"/menu/getMenu", "POST"},
		[2]string{"/user/getUserInfo", "GET"},
	)
	for _, api := range xianyuAPIs {
		policies = append(policies, [2]string{api.Path, api.Method})
	}
	return policies
}
