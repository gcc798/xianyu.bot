package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	xyResponse "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/response"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

const (
	resourceDeliveryTTL = 10 * time.Minute
	resourceDownloadTTL = 10 * time.Minute
)

var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type resourceObjectStore interface {
	PresignedPutObject(ctx context.Context, bucket, objectKey string, expiry time.Duration) (*url.URL, error)
	PresignedGetObject(ctx context.Context, bucket, objectKey string, expiry time.Duration, reqParams url.Values) (*url.URL, error)
	StatObject(ctx context.Context, bucket, objectKey string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	BucketExists(ctx context.Context, bucket string) (bool, error)
	MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error
}

type resourceCatalog struct {
	objectStore resourceObjectStore
	bucket      string
	registry    *mcpRegistry
}

type ResourceDeliveryGrant struct {
	Task      model.ResourceTask
	Token     string
	ExpiresAt time.Time
}

func (s *resourceCatalog) List(ctx context.Context, ownerID uint, input xyRequest.ResourceSearch) ([]model.Resource, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := global.GVA_DB.WithContext(ctx).Model(&model.Resource{}).Where("owner_id = ?", ownerID)
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	if mimeType := strings.TrimSpace(input.MIMEType); mimeType != "" {
		db = db.Where("mime_type = ?", mimeType)
	}
	if state := strings.TrimSpace(input.State); state != "" {
		db = db.Where("state = ?", state)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	resources := make([]model.Resource, 0)
	err := db.Scopes(input.Paginate()).Order("id DESC").Find(&resources).Error
	return resources, total, err
}

func (s *resourceCatalog) Get(ctx context.Context, ownerID, resourceID uint) (xyResponse.ResourceDetail, error) {
	if ownerID == 0 || resourceID == 0 {
		return xyResponse.ResourceDetail{}, errors.New("资源 ID 不能为空")
	}
	var resource model.Resource
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", resourceID, ownerID).First(&resource).Error; err != nil {
		return xyResponse.ResourceDetail{}, err
	}
	sources := make([]model.ResourceSource, 0)
	if err := global.GVA_DB.WithContext(ctx).Where("resource_id = ?", resourceID).Order("id").Find(&sources).Error; err != nil {
		return xyResponse.ResourceDetail{}, err
	}
	return xyResponse.ResourceDetail{Resource: resource, Sources: sources}, nil
}

func (s *resourceCatalog) GetTask(ctx context.Context, ownerID, taskID uint) (model.ResourceTask, error) {
	if ownerID == 0 || taskID == 0 {
		return model.ResourceTask{}, errors.New("任务 ID 不能为空")
	}
	var task model.ResourceTask
	err := global.GVA_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", taskID, ownerID).First(&task).Error
	return task, err
}

func (s *resourceCatalog) CancelTask(ctx context.Context, ownerID, taskID uint) (model.ResourceTask, error) {
	task, err := s.GetTask(ctx, ownerID, taskID)
	if err != nil {
		return model.ResourceTask{}, err
	}
	if task.State == "cancelled" {
		return task, nil
	}
	if task.State != "queued" && task.State != "running" || task.ExternalTaskID == "" {
		return model.ResourceTask{}, errors.New("资源任务当前不能取消")
	}
	registry := s.registry
	if registry == nil {
		registry = new(mcpRegistry)
	}
	connections, err := registry.Discover(ctx)
	if err != nil {
		return model.ResourceTask{}, errors.New("MCP 能力暂不可用")
	}
	defer func() { _ = closeMCPConnections(connections) }()
	connection, definition := approvedTool(connections, task.MCPServerCode, "task_cancel")
	if connection == nil {
		return model.ResourceTask{}, errors.New("MCP 能力暂不可用")
	}
	requestContext, cancel := context.WithTimeout(ctx, min(toolTimeout(definition), time.Duration(connection.Server.RequestTimeoutSeconds)*time.Second))
	defer cancel()
	result, err := connection.Client.CallTool(requestContext, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "task_cancel", Arguments: map[string]any{"task_id": task.ExternalTaskID},
	}})
	if err != nil || result.IsError || !approvedResultSucceeded(mcpToolResultValue(result)) {
		return model.ResourceTask{}, errors.New("MCP 能力暂不可用")
	}
	completedAt := time.Now()
	if err := global.GVA_DB.WithContext(ctx).Model(&task).Where("state IN ?", []string{"queued", "running"}).Updates(map[string]any{
		"state": "cancelled", "stage": "cancelled", "completed_at": completedAt,
	}).Error; err != nil {
		return model.ResourceTask{}, err
	}
	task.State, task.Stage, task.CompletedAt = "cancelled", "cancelled", &completedAt
	publishResourceTaskEvent(ctx, task, "task.updated", nil)
	return task, nil
}

