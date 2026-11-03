package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
)

type Account struct {
	db           *gorm.DB
	events       *BusinessEvent
	loginChecker loginChecker
	webSocket    *WebSocketService
}

func NewAccount(db *gorm.DB, events *BusinessEvent) *Account {
	return &Account{db: db, events: events}
}

func (s *Account) Create(ctx context.Context, ownerID uint, input xyRequest.CreateAccount) (model.Account, error) {
	label := strings.TrimSpace(input.Label)
	if err := validateAccount(ownerID, label, input.Note); err != nil {
		return model.Account{}, err
	}
	item := model.Account{OwnerID: ownerID, Label: label, Note: strings.TrimSpace(input.Note), Status: "unconfigured", Enabled: true}
	return item, s.db.WithContext(ctx).Create(&item).Error
}

func (s *Account) Update(ctx context.Context, ownerID uint, input xyRequest.UpdateAccount) error {
	label := strings.TrimSpace(input.Label)
	if input.ID == 0 {
		return errors.New("账号 ID 不能为空")
	}
	if err := validateAccount(ownerID, label, input.Note); err != nil {
		return err
	}
	result := s.db.WithContext(ctx).Model(&model.Account{}).
		Where("id = ? AND owner_id = ?", input.ID, ownerID).
		Updates(map[string]any{"label": label, "note": strings.TrimSpace(input.Note), "enabled": input.Enabled})
	if err := affected(result); err != nil {
		return err
	}
	if !input.Enabled && s.webSocket != nil {
		s.webSocket.stopRuntime(input.ID)
	}
	return nil
}

func (s *Account) Delete(ctx context.Context, ownerID, id uint) error {
	if ownerID == 0 || id == 0 {
		return errors.New("账号 ID 不能为空")
	}
	var item model.Account
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&item).Error; err != nil {
		return err
	}
	if s.webSocket != nil {
		s.webSocket.stopRuntime(id)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id = ?", id).Delete(&model.Credential{}).Error; err != nil {
			return err
		}
		return tx.Delete(&item).Error
	})
}

func (s *Account) List(ctx context.Context, ownerID uint, input xyRequest.AccountSearch) ([]model.Account, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Model(&model.Account{}).Where("owner_id = ?", ownerID)
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("label LIKE ? OR note LIKE ? OR external_user_id LIKE ?", like, like, like)
	}
	if input.Status != "" {
		db = db.Where("status = ?", input.Status)
	}
	if input.Enabled != nil {
		db = db.Where("enabled = ?", *input.Enabled)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]model.Account, 0)
	err := db.Select("xy_accounts.*", "EXISTS (SELECT 1 FROM xy_credentials WHERE xy_credentials.account_id = xy_accounts.id AND xy_credentials.deleted_at IS NULL) AS has_credential").
		Scopes(input.Paginate()).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (s *Account) SaveCredential(ctx context.Context, ownerID uint, input xyRequest.SaveCredential) error {
	if ownerID == 0 || input.ID == 0 {
		return errors.New("账号 ID 不能为空")
	}
	cookie := strings.TrimSpace(input.Cookie)
	if cookie == "" {
		return errors.New("Cookie 不能为空")
	}
	if len(cookie) > 64*1024 || len(input.Token) > 16*1024 {
		return errors.New("凭证长度超出限制")
	}
	var item model.Account
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", input.ID, ownerID).First(&item).Error; err != nil {
		return err
	}
	cipherCookie, err := encryptCredential(s.db, cookie, credentialAAD(input.ID, "cookie"))
	if err != nil {
		return err
	}
	cipherToken := ""
	if strings.TrimSpace(input.Token) != "" {
		cipherToken, err = encryptCredential(s.db, strings.TrimSpace(input.Token), credentialAAD(input.ID, "token"))
		if err != nil {
			return err
		}
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var credential model.Credential
		err := tx.Where("account_id = ?", input.ID).First(&credential).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			credential = model.Credential{AccountID: input.ID, CookieCiphertext: cipherCookie, TokenCiphertext: cipherToken, Version: 1}
			if err := tx.Create(&credential).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			if err := tx.Model(&credential).Updates(map[string]any{
				"cookie_ciphertext": cipherCookie,
				"token_ciphertext":  cipherToken,
				"version":           gorm.Expr("version + 1"),
			}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.Account{}).Where("id = ? AND owner_id = ?", input.ID, ownerID).Update("status", "ready").Error
	})
}

func validateAccount(ownerID uint, label, note string) error {
	if ownerID == 0 {
		return errors.New("用户未登录")
	}
	if label == "" || len([]rune(label)) > 80 {
		return errors.New("账号名称不能为空且不能超过 80 个字符")
	}
	if len([]rune(strings.TrimSpace(note))) > 500 {
		return errors.New("备注不能超过 500 个字符")
	}
	return nil
}

func credentialAAD(accountID uint, kind string) string {
	return fmt.Sprintf("xy/account/%d/%s", accountID, kind)
}

func affected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
