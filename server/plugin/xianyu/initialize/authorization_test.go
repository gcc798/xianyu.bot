package initialize

import (
	"testing"

	"github.com/casbin/casbin/v3/util"
)

func TestTelegramProxyPolicyMatchesNestedPath(t *testing.T) {
	if !util.KeyMatch2("/xianyu/proxy/telegram/api/admin/status", telegramProxyPolicyPath) {
		t.Fatalf("Casbin policy %q does not match Telegram API path", telegramProxyPolicyPath)
	}
}
