package agent

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type telegramDialogResult struct {
	ID            int64      `json:"telegram_id"`
	AccessHash    *int64     `json:"access_hash"`
	Name          string     `json:"name"`
	Username      string     `json:"username"`
	DialogType    string     `json:"dialog_type"`
	MemberCount   *int       `json:"member_count"`
	IsChannel     bool       `json:"is_channel"`
	IsGroup       bool       `json:"is_group"`
	IsMegagroup   bool       `json:"is_megagroup"`
	IsBroadcast   bool       `json:"is_broadcast"`
	IsArchived    bool       `json:"is_archived"`
	IsMuted       bool       `json:"is_muted"`
	UnreadCount   int        `json:"unread_count"`
	LastMessageID *int64     `json:"last_message_id"`
	LastMessageAt *time.Time `json:"last_message_at"`
	LastReadID    *int64     `json:"last_read_message_id"`
	Status        string     `json:"status"`
	IsCreator     bool       `json:"is_creator"`
}

type telegramMessageResult struct {
	ID   int64 `json:"id"`
	File *struct {
		Name     string `json:"name"`
		Size     int64  `json:"size"`
		MIMEType string `json:"mime_type"`
	} `json:"file"`
}

type telegramMessagesResult struct {
	Status string `json:"status"`
	Data   struct {
		Messages []telegramMessageResult `json:"messages"`
	} `json:"data"`
}

var adultTelegramSourcePattern = regexp.MustCompile(`成人|色情|情色|成人视频|裸聊|裸舞|裸播|裸照|约炮|炮友|一夜情|援交|福利姬|福利视频|福利资源|熟女|少妇|人妻|萝莉|乱伦|母狗|调教|换妻|群交|迷奸|强奸|自拍偷拍|国产自拍|鲍鱼资源|草榴|杏吧|色花堂|海角社区|stripchat|onlyfans|pornhub|xvideos|xhamster|redtube|18禁|无码|有码|av女优|\b(porn|xxx|nsfw|jav)\b`)

// ListTelegramSources 返回当前用户可用的 Telegram 来源。
func (s *ResourceCatalog) ListTelegramSources(ctx context.Context) ([]model.TelegramDialog, error) {
	sources := make([]model.TelegramDialog, 0)
	adultMode, adultIDs, err := s.telegramAdultSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	err = s.db.WithContext(ctx).Where("status = ? AND resource_compliance_status = ? AND (is_group OR is_broadcast)", "normal", "allowed").Order("last_message_at DESC NULLS LAST, name").Find(&sources).Error
	if err != nil {
		return nil, err
	}
	return filterTelegramSources(sources, adultMode, adultIDs), nil
}

func (s *ResourceCatalog) telegramAdultSettings(ctx context.Context, db *gorm.DB) (bool, map[int64]bool, error) {
	var settings []struct {
		Key   string
		Value string
	}
	if err := db.WithContext(ctx).Table("tel_settings").Where("key IN ?", []string{"adult_mode", "adult_dialog_ids"}).Find(&settings).Error; err != nil {
		return false, nil, err
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}
	if values["adult_mode"] == "true" {
		return true, nil, nil
	}
	var configuredIDs []int64
	if value := values["adult_dialog_ids"]; value != "" {
		if err := json.Unmarshal([]byte(value), &configuredIDs); err != nil {
			return false, nil, errors.New("Telegram 来源合规配置无效")
		}
	}
	adultIDs := make(map[int64]bool, len(configuredIDs))
	for _, id := range configuredIDs {
		adultIDs[id] = true
	}
	return false, adultIDs, nil
}

func filterTelegramSources(sources []model.TelegramDialog, adultMode bool, adultIDs map[int64]bool) []model.TelegramDialog {
	if adultMode {
		return sources
	}
	filtered := sources[:0]
	for _, source := range sources {
		if adultIDs[int64(source.ID)] || adultTelegramSourcePattern.MatchString(strings.ToLower(source.Name+" "+source.Username)) {
			continue
		}
		filtered = append(filtered, source)
	}
	return filtered
}

