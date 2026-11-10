package agent

import (
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTelegramResourceTaskOwnershipAndCancellation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.TelegramDialog{}, &model.TelegramResourceTask{}))
	require.NoError(t, db.Exec("CREATE TABLE tel_settings (key text PRIMARY KEY, value text)").Error)
	require.NoError(t, db.Create(&model.TelegramDialog{
		TelegramID: 100, Name: "公开群", DialogType: "群组", IsGroup: true, Status: "normal",
		ResourceComplianceStatus: "allowed",
	}).Error)
	resources := NewResourceCatalog(db, nil, nil)

	task, err := resources.CreateTelegramTask(t.Context(), 7, xyRequest.CreateTelegramResourceTask{
		Keyword: "资料", FileTypes: []string{"PDF", "pdf"}, SourceMode: "manual", SourceDialogIDs: []uint{1, 1}, MaxDownloads: 4,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"pdf"}, task.FileTypes)
	require.Equal(t, []uint{1}, task.SourceDialogIDs)

	page, total, err := resources.ListTelegramTasks(t.Context(), 7, xyRequest.ResourceTaskSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, page, 1)
	_, total, err = resources.ListTelegramTasks(t.Context(), 8, xyRequest.ResourceTaskSearch{})
	require.NoError(t, err)
	require.Zero(t, total)

	_, err = resources.CancelTelegramTask(t.Context(), 8, task.ID)
	require.Error(t, err)
	cancelled, err := resources.CancelTelegramTask(t.Context(), 7, task.ID)
	require.NoError(t, err)
	require.Equal(t, "cancelled", cancelled.Status)
}

func TestCreateTelegramResourceTaskRejectsInvalidInput(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.TelegramDialog{}, &model.TelegramResourceTask{}))
	resources := NewResourceCatalog(db, nil, nil)

	_, err = resources.CreateTelegramTask(t.Context(), 7, xyRequest.CreateTelegramResourceTask{
		Keyword: "资料", SourceMode: "unknown", MaxDownloads: 1,
	})
	require.Error(t, err)
	_, err = resources.CreateTelegramTask(t.Context(), 7, xyRequest.CreateTelegramResourceTask{
		Keyword: "资料", FileTypes: []string{"exe"}, SourceMode: "all", MaxDownloads: 1,
	})
	require.Error(t, err)
}

func TestTelegramResourceFilters(t *testing.T) {
	sources := []model.TelegramDialog{
		{ID: 1, Name: "学习资料群"},
		{ID: 2, Name: "色情频道"},
		{ID: 3, Name: "一般频道"},
	}
	visible := filterTelegramSources(sources, false, map[int64]bool{3: true})
	require.Len(t, visible, 1)
	require.Equal(t, uint(1), visible[0].ID)
	require.Len(t, filterTelegramSources(sources, true, nil), 3)

	require.True(t, telegramFileAllowed("book.PDF", 10, []string{"pdf"}))
	require.False(t, telegramFileAllowed("book.exe", 10, []string{"pdf"}))
	require.False(t, telegramFileAllowed("book.pdf", 0, []string{"pdf"}))
}
