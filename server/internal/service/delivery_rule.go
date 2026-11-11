package service

import (
	"context"
	"errors"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"gorm.io/gorm"
)

// DeliveryRule 管理商品 SKU 的文本或卡密发货规则。
type DeliveryRule struct {
	db *gorm.DB // db 持久化自动发货规则。
}

// NewDeliveryRule 创建自动发货规则服务。
func NewDeliveryRule(db *gorm.DB) *DeliveryRule {
	return &DeliveryRule{db: db}
}

// List 分页查询 ownerID 名下规则并补充商品标题。
func (s *DeliveryRule) List(ctx context.Context, ownerID uint, input xyRequest.DeliveryRuleSearch) ([]xyResponse.DeliveryRule, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Table("xy_delivery_rules AS r").Where("r.owner_id = ? AND r.deleted_at IS NULL", ownerID)
	if input.ProductID != 0 {
		db = db.Where("r.product_id = ?", input.ProductID)
	}
	if input.EnabledOnly {
		db = db.Where("r.enabled = ?", true)
	}
	if input.Mode != "" {
		db = db.Where("r.mode = ?", input.Mode)
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("r.sku_key LIKE ? OR r.sku_name LIKE ? OR r.template LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]xyResponse.DeliveryRule, 0)
	err := db.Select(`r.id, r.created_at, r.updated_at, r.product_id, p.title AS product_title,
		r.sku_key, r.sku_name, r.mode, r.template, r.enabled`).
		Joins("JOIN xy_products AS p ON p.id = r.product_id AND p.deleted_at IS NULL").
		Scopes(input.Paginate()).Order("r.product_id DESC, r.sku_key DESC").Scan(&list).Error
	return list, total, err
}

// Save 校验商品归属、模式和模板后创建或更新发货规则。
func (s *DeliveryRule) Save(ctx context.Context, ownerID uint, input xyRequest.SaveDeliveryRule) error {
	input.SKUKey = strings.TrimSpace(input.SKUKey)
	input.SKUName = strings.TrimSpace(input.SKUName)
	input.Mode = strings.TrimSpace(input.Mode)
	input.Template = strings.TrimSpace(input.Template)
	if err := validateDeliveryRule(ownerID, input); err != nil {
		return err
	}
	db := s.db.WithContext(ctx)
	var product model.Product
	if err := db.Where("id = ? AND owner_id = ?", input.ProductID, ownerID).First(&product).Error; err != nil {
		return err
	}
	var rule model.DeliveryRule
	if input.ID != 0 {
		if err := db.Where("id = ? AND owner_id = ?", input.ID, ownerID).First(&rule).Error; err != nil {
			return err
		}
	}
	rule.OwnerID = ownerID
	rule.ProductID = input.ProductID
	rule.SKUKey = input.SKUKey
	rule.SKUName = input.SKUName
	rule.Mode = input.Mode
	rule.Template = input.Template
	rule.Enabled = input.Enabled
	if rule.ID == 0 {
		return db.Create(&rule).Error
	}
	return db.Save(&rule).Error
}

// Delete 校验 ownerID 归属后物理删除发货规则。
func (s *DeliveryRule) Delete(ctx context.Context, ownerID, id uint) error {
	if ownerID == 0 || id == 0 {
		return errors.New("发货规则 ID 不能为空")
	}
	result := s.db.WithContext(ctx).Unscoped().Where("id = ? AND owner_id = ?", id, ownerID).Delete(&model.DeliveryRule{})
	return affected(result)
}

// Resolve 优先匹配 skuKey，未命中时回退到商品默认规则。
func (s *DeliveryRule) Resolve(ctx context.Context, ownerID, productID uint, skuKey string) (*xyResponse.DeliveryRule, error) {
	skuKey = strings.TrimSpace(skuKey)
	var rule xyResponse.DeliveryRule
	db := s.db.WithContext(ctx).Table("xy_delivery_rules AS r").
		Select("r.id, r.product_id, r.sku_key, r.sku_name, r.mode, r.template, r.enabled").
		Where("r.owner_id = ? AND r.product_id = ? AND r.enabled = ? AND r.deleted_at IS NULL", ownerID, productID, true)
	if skuKey == "" {
		db = db.Where("r.sku_key = ''")
	} else {
		db = db.Where("r.sku_key IN (?, '')", skuKey).Order("r.sku_key DESC")
	}
	err := db.First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rule, err
}

// validateDeliveryRule 校验发货模式、SKU 长度和卡密占位符。
func validateDeliveryRule(ownerID uint, input xyRequest.SaveDeliveryRule) error {
	if ownerID == 0 || input.ProductID == 0 {
		return errors.New("商品 ID 不能为空")
	}
	if len([]rune(input.SKUKey)) > 128 || len([]rune(input.SKUName)) > 255 {
		return errors.New("SKU 信息长度超出限制")
	}
	if input.Mode != model.DeliveryModeText && input.Mode != model.DeliveryModeCard {
		return errors.New("不支持的发货模式")
	}
	if input.Template == "" || len([]rune(input.Template)) > 4000 {
		return errors.New("发货模板不能为空且不能超过 4000 个字符")
	}
	if input.Mode == model.DeliveryModeCard && !strings.Contains(input.Template, "{card}") {
		return errors.New("卡密模板必须包含 {card} 占位符")
	}
	return nil
}
