package service

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTelegramAPIConfigEncryptsHashAndSelectsNextAvailable(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString(make([]byte, 32)))
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.TelegramAPIConfig{}, &model.Setting{}))
	configs := NewTelegramAPIConfig(db)

	first, err := configs.Create(t.Context(), xyRequest.SaveTelegramAPIConfig{
		Name: "lower", APIID: 100, APIHash: "0123456789abcdef0123456789abcdef",
	})
	require.NoError(t, err)
	_, err = configs.Create(t.Context(), xyRequest.SaveTelegramAPIConfig{
		Name: "higher", APIID: 200, APIHash: "abcdef0123456789abcdef0123456789",
	})
	require.NoError(t, err)
	var stored model.TelegramAPIConfig
	require.NoError(t, db.First(&stored, first.ID).Error)
	require.NotContains(t, stored.APIHash, "0123456789abcdef")

	active, err := configs.Active(t.Context())
	require.NoError(t, err)
	require.EqualValues(t, 200, active.APIID)
	require.Equal(t, "abcdef0123456789abcdef0123456789", active.APIHash)
	require.NoError(t, configs.RecordFloodWait(t.Context(), active.APIID, 60))

	active, err = configs.Active(t.Context())
	require.NoError(t, err)
	require.EqualValues(t, 100, active.APIID)

	items, total, err := configs.List(t.Context(), xyRequest.TelegramAPIConfigSearch{Pagination: xyRequest.Pagination{Page: 1, PageSize: 10}})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.True(t, items[0].HasAPIHash)
	encoded, err := json.Marshal(items)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "0123456789abcdef")
	require.NotNil(t, items[0].FloodWaitUntil)
	require.WithinDuration(t, time.Now().Add(time.Minute), *items[0].FloodWaitUntil, time.Second*5)
}

func TestTelegramAPIConfigValidatesHash(t *testing.T) {
	err := validateTelegramAPIConfig(xyRequest.SaveTelegramAPIConfig{
		Name: "nagram", APIID: 11535358, APIHash: "not-a-hash",
	}, true)
	require.ErrorContains(t, err, "32 位十六进制")
}
