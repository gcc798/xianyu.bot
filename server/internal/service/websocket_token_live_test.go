package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLiveWebSocketTokenCanBeFetched(t *testing.T) {
	db, ownerID, accountID := liveAccount(t)
	tokenService := NewWebSocketToken(db, &BusinessEvent{db: db})
	token, deviceID, err := tokenService.Get(t.Context(), ownerID, accountID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, deviceID)
}

func TestLiveWebSocketCanRegister(t *testing.T) {
	db, _, accountID := liveAccount(t)
	dialer := &XYDialer{db: db, token: func() webSocketTokenGetter {
		token := NewWebSocketToken(db, &BusinessEvent{db: db})
		return &token
	}(), endpoint: webSocketEndpoint}
	ctx, cancel := context.WithTimeoutCause(t.Context(), 20*time.Second, errors.New("在线 WebSocket 连接测试超时"))
	defer cancel()
	connection, err := dialer.Dial(ctx, accountID)
	require.NoError(t, err)
	require.NoError(t, connection.Ping(ctx))
	require.NoError(t, connection.Close())
}

func liveAccount(t *testing.T) (*gorm.DB, uint, uint) {
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
	return db, uint(ownerID), uint(accountID)
}
