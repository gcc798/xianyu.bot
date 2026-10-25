package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model"
	_ "golang.org/x/image/webp"
)

const publishImageUploadURL = "https://stream-upload.goofish.com/api/upload.api?floderId=0&appkey=fleamarket&_input_charset=utf-8"

func reuploadPublishImages(ctx context.Context, ownerID, accountID uint, payload map[string]any) error {
	if _, err := ownedAccount(ctx, ownerID, accountID); err != nil {
		return err
	}
	images, ok := payload["imageInfoDOList"].([]any)
	if !ok || len(images) == 0 {
		return errors.New("续售发布快照缺少图片")
	}
	var credential model.Credential
	if err := global.GVA_DB.WithContext(ctx).Where("account_id = ?", accountID).First(&credential).Error; err != nil {
		return err
	}
	cookie, err := decryptCredential(credential.CookieCiphertext, credentialAAD(accountID, "cookie"))
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 90 * time.Second}
	for index, raw := range images {
		item, ok := raw.(map[string]any)
		if !ok || fmt.Sprint(item["type"]) != "0" {
			continue
		}
		source := strings.TrimSpace(fmt.Sprint(item["url"]))
		content, name, err := downloadPublishImage(ctx, client, source)
		if err != nil {
			return fmt.Errorf("准备第 %d 张续售图片: %w", index+1, err)
		}
		uploaded, err := uploadPublishImageContent(ctx, client, publishImageUploadURL, cookie, content, name)
		if err != nil {
			return fmt.Errorf("上传第 %d 张续售图片: %w", index+1, err)
		}
		uploaded["major"] = index == 0
		images[index] = uploaded
	}
	payload["imageInfoDOList"] = images
	return nil
}

func downloadPublishImage(ctx context.Context, client *http.Client, source string) ([]byte, string, error) {
	target, err := url.Parse(source)
	if err == nil && target.Scheme == "http" && allowedPublishImageHost(target.Hostname()) {
		target.Scheme = "https"
	}
	if err != nil || target.Scheme != "https" || !allowedPublishImageHost(target.Hostname()) {
		return nil, "", errors.New("图片必须来自闲鱼或阿里图片域名的 HTTPS 地址")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, "", err
	}
	request.Header.Set("Accept", "image/jpeg,image/png,image/*;q=0.8")
	request.Header.Set("Referer", "https://www.goofish.com/")
	request.Header.Set("User-Agent", webBrowserUA)
	response, err := client.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("图片下载 HTTP 状态 %d", response.StatusCode)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, (15<<20)+1))
	if err != nil {
		return nil, "", err
	}
	if len(content) == 0 || len(content) > 15<<20 {
		return nil, "", errors.New("图片为空或超过 15MB")
	}
	name := path.Base(target.Path)
	if name == "." || name == "/" || name == "" {
		name = "publish-image.jpg"
	}
	return content, name, nil
}

func allowedPublishImageHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, suffix := range []string{"alicdn.com", "goofish.com", "tbcdn.cn", "taobaocdn.com"} {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}

func uploadPublishImageContent(ctx context.Context, client *http.Client, endpoint, cookie string, content []byte, name string) (map[string]any, error) {
	config, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return nil, errors.New("图片格式无法识别")
	}
	name = "publish-image." + format
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(content); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Cookie", cookie)
	request.Header.Set("Origin", "https://seller.goofish.com")
	request.Header.Set("Referer", "https://seller.goofish.com/?site=COMMONPRO")
	request.Header.Set("User-Agent", webBrowserUA)
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errMTopTransport, err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errMTopTransport, err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("闲鱼图片上传 HTTP 状态 %d", response.StatusCode)
	}
	var result struct {
		Success bool `json:"success"`
		Object  struct {
			URL string `json:"url"`
			Pix string `json:"pix"`
		} `json:"object"`
	}
	if json.Unmarshal(raw, &result) != nil || !result.Success || strings.TrimSpace(result.Object.URL) == "" {
		return nil, errors.New("闲鱼图片上传未返回有效地址")
	}
	width, height := config.Width, config.Height
	if _, scanErr := fmt.Sscanf(strings.ToLower(result.Object.Pix), "%dx%d", &width, &height); scanErr != nil {
		width, height = config.Width, config.Height
	}
	return map[string]any{
		"extraInfo": map[string]string{"isH": "false", "isT": "false", "raw": "false"},
		"isQrCode":  false, "url": result.Object.URL, "heightSize": height, "widthSize": width,
		"major": false, "type": 0, "status": "done",
	}, nil
}
