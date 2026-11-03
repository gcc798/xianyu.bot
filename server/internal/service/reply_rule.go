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

type ReplyRule struct {
	db *gorm.DB
}

func NewReplyRule(db *gorm.DB) *ReplyRule {
	return &ReplyRule{db: db}
}

func (s *ReplyRule) List(ctx context.Context, ownerID uint, input xyRequest.ReplyRuleSearch) ([]xyResponse.ReplyRule, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Table("xy_reply_rules AS r").
		Joins("JOIN xy_reply_contents AS c ON c.rule_id = r.id AND c.deleted_at IS NULL").
		Where("r.owner_id = ? AND r.deleted_at IS NULL", ownerID)
	if input.ProductID != 0 {
		db = db.Where("r.product_id = ?", input.ProductID)
	}
	if input.Type != "" {
		if input.Type == model.ReplyRuleFuzzy {
			db = db.Where("r.type IN ?", []string{model.ReplyRuleExact, model.ReplyRuleFuzzy})
		} else {
			db = db.Where("r.type = ?", input.Type)
		}
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("r.keyword LIKE ? OR c.text LIKE ?", like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]xyResponse.ReplyRule, 0)
	err := db.Select(`r.id, r.created_at, r.updated_at, r.product_id, p.title AS product_title,
		r.type, r.keyword, r.priority, r.enabled, c.text AS content`).
		Joins("JOIN xy_products AS p ON p.id = r.product_id AND p.deleted_at IS NULL").
		Scopes(input.Paginate()).Order("CASE WHEN r.type = 'fallback' THEN 1 ELSE 0 END, r.priority DESC, r.id DESC").Scan(&list).Error
	return list, total, err
}

func (s *ReplyRule) Save(ctx context.Context, ownerID uint, input xyRequest.SaveReplyRule) error {
	input.Type = strings.TrimSpace(input.Type)
	input.Keyword = strings.TrimSpace(input.Keyword)
	input.Content = strings.TrimSpace(input.Content)
	if input.Type == model.ReplyRuleFallback {
		input.Keyword = ""
	}
	if err := validateReplyRule(ownerID, input); err != nil {
		return err
	}
	db := s.db.WithContext(ctx)
	var product model.Product
	if err := db.Where("id = ? AND owner_id = ?", input.ProductID, ownerID).First(&product).Error; err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var rule model.ReplyRule
		if input.ID != 0 {
			if err := tx.Where("id = ? AND owner_id = ?", input.ID, ownerID).First(&rule).Error; err != nil {
				return err
			}
			if input.Type == model.ReplyRuleExact && rule.Type != model.ReplyRuleExact {
				return errors.New("精确匹配仅用于兼容已有规则")
			}
		}
		rule.OwnerID = ownerID
		rule.ProductID = input.ProductID
		rule.Type = input.Type
		rule.Keyword = input.Keyword
		rule.Priority = input.Priority
		rule.Enabled = input.Enabled
		if rule.ID == 0 {
			if err := tx.Create(&rule).Error; err != nil {
				return err
			}
		} else if err := tx.Save(&rule).Error; err != nil {
			return err
		}
		var content model.ReplyContent
		err := tx.Where("rule_id = ?", rule.ID).First(&content).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		content.RuleID = rule.ID
		content.Text = input.Content
		if content.ID == 0 {
			return tx.Create(&content).Error
		}
		return tx.Save(&content).Error
	})
}

func (s *ReplyRule) Delete(ctx context.Context, ownerID, id uint) error {
	if ownerID == 0 || id == 0 {
		return errors.New("规则 ID 不能为空")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rule model.ReplyRule
		if err := tx.Where("id = ? AND owner_id = ?", id, ownerID).First(&rule).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("rule_id = ?", id).Delete(&model.ReplyContent{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&rule).Error
	})
}

func (s *ReplyRule) Resolve(ctx context.Context, ownerID, productID uint, message string) (*xyResponse.ReplyRule, error) {
	rules := make([]xyResponse.ReplyRule, 0)
	err := s.db.WithContext(ctx).Table("xy_reply_rules AS r").
		Select("r.id, r.product_id, r.type, r.keyword, r.priority, r.enabled, c.text AS content").
		Joins("JOIN xy_reply_contents AS c ON c.rule_id = r.id AND c.deleted_at IS NULL").
		Where("r.owner_id = ? AND r.product_id = ? AND r.enabled = ? AND r.deleted_at IS NULL", ownerID, productID, true).
		Order("r.priority DESC, r.id DESC").Scan(&rules).Error
	if err != nil {
		return nil, err
	}
	message = strings.TrimSpace(message)
	for i := range rules {
		rule := &rules[i]
		if !rule.Enabled || (rule.Type != model.ReplyRuleExact && rule.Type != model.ReplyRuleFuzzy) {
			continue
		}
		if strings.Contains(strings.ToLower(message), strings.ToLower(rule.Keyword)) {
			return rule, nil
		}
	}
	for i := range rules {
		rule := &rules[i]
		if rule.Enabled && rule.Type == model.ReplyRuleFallback {
			return rule, nil
		}
	}
	return nil, nil
}

func validateReplyRule(ownerID uint, input xyRequest.SaveReplyRule) error {
	if ownerID == 0 || input.ProductID == 0 {
		return errors.New("商品 ID 不能为空")
	}
	if input.Type != model.ReplyRuleFuzzy && input.Type != model.ReplyRuleFallback && input.Type != model.ReplyRuleExact {
		return errors.New("不支持的回复规则类型")
	}
	if input.Type == model.ReplyRuleExact && input.ID == 0 {
		return errors.New("不支持新建精确匹配规则")
	}
	if input.Type != model.ReplyRuleFallback && (input.Keyword == "" || len([]rune(input.Keyword)) > 255) {
		return errors.New("关键词不能为空且不能超过 255 个字符")
	}
	if input.Content == "" || len([]rune(input.Content)) > 4000 {
		return errors.New("回复内容不能为空且不能超过 4000 个字符")
	}
	if input.Priority < 0 || input.Priority > 10000 {
		return errors.New("优先级必须在 0 到 10000 之间")
	}
	return nil
}
