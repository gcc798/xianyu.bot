package service

import (
	"context"
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	xyResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/response"
)

type resourceCatalog struct{}

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