// TelegramResourceOptions 返回批量采集的文件类型和大小限制。
func (s *ResourceCatalog) TelegramResourceOptions() model.TelegramResourceOptions {
	return model.TelegramResourceOptions{
		FileTypes:   []string{"pdf", "md", "txt", "csv", "json", "xml", "epub", "mobi", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "rtf", "odt", "ods", "odp", "jpg", "jpeg", "png", "webp", "mp4", "webm", "mov", "m4v", "mp3", "m4a", "wav", "ogg", "zip"},
		MaxFileSize: 200 * 1024 * 1024,
	}
}

// SyncTelegramSources 从 Telegram MCP 同步来源目录。
func (s *ResourceCatalog) SyncTelegramSources(ctx context.Context) ([]model.TelegramDialog, error) {
	connections, err := s.telegramRegistry().Discover(ctx)
	if err != nil {
		return nil, errors.New("Telegram MCP 当前不可用")
	}
	defer func() { _ = closeMCPConnections(connections) }()
	connection, definition := findTelegramTool(connections, "dialog_list")
	if connection == nil {
		return nil, errors.New("Telegram MCP 未提供 dialog_list")
	}
	var response struct {
		Status string `json:"status"`
		Data   struct {
			Dialogs []telegramDialogResult `json:"dialogs"`
			Total   int                    `json:"total"`
		} `json:"data"`
	}
	for page := 1; ; page++ {
		requestContext, cancel := context.WithTimeout(ctx, time.Duration(connection.Server.RequestTimeoutSeconds)*time.Second)
		result, callErr := connection.Client.CallTool(requestContext, mcp.CallToolRequest{Params: mcp.CallToolParams{
			Name: definition.Name, Arguments: map[string]any{"page": page, "page_size": 100, "kind": ""},
		}})
		cancel()
		if callErr != nil || result.IsError {
			return nil, errors.New("Telegram 来源列表读取失败")
		}
		payload, marshalErr := json.Marshal(mcpToolResultValue(result))
		if marshalErr != nil {
			return nil, marshalErr
		}
		var pageResult struct {
			Status string `json:"status"`
			Data   struct {
				Dialogs []telegramDialogResult `json:"dialogs"`
				Total   int                    `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(payload, &pageResult); err != nil || pageResult.Status != "success" {
			return nil, errors.New("Telegram 来源列表响应无效")
		}
		response.Data.Dialogs = append(response.Data.Dialogs, pageResult.Data.Dialogs...)
		response.Data.Total = pageResult.Data.Total
		if len(pageResult.Data.Dialogs) == 0 || page*100 >= response.Data.Total {
			break
		}
	}
	now := time.Now()
	for _, item := range response.Data.Dialogs {
		if item.ID == 0 || strings.TrimSpace(item.Name) == "" {
			continue
		}
		source := model.TelegramDialog{
			TelegramID: item.ID, AccessHash: item.AccessHash, Name: item.Name, Username: item.Username,
			DialogType: item.DialogType, MemberCount: item.MemberCount, IsChannel: item.IsChannel,
			IsGroup: item.IsGroup, IsMegagroup: item.IsMegagroup, IsBroadcast: item.IsBroadcast,
			IsArchived: item.IsArchived, IsMuted: item.IsMuted, UnreadCount: item.UnreadCount,
			LastMessageID: item.LastMessageID, LastMessageAt: item.LastMessageAt, LastReadMessageID: item.LastReadID,
			Status: item.Status, IsCreator: item.IsCreator, ResourceComplianceStatus: "allowed",
			CreatedAt: now, UpdatedAt: now,
		}
		if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "telegram_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"access_hash", "name", "username", "dialog_type", "member_count", "is_channel", "is_group",
				"is_megagroup", "is_broadcast", "is_archived", "is_muted", "unread_count", "last_message_id",
				"last_message_at", "last_read_message_id", "status", "is_creator", "updated_at",
			}),
		}).Create(&source).Error; err != nil {
			return nil, fmt.Errorf("保存 Telegram 来源: %w", err)
		}
	}
	return s.ListTelegramSources(ctx)
}

// CreateTelegramTask 校验权限范围并创建 Telegram 批量采集任务。
func (s *ResourceCatalog) CreateTelegramTask(ctx context.Context, ownerID uint, input xyRequest.CreateTelegramResourceTask) (model.TelegramResourceTask, error) {
	if ownerID == 0 || strings.TrimSpace(input.Keyword) == "" || input.MaxDownloads < 1 || input.MaxDownloads > 100 || !slices.Contains([]string{"smart", "all", "manual"}, input.SourceMode) {
		return model.TelegramResourceTask{}, errors.New("Telegram 采集任务参数无效")
	}
	input.Keyword = strings.TrimSpace(input.Keyword)
	input.FileTypes = uniqueStrings(input.FileTypes)
	allowedTypes := s.TelegramResourceOptions().FileTypes
	for _, fileType := range input.FileTypes {
		if !slices.Contains(allowedTypes, fileType) {
			return model.TelegramResourceTask{}, errors.New("文件类型无效")
		}
	}
	slices.Sort(input.SourceDialogIDs)
	input.SourceDialogIDs = slices.Compact(input.SourceDialogIDs)
	if input.SourceMode == "manual" && len(input.SourceDialogIDs) == 0 {
		return model.TelegramResourceTask{}, errors.New("手动模式至少选择一个来源")
	}
	strategy := "pending"
	if input.SourceMode == "all" {
		strategy = "all"
	}
	if input.SourceMode == "manual" {
		strategy = "manual"
	}
	task := model.TelegramResourceTask{
		OwnerID: ownerID, Keyword: input.Keyword, MaxDownloads: input.MaxDownloads,
		SourceMode: input.SourceMode, SourceStrategy: strategy, SourceDialogIDs: input.SourceDialogIDs,
		FileTypes: input.FileTypes, Status: "queued",
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if input.SourceMode == "manual" {
			adultMode, adultIDs, err := s.telegramAdultSettings(ctx, tx)
			if err != nil {
				return err
			}
			var selected []model.TelegramDialog
			if err := tx.Where("id IN ? AND status = ? AND resource_compliance_status = ? AND (is_group OR is_broadcast)", input.SourceDialogIDs, "normal", "allowed").Find(&selected).Error; err != nil {
				return err
			}
			if len(filterTelegramSources(selected, adultMode, adultIDs)) != len(input.SourceDialogIDs) {
				return errors.New("选择的 Telegram 来源不可用")
			}
		}
		return tx.Create(&task).Error
	})
	return task, err
}

// ListTelegramTasks 按所有者分页查询 Telegram 采集任务。
func (s *ResourceCatalog) ListTelegramTasks(ctx context.Context, ownerID uint, input xyRequest.ResourceTaskSearch) ([]model.TelegramResourceTask, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Model(&model.TelegramResourceTask{}).Where("owner_id = ?", ownerID)
	if input.State != "" {
		db = db.Where("status = ?", input.State)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	tasks := make([]model.TelegramResourceTask, 0)
	err := db.Scopes(input.Paginate()).Order("id DESC").Find(&tasks).Error
	return tasks, total, err
}

// CancelTelegramTask 取消所有者名下的 Telegram 采集任务。
func (s *ResourceCatalog) CancelTelegramTask(ctx context.Context, ownerID, taskID uint) (model.TelegramResourceTask, error) {
	if ownerID == 0 || taskID == 0 {
		return model.TelegramResourceTask{}, errors.New("任务 ID 无效")
	}
	var task model.TelegramResourceTask
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", taskID, ownerID).First(&task).Error; err != nil {
		return task, err
	}
	if task.Status == "queued" {
		task.Status, task.SourceStrategy = "cancelled", "cancelled"
		completedAt := time.Now()
		task.CompletedAt = &completedAt
		return task, s.db.WithContext(ctx).Model(&task).Updates(map[string]any{"status": task.Status, "source_strategy": task.SourceStrategy, "completed_at": task.CompletedAt}).Error
	}
	if task.Status != "running" {
		return model.TelegramResourceTask{}, errors.New("任务当前不能取消")
	}
	if err := s.db.WithContext(ctx).Model(&task).Update("cancel_requested", true).Error; err != nil {
		return model.TelegramResourceTask{}, err
	}
	task.CancelRequested = true
	return task, nil
}

// StartTelegramTasks 恢复并执行持久化队列中的采集任务。
func (s *ResourceCatalog) StartTelegramTasks(ctx context.Context) {
	go func() {
		db := s.db.WithContext(ctx).Model(&model.TelegramResourceTask{}).Where("status = ?", "running")
		_ = db.Where("cancel_requested = ?", true).Updates(map[string]any{
			"status": "cancelled", "completed_at": time.Now(),
		}).Error
		_ = db.Where("cancel_requested = ?", false).Updates(map[string]any{
			"status": "queued", "started_at": nil, "files_found": 0, "files_completed": 0,
		}).Error
		for {
			var task model.TelegramResourceTask
			err := s.db.WithContext(ctx).Where("status = ?", "queued").Order("id").First(&task).Error
			if err == nil {
				claim := s.db.WithContext(ctx).Model(&model.TelegramResourceTask{}).
					Where("id = ? AND status = ?", task.ID, "queued").Updates(map[string]any{"status": "running", "started_at": time.Now()})
				if claim.Error == nil && claim.RowsAffected == 1 {
					task.Status = "running"
					s.runTelegramTask(ctx, task)
				}
				continue
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
}

func (s *ResourceCatalog) runTelegramTask(ctx context.Context, task model.TelegramResourceTask) {
	fail := func() {
		_ = s.db.WithContext(context.WithoutCancel(ctx)).Model(&task).Updates(map[string]any{
			"status": "failed", "error_message": "Telegram 采集失败，请检查 MCP 与对象存储状态", "completed_at": time.Now(),
		}).Error
	}
	sourceIDs := task.SourceDialogIDs
	fileTypes := task.FileTypes
	if task.SourceMode == "smart" || task.SourceMode == "all" {
		sources, err := s.ListTelegramSources(ctx)
		if err != nil {
			fail()
			return
		}
		if task.SourceMode == "smart" {
			sources, task.SourceStrategy = s.rankTelegramSources(ctx, task.Keyword, sources)
		}
		sourceIDs = make([]uint, len(sources))
		for i := range sources {
			sourceIDs[i] = sources[i].ID
		}
		task.SourceDialogIDs = sourceIDs
		if err := s.db.WithContext(ctx).Model(&task).Select("source_selection", "source_strategy").Updates(&task).Error; err != nil {
			fail()
			return
		}
	}
	var sources []model.TelegramDialog
	if err := s.db.WithContext(ctx).Where("id IN ? AND status = ? AND resource_compliance_status = ?", sourceIDs, "normal", "allowed").Find(&sources).Error; err != nil {
		fail()
		return
	}
	connections, err := s.telegramRegistry().Discover(ctx)
	if err != nil {
		fail()
		return
	}
	defer func() { _ = closeMCPConnections(connections) }()
	connection, searchTool := findTelegramTool(connections, "message_search")
	var uploadTool mcp.Tool
	if connection != nil {
		for _, definition := range connection.Tools {
			if definition.Name == "telegram_upload" {
				uploadTool = definition
				break
			}
		}
	}
	if connection == nil || searchTool.Name == "" || uploadTool.Name == "" {
		fail()
		return
	}
	found, completed, failures := 0, 0, 0
	for _, source := range sources {
		if found >= task.MaxDownloads || telegramTaskCancelled(ctx, s.db, task.ID) {
			break
		}
		searchContext, cancelSearch := context.WithTimeout(ctx, time.Duration(connection.Server.TaskTimeoutSeconds)*time.Second)
		result, err := connection.Client.CallTool(searchContext, mcp.CallToolRequest{Params: mcp.CallToolParams{
			Name: searchTool.Name, Arguments: map[string]any{"dialog_id": float64(source.TelegramID), "query": task.Keyword, "limit": 100},
		}})
		cancelSearch()
		if err != nil || result.IsError {
			failures++
			continue
		}
		payload, err := json.Marshal(mcpToolResultValue(result))
		if err != nil {
			failures++
			continue
		}
		var messages telegramMessagesResult
		if json.Unmarshal(payload, &messages) != nil || messages.Status != "success" {
			failures++
			continue
		}
		for _, message := range messages.Data.Messages {
			if found >= task.MaxDownloads || telegramTaskCancelled(ctx, s.db, task.ID) {
				break
			}
			if message.ID <= 0 || message.File == nil || message.File.Size > s.TelegramResourceOptions().MaxFileSize || !telegramFileAllowed(message.File.Name, message.File.Size, fileTypes) {
				continue
			}
			externalID := fmt.Sprintf("%d:%d", source.TelegramID, message.ID)
			var existing int64
			if err := s.db.WithContext(ctx).Table("xy_resource_sources AS rs").
				Joins("JOIN xy_resources AS r ON r.id = rs.resource_id").
				Where("r.owner_id = ? AND rs.mcp_server_code = ? AND rs.external_id = ?", task.OwnerID, "telegram", externalID).
				Count(&existing).Error; err != nil {
				fail()
				return
			}
			if existing > 0 {
				found++
				completed++
				_ = s.db.WithContext(ctx).Model(&task).Updates(map[string]any{"files_found": found, "files_completed": completed}).Error
				continue
			}
			found++
			_ = s.db.WithContext(ctx).Model(&task).Update("files_found", found).Error
			key := fmt.Sprintf("telegram/%d/%d/%d-%d", task.OwnerID, task.ID, source.TelegramID, message.ID)
			store, bucket, err := s.storage()
			if err != nil {
				fail()
				return
			}
			if err := ensureResourceBucket(ctx, store, bucket); err != nil {
				fail()
				return
			}
			uploadURL, err := store.PresignedPutObject(ctx, bucket, key, resourceDeliveryTTL)
			if err != nil {
				failures++
				continue
			}
			uploadContext, cancel := context.WithTimeout(ctx, time.Duration(connection.Server.TaskTimeoutSeconds)*time.Second)
			uploadResult, uploadErr := connection.Client.CallTool(uploadContext, mcp.CallToolRequest{Params: mcp.CallToolParams{
				Name: uploadTool.Name, Arguments: map[string]any{"telegram_id": float64(source.TelegramID), "message_id": float64(message.ID), "upload_url": uploadURL.String()},
			}})
			cancel()
			if uploadErr != nil || uploadResult.IsError {
				_ = store.RemoveObject(context.WithoutCancel(ctx), bucket, key, minio.RemoveObjectOptions{})
				failures++
				continue
			}
			encoded, err := json.Marshal(mcpToolResultValue(uploadResult))
			if err != nil {
				_ = store.RemoveObject(context.WithoutCancel(ctx), bucket, key, minio.RemoveObjectOptions{})
				failures++
				continue
			}
			var uploaded struct {
				Status string `json:"status"`
				Data   struct {
					Name     string `json:"name"`
					Size     int64  `json:"size"`
					MIMEType string `json:"mime_type"`
					SHA256   string `json:"sha256"`
				} `json:"data"`
			}
			if json.Unmarshal(encoded, &uploaded) != nil || uploaded.Status != "success" {
				_ = store.RemoveObject(context.WithoutCancel(ctx), bucket, key, minio.RemoveObjectOptions{})
				failures++
				continue
			}
			_, err = s.RegisterTelegramResource(ctx, xyRequest.RegisterTelegramResource{
				OwnerID: task.OwnerID, Name: uploaded.Data.Name, Size: uploaded.Data.Size,
				MIMEType: uploaded.Data.MIMEType, SHA256: uploaded.Data.SHA256, ObjectKey: key,
				SourceName: source.Name, SourceReference: fmt.Sprintf("telegram://%d/%d", source.TelegramID, message.ID),
				ExternalID: externalID,
			})
			if err == nil {
				completed++
			} else {
				_ = store.RemoveObject(context.WithoutCancel(ctx), bucket, key, minio.RemoveObjectOptions{})
				failures++
			}
			_ = s.db.WithContext(ctx).Model(&task).Updates(map[string]any{"files_found": found, "files_completed": completed}).Error
		}
	}
	if ctx.Err() != nil {
		return
	}
	status := "completed"
	errorMessage := ""
	if telegramTaskCancelled(ctx, s.db, task.ID) {
		status = "cancelled"
	} else if failures > 0 {
		errorMessage = "部分来源检索或文件交付失败"
		if completed == 0 {
			status = "failed"
			errorMessage = "Telegram 来源检索或文件交付失败"
		}
	}
	_ = s.db.WithContext(context.WithoutCancel(ctx)).Model(&task).Updates(map[string]any{
		"files_found": found, "files_completed": completed, "status": status, "completed_at": time.Now(), "error_message": errorMessage,
	}).Error
}

func (s *ResourceCatalog) rankTelegramSources(ctx context.Context, keyword string, sources []model.TelegramDialog) ([]model.TelegramDialog, string) {
	local := append([]model.TelegramDialog(nil), sources...)
	slices.SortStableFunc(local, func(a, b model.TelegramDialog) int {
		return cmp.Compare(telegramSourceScore(keyword, b), telegramSourceScore(keyword, a))
	})
	var models []model.AIModel
	if s.db.WithContext(ctx).Where("enabled = ? AND api_key <> ''", true).Order("priority, id").Find(&models).Error != nil || len(models) == 0 {
		return local, "metadata"
	}
	metadata := make([]map[string]any, 0, len(local))
	for _, source := range local {
		metadata = append(metadata, map[string]any{"id": source.ID, "name": source.Name, "username": source.Username, "type": source.DialogType})
	}
	encoded, _ := json.Marshal(metadata)
	for _, configured := range models {
		chat, err := newConfiguredAgentModel(ctx, configured, model.DefaultAgentConfig())
		if err != nil {
			continue
		}
		response, err := chat.Generate(ctx, []*schema.Message{
			{Role: schema.System, Content: `根据搜索关键词对来源名称进行相关性排序。只输出 JSON：{"source_ids":[整数ID]}，不得编造 ID。`},
			{Role: schema.User, Content: fmt.Sprintf("关键词：%s\n来源：%s", keyword, encoded)},
		})
		if err != nil {
			continue
		}
		var ranked struct {
			SourceIDs []uint `json:"source_ids"`
		}
		if json.Unmarshal([]byte(response.Content), &ranked) != nil {
			continue
		}
		byID := make(map[uint]model.TelegramDialog, len(local))
		for _, source := range local {
			byID[source.ID] = source
		}
		ordered := make([]model.TelegramDialog, 0, len(local))
		seen := make(map[uint]bool, len(local))
		for _, id := range ranked.SourceIDs {
			if source, ok := byID[id]; ok && !seen[id] {
				ordered = append(ordered, source)
				seen[id] = true
			}
		}
		for _, source := range local {
			if !seen[source.ID] {
				ordered = append(ordered, source)
			}
		}
		return ordered, "ai"
	}
	return local, "metadata"
}

func telegramSourceScore(keyword string, source model.TelegramDialog) int {
	name := strings.ToLower(source.Name + " " + source.Username)
	query := strings.ToLower(keyword)
	if strings.Contains(name, query) {
		return 100
	}
	return 0
}

func findTelegramTool(connections []*mcpConnection, name string) (*mcpConnection, mcp.Tool) {
	for _, connection := range connections {
		for _, definition := range connection.Tools {
			if definition.Name == name {
				return connection, definition
			}
		}
	}
	return nil, mcp.Tool{}
}

func (s *ResourceCatalog) telegramRegistry() *MCPRegistry {
	if s.registry != nil {
		return s.registry
	}
	return NewMCPRegistry(s.db)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(value), "."))
		if value != "" && !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}
	return unique
}

func telegramFileAllowed(name string, size int64, types []string) bool {
	if size <= 0 {
		return false
	}
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return false
	}
	extension := strings.ToLower(parts[len(parts)-1])
	if len(types) == 0 {
		return true
	}
	for _, fileType := range types {
		if extension == fileType {
			return true
		}
	}
	return false
}

func telegramTaskCancelled(ctx context.Context, db *gorm.DB, taskID uint) bool {
	var task model.TelegramResourceTask
	if err := db.WithContext(ctx).Select("status, cancel_requested").First(&task, taskID).Error; err != nil {
		return true
	}
	return task.CancelRequested || task.Status == "cancelled"
}