func (s *resourceCatalog) Download(ctx context.Context, ownerID, resourceID uint) (string, time.Time, error) {
	if ownerID == 0 || resourceID == 0 {
		return "", time.Time{}, errors.New("资源 ID 不能为空")
	}
	var resource model.Resource
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND owner_id = ? AND state = ?", resourceID, ownerID, "available").First(&resource).Error; err != nil {
		return "", time.Time{}, err
	}
	store, _, err := s.storage()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(resourceDownloadTTL)
	downloadURL, err := store.PresignedGetObject(ctx, resource.Bucket, resource.ObjectKey, resourceDownloadTTL, url.Values{
		"response-content-disposition": {fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(resource.Name, `"`, ""))},
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("生成资源下载地址: %w", err)
	}
	return downloadURL.String(), expiresAt, nil
}

func (s *resourceCatalog) CreateDeliveryTask(ctx context.Context, ownerID, candidateID uint, idempotencyKey string) (ResourceDeliveryGrant, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if ownerID == 0 || candidateID == 0 || idempotencyKey == "" || len(idempotencyKey) > 100 {
		return ResourceDeliveryGrant{}, errors.New("资源候选和幂等键不能为空")
	}
	token := rand.Text()
	expiresAt := time.Now().Add(resourceDeliveryTTL)
	tokenHash := sha256.Sum256([]byte(token))
	var task model.ResourceTask
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var candidate model.ResourceCandidate
		if err := tx.Table("xy_resource_candidates AS c").
			Joins("JOIN xy_resource_searches AS s ON s.id = c.search_id AND s.deleted_at IS NULL").
			Where("c.id = ? AND c.deleted_at IS NULL AND s.owner_id = ?", candidateID, ownerID).
			Select("c.*").First(&candidate).Error; err != nil {
			return err
		}
		if candidate.ExpiresAt != nil && time.Now().After(*candidate.ExpiresAt) {
			return errors.New("资源候选已过期")
		}
		err := tx.Where("owner_id = ? AND idempotency_key = ?", ownerID, idempotencyKey).First(&task).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			task = model.ResourceTask{
				OwnerID: ownerID, CandidateID: candidateID, MCPServerCode: candidate.MCPServerCode,
				IdempotencyKey: idempotencyKey, State: "queued",
			}
			if err := tx.Create(&task).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		case task.CandidateID != candidateID:
			return errors.New("幂等键已用于其他资源")
		case task.State == "completed" || task.State == "cancelled":
			return errors.New("资源任务已结束")
		}
		objectKey := task.UploadObjectKey
		if objectKey == "" {
			objectKey = fmt.Sprintf("resources/%d/%d/%s", ownerID, task.ID, rand.Text())
		}
		return tx.Model(&task).Updates(map[string]any{
			"delivery_token_hash": hex.EncodeToString(tokenHash[:]),
			"delivery_expires_at": expiresAt,
			"upload_object_key":   objectKey,
		}).Error
	})
	if err != nil {
		return ResourceDeliveryGrant{}, err
	}
	task.DeliveryTokenHash = hex.EncodeToString(tokenHash[:])
	task.DeliveryExpiresAt = &expiresAt
	return ResourceDeliveryGrant{Task: task, Token: token, ExpiresAt: expiresAt}, nil
}

