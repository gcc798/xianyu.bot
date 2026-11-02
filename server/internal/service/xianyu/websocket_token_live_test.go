package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/infra"
	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLiveWebSocketTokenCanBeFetched(t *testing.T) {
	ownerID, accountID := liveAccount(t)
	tokenService := newWebSocketToken(infra.DB, &businessEvent{db: infra.DB})
	token, deviceID, err := tokenService.Get(t.Context(), ownerID, accountID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, deviceID)
}

func TestLiveWebSocketCanRegister(t *testing.T) {
	_, accountID := liveAccount(t)
	dialer := &xyDialer{db: infra.DB, token: func() webSocketTokenGetter {
		token := newWebSocketToken(infra.DB, &businessEvent{db: infra.DB})
		return &token
	}(), endpoint: webSocketEndpoint}
	ctx, cancel := context.WithTimeoutCause(t.Context(), 20*time.Second, errors.New("在线 WebSocket 连接测试超时"))
	defer cancel()
	connection, err := dialer.Dial(ctx, accountID)
	require.NoError(t, err)
	require.NoError(t, connection.Ping(ctx))
	require.NoError(t, connection.Close())
}

func liveAccount(t *testing.T) (uint, uint) {
	t.Helper()
	if os.Getenv("XY_LIVE_TEST") != "1" {
		t.Skip("set XY_LIVE_TEST=1 with an authorized local account")
	}
	databasePath := os.Getenv("XY_LIVE_DB")
	accountID, err := strconv.ParseUint(os.Getenv("XY_LIVE_ACCOUNT_ID"), 10, 64)
	require.NoError(t, err)
	ownerID, err := strconv.ParseUint(os.Getenv("XY_LIVE_OWNER_ID"), 10, 64)
	require.NoError(t, err)
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))
	oldDB := infra.DB
	t.Cleanup(func() { infra.DB = oldDB })
	infra.DB = db
	return uint(ownerID), uint(accountID)
}
