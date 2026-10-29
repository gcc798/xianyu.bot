package service

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestResourceCatalogScopesQueriesByOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Resource{}, &model.ResourceSource{}, &model.ResourceTask{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })

	ownerResource := model.Resource{OwnerID: 11, TaskID: 1, Name: "Go.pdf", Size: 100, Bucket: "secret-bucket", ObjectKey: "owners/11/go.pdf"}
	otherResource := model.Resource{OwnerID: 22, TaskID: 2, Name: "Private.pdf", Size: 200, Bucket: "secret-bucket", ObjectKey: "owners/22/private.pdf"}
	require.NoError(t, db.Create(&ownerResource).Error)
	require.NoError(t, db.Create(&otherResource).Error)
	require.NoError(t, db.Create(&model.ResourceSource{ResourceID: ownerResource.ID, MCPServerCode: "telegram", ExternalID: "candidate", SourceName: "Go 资料"}).Error)
	require.NoError(t, db.Create(&model.ResourceTask{OwnerID: 11, CandidateID: 1, MCPServerCode: "telegram", IdempotencyKey: "task-owner", State: "queued"}).Error)
	require.NoError(t, db.Create(&model.ResourceTask{OwnerID: 22, CandidateID: 2, MCPServerCode: "telegram", IdempotencyKey: "task-other", State: "queued"}).Error)

	catalog := new(resourceCatalog)
	resources, total, err := catalog.List(t.Context(), 11, xyRequest.ResourceSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "Go.pdf", resources[0].Name)

	detail, err := catalog.Get(t.Context(), 11, ownerResource.ID)
	require.NoError(t, err)
	require.Len(t, detail.Sources, 1)
	_, err = catalog.Get(t.Context(), 11, otherResource.ID)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = catalog.GetTask(t.Context(), 11, 2)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	payload, err := json.Marshal(detail)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "owners/11/go.pdf")
	require.NotContains(t, string(payload), "secret-bucket")
}