func (s *resourceCatalog) PresignDelivery(ctx context.Context, taskID uint, token string) (string, time.Time, error) {
	task, err := s.deliveryTask(ctx, taskID, token)
	if err != nil {
		return "", time.Time{}, err
	}
	store, bucket, err := s.storage()
	if err != nil {
		return "", time.Time{}, err
	}
	if err := ensureResourceBucket(ctx, store, bucket); err != nil {
		return "", time.Time{}, err
	}
	expiresAt := *task.DeliveryExpiresAt
	if maximum := time.Now().Add(resourceDeliveryTTL); expiresAt.After(maximum) {
		expiresAt = maximum
	}
	if task.State != "queued" && task.State != "running" {
		return "", time.Time{}, errors.New("资源任务当前不能上传")
	}
	uploadURL, err := store.PresignedPutObject(ctx, bucket, task.UploadObjectKey, time.Until(expiresAt))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("生成上传地址: %w", err)
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&task).Where("state = ?", "queued").Update("state", "running").Error; err != nil {
		return "", time.Time{}, err
	}
	return uploadURL.String(), expiresAt, nil
}

func (s *resourceCatalog) CompleteDelivery(ctx context.Context, taskID uint, token string, size int64, checksum string) (model.Resource, error) {
	task, err := s.deliveryTask(ctx, taskID, token)
	if err != nil {
		return model.Resource{}, err
	}
	if task.State != "queued" && task.State != "running" && task.State != "completed" {
		return model.Resource{}, errors.New("资源任务当前不能完成交付")
	}
	checksum = strings.ToLower(strings.TrimSpace(checksum))
	if size <= 0 || !sha256Pattern.MatchString(checksum) {
		return model.Resource{}, errors.New("文件大小或 SHA-256 无效")
	}
	store, bucket, err := s.storage()
	if err != nil {
		return model.Resource{}, err
	}
	object, err := store.StatObject(ctx, bucket, task.UploadObjectKey, minio.StatObjectOptions{})
	if err != nil {
		return model.Resource{}, fmt.Errorf("验证上传对象: %w", err)
	}
	if object.Size != size {
		return model.Resource{}, errors.New("上传对象大小不匹配")
	}

	var resource model.Resource
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ? AND owner_id = ?", task.ID, task.OwnerID).First(&resource).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var candidate model.ResourceCandidate
		if err := tx.First(&candidate, task.CandidateID).Error; err != nil {
			return err
		}
		if candidate.Size > 0 && candidate.Size != size {
			return errors.New("上传对象与候选资源大小不匹配")
		}
		resource = model.Resource{
			OwnerID: task.OwnerID, TaskID: task.ID, Name: candidate.Name, Size: size,
			MIMEType: candidate.MIMEType, SHA256: checksum, Bucket: bucket,
			ObjectKey: task.UploadObjectKey, State: "available",
		}
		if err := tx.Create(&resource).Error; err != nil {
			return err
		}
		var source struct {
			Name     string `json:"name"`
			Username string `json:"username"`
		}
		_ = json.Unmarshal([]byte(candidate.Source), &source)
		resourceSource := model.ResourceSource{
			ResourceID: resource.ID, MCPServerCode: candidate.MCPServerCode, ExternalID: candidate.ExternalID,
			SourceName: source.Name, SourceReference: source.Username, Metadata: candidate.Source,
		}
		if err := tx.Create(&resourceSource).Error; err != nil {
			return err
		}
		now := time.Now()
		result, err := json.Marshal(map[string]any{"resource_id": resource.ID})
		if err != nil {
			return err
		}
		return tx.Model(&task).Updates(map[string]any{
			"state": "completed", "stage": "completed", "progress_percent": 100,
			"progress_done": size, "progress_total": size, "progress_unit": "bytes",
			"result": string(result), "completed_at": now,
		}).Error
	})
	if err == nil {
		task.State, task.Stage, task.ProgressPercent = "completed", "completed", 100
		task.ProgressDone, task.ProgressTotal, task.ProgressUnit = size, size, "bytes"
		publishResourceTaskEvent(ctx, task, "task.completed", map[string]any{"resourceId": resource.ID})
	}
	return resource, err
}

func (s *resourceCatalog) UpdateDelivery(ctx context.Context, taskID uint, token string, input xyRequest.UpdateResourceDelivery) (model.ResourceTask, error) {
	task, err := s.deliveryTask(ctx, taskID, token)
	if err != nil {
		return model.ResourceTask{}, err
	}
	if task.State != "queued" && task.State != "running" {
		return model.ResourceTask{}, errors.New("资源任务当前不能更新")
	}
	updates := map[string]any{
		"state": "running", "stage": input.Stage, "progress_percent": input.ProgressPercent,
		"progress_done": input.ProgressDone, "progress_total": input.ProgressTotal, "progress_unit": input.ProgressUnit,
	}
	if task.StartedAt == nil {
		updates["started_at"] = time.Now()
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
		return model.ResourceTask{}, err
	}
	if err := global.GVA_DB.WithContext(ctx).First(&task, task.ID).Error; err != nil {
		return model.ResourceTask{}, err
	}
	publishResourceTaskEvent(ctx, task, "task.updated", nil)
	return task, nil
}

