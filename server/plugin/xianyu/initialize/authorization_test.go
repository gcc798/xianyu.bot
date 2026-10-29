package initialize

import (
	"slices"
	"testing"

	"github.com/casbin/casbin/v3/util"
)

func TestTelegramProxyPolicyMatchesNestedPath(t *testing.T) {
	if !util.KeyMatch2("/xianyu/proxy/telegram/api/admin/status", telegramProxyPolicyPath) {
		t.Fatalf("Casbin policy %q does not match Telegram API path", telegramProxyPolicyPath)
	}
}

func TestAdministratorPoliciesIncludeDashboardOverview(t *testing.T) {
	want := [2]string{"/xianyu/dashboard/overview", "GET"}
	if !slices.Contains(administratorPolicies(), want) {
		t.Fatalf("administrator policies do not include %v", want)
	}
}
