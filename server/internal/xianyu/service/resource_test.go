package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeResourceObjectStore struct {
	size          int64
	bucketCreated bool
}

func (s *fakeResourceObjectStore) PresignedPutObject(_ context.Context, _, objectKey string, _ time.Duration) (*url.URL, error) {
	return url.Parse("http://rustfs:9000/resources/" + objectKey)
}

func (s *fakeResourceObjectStore) PresignedGetObject(_ context.Context, _, objectKey string, _ time.Duration, _ url.Values) (*url.URL, error) {
	return url.Parse("http://rustfs:9000/download/" + objectKey)
}

func (s *fakeResourceObjectStore) StatObject(context.Context, string, string, minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return minio.ObjectInfo{Size: s.size}, nil
}

func (s *fakeResourceObjectStore) BucketExists(context.Context, string) (bool, error) {
	return s.bucketCreated, nil
}

func (s *fakeResourceObjectStore) MakeBucket(context.Context, string, minio.MakeBucketOptions) error {
	s.bucketCreated = true
	return nil
}

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
	catalog.objectStore, catalog.bucket = new(fakeResourceObjectStore), "resources"
	downloadURL, expiresAt, err := catalog.Download(t.Context(), 11, ownerResource.ID)
	require.NoError(t, err)
	require.Contains(t, downloadURL, "owners/11/go.pdf")
	require.True(t, expiresAt.After(time.Now()))
	_, _, err = catalog.Download(t.Context(), 11, otherResource.ID)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestResourceDeliveryGrantAndCompletion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ResourceSearch{}, &model.ResourceCandidate{}, &model.ResourceTask{}, &model.Resource{}, &model.ResourceSource{}, &model.AgentSession{}, &model.AgentRun{}, &model.AgentEvent{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })

	search := model.ResourceSearch{OwnerID: 11, Query: "Go", Status: "completed"}
	require.NoError(t, db.Create(&search).Error)
	candidate := model.ResourceCandidate{
		SearchID: search.ID, MCPServerCode: "telegram", ExternalID: "signed-candidate", Name: "Go.pdf",
		Size: 2048, MIMEType: "application/pdf", Source: `{"name":"Go 资料","username":"go_docs"}`, License: "unknown",
	}
	require.NoError(t, db.Create(&candidate).Error)
	store := &fakeResourceObjectStore{size: candidate.Size}
	catalog := &resourceCatalog{objectStore: store, bucket: "resources"}
	session := model.AgentSession{OwnerID: 11, Title: "资源交付"}
	require.NoError(t, db.Create(&session).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: session.ID, ModelID: 1, ModelName: "test", Status: "completed", ConfigSnapshot: `{}`, StartedAt: time.Now()}
	require.NoError(t, db.Create(&run).Error)
	oldEvents := Service.AgentEvents
	Service.AgentEvents = newAgentEventHub()
	t.Cleanup(func() { Service.AgentEvents = oldEvents })
	_, taskEvents, unsubscribe, err := Service.AgentEvents.Subscribe(t.Context(), run.SessionID, 0)
	require.NoError(t, err)
	defer unsubscribe()

	grant, err := catalog.CreateDeliveryTask(t.Context(), 11, candidate.ID, "request-1")
	require.NoError(t, err)
	require.NoError(t, db.Model(&grant.Task).Update("agent_run_id", run.ID).Error)
	grant.Task.AgentRunID = new(run.ID)
	require.NotEmpty(t, grant.Token)
	_, err = catalog.CreateDeliveryTask(t.Context(), 22, candidate.ID, "request-other")
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	expiredAt := time.Now().Add(-time.Minute)
	expiredCandidate := candidate
	expiredCandidate.ID = 0
	expiredCandidate.ExternalID = "expired-candidate"
	expiredCandidate.ExpiresAt = &expiredAt
	require.NoError(t, db.Create(&expiredCandidate).Error)
	_, err = catalog.CreateDeliveryTask(t.Context(), 11, expiredCandidate.ID, "request-expired")
	require.ErrorContains(t, err, "候选已过期")

	uploadURL, _, err := catalog.PresignDelivery(t.Context(), grant.Task.ID, grant.Token)
	require.NoError(t, err)
	require.Contains(t, uploadURL, "resources/11/")
	require.True(t, store.bucketCreated)
	updated, err := catalog.UpdateDelivery(t.Context(), grant.Task.ID, grant.Token, xyRequest.UpdateResourceDelivery{
		Stage: "downloading", ProgressPercent: 25, ProgressDone: 512, ProgressTotal: 2048, ProgressUnit: "bytes",
	})
	require.NoError(t, err)
	require.Equal(t, "running", updated.State)
	require.Equal(t, "task.updated", (<-taskEvents).Type)

	checksum := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	resource, err := catalog.CompleteDelivery(t.Context(), grant.Task.ID, grant.Token, candidate.Size, checksum)
	require.NoError(t, err)
	require.Equal(t, uint(11), resource.OwnerID)
	require.Equal(t, checksum, resource.SHA256)
	require.Equal(t, "task.completed", (<-taskEvents).Type)

	repeated, err := catalog.CompleteDelivery(t.Context(), grant.Task.ID, grant.Token, candidate.Size, checksum)
	require.NoError(t, err)
	require.Equal(t, resource.ID, repeated.ID)
	_, _, err = catalog.PresignDelivery(t.Context(), grant.Task.ID, grant.Token)
	require.ErrorContains(t, err, "当前不能上传")
	_, _, err = catalog.PresignDelivery(t.Context(), grant.Task.ID, "wrong-token")
	require.ErrorContains(t, err, "授权无效")

	failedGrant, err := catalog.CreateDeliveryTask(t.Context(), 11, candidate.ID, "request-failed")
	require.NoError(t, err)
	failed, err := catalog.FailDelivery(t.Context(), failedGrant.Task.ID, failedGrant.Token, xyRequest.FailResourceDelivery{
		ErrorCode: "delivery_failed", ErrorMessage: "资源获取失败",
	})
	require.NoError(t, err)
	require.Equal(t, "failed", failed.State)
}

func TestResourceCatalogCancelsTaskThroughOwningMCPServer(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ResourceTask{}, &model.MCPServer{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })
	server := model.MCPServer{Name: "资源服务", Code: "resources", BaseURL: "http://resources/mcp", Enabled: true, RequestTimeoutSeconds: 15}
	require.NoError(t, db.Create(&server).Error)
	task := model.ResourceTask{
		OwnerID: 11, CandidateID: 1, MCPServerCode: server.Code, ExternalTaskID: "resource_acquire:42",
		IdempotencyKey: "cancel-task", State: "running",
	}
	require.NoError(t, db.Create(&task).Error)
	client := &recordingMCPClient{result: map[string]any{"status": "success", "data": map[string]any{}}}
	registry := &mcpRegistry{connect: func(context.Context, model.MCPServer) (*mcpConnection, error) {
		return &mcpConnection{Server: server, Client: client, Tools: []mcp.Tool{{Name: "task_cancel"}}}, nil
	}}
	catalog := &resourceCatalog{registry: registry}

	_, err = catalog.CancelTask(t.Context(), 22, task.ID)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	cancelled, err := catalog.CancelTask(t.Context(), 11, task.ID)
	require.NoError(t, err)
	require.Equal(t, "cancelled", cancelled.State)
	require.Equal(t, "task_cancel", client.lastName)
	require.Equal(t, task.ExternalTaskID, client.lastArguments["task_id"])
}