func (s *resourceCatalog) FailDelivery(ctx context.Context, taskID uint, token string, input xyRequest.FailResourceDelivery) (model.ResourceTask, error) {
	task, err := s.deliveryTask(ctx, taskID, token)
	if err != nil {
		return model.ResourceTask{}, err
	}
	if task.State == "completed" || task.State == "cancelled" {
		return model.ResourceTask{}, errors.New("资源任务已结束")
	}
	completedAt := time.Now()
	if err := global.GVA_DB.WithContext(ctx).Model(&task).Updates(map[string]any{
		"state": "failed", "stage": "failed", "error_code": input.ErrorCode,
		"error_message": input.ErrorMessage, "completed_at": completedAt,
	}).Error; err != nil {
		return model.ResourceTask{}, err
	}
	task.State, task.Stage, task.ErrorCode, task.ErrorMessage, task.CompletedAt = "failed", "failed", input.ErrorCode, input.ErrorMessage, &completedAt
	publishResourceTaskEvent(ctx, task, "task.updated", nil)
	return task, nil
}

func publishResourceTaskEvent(ctx context.Context, task model.ResourceTask, eventType string, extra map[string]any) {
	if task.AgentRunID == nil || Service.AgentEvents == nil {
		return
	}
	var run model.AgentRun
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", *task.AgentRunID, task.OwnerID).First(&run).Error; err != nil {
		return
	}
	data := map[string]any{
		"runId": run.ID, "taskId": task.ID, "state": task.State, "stage": task.Stage,
		"progressPercent": task.ProgressPercent, "progressDone": task.ProgressDone,
		"progressTotal": task.ProgressTotal, "progressUnit": task.ProgressUnit,
	}
	for key, value := range extra {
		data[key] = value
	}
	_, _ = Service.AgentEvents.Publish(run.SessionID, eventType, data)
}

func (s *resourceCatalog) deliveryTask(ctx context.Context, taskID uint, token string) (model.ResourceTask, error) {
	if taskID == 0 || token == "" {
		return model.ResourceTask{}, errors.New("交付授权无效")
	}
	var task model.ResourceTask
	if err := global.GVA_DB.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return model.ResourceTask{}, err
	}
	providedHash := sha256.Sum256([]byte(token))
	expectedHash, err := hex.DecodeString(task.DeliveryTokenHash)
	if err != nil || len(expectedHash) != len(providedHash) || subtle.ConstantTimeCompare(expectedHash, providedHash[:]) != 1 {
		return model.ResourceTask{}, errors.New("交付授权无效")
	}
	if task.DeliveryExpiresAt == nil || time.Now().After(*task.DeliveryExpiresAt) {
		return model.ResourceTask{}, errors.New("交付授权已过期")
	}
	return task, nil
}

func ensureResourceBucket(ctx context.Context, store resourceObjectStore, bucket string) error {
	exists, err := store.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("检查 RustFS bucket: %w", err)
	}
	if exists {
		return nil
	}
	if err := store.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		exists, checkErr := store.BucketExists(ctx, bucket)
		if checkErr != nil || !exists {
			return fmt.Errorf("创建 RustFS bucket: %w", err)
		}
	}
	return nil
}

func (s *resourceCatalog) storage() (resourceObjectStore, string, error) {
	if s.objectStore != nil {
		return s.objectStore, s.bucket, nil
	}
	endpoint := strings.TrimSpace(os.Getenv("RUSTFS_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("RUSTFS_ACCESS_KEY"))
	secretKey := strings.TrimSpace(os.Getenv("RUSTFS_SECRET_KEY"))
	bucket := strings.TrimSpace(os.Getenv("RUSTFS_BUCKET"))
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, "", errors.New("RustFS 配置不完整")
	}
	client, err := minio.New(parsed.Host, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: parsed.Scheme == "https",
	})
	if err != nil {
		return nil, "", err
	}
	return client, bucket, nil
}
