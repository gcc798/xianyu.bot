package response

import "time"

type Product struct {
	ID                  uint      `json:"ID"`
	CreatedAt           time.Time `json:"CreatedAt"`
	UpdatedAt           time.Time `json:"UpdatedAt"`
	AccountID           uint      `json:"accountId"`
	AccountLabel        string    `json:"accountLabel"`
	ExternalProductID   string    `json:"externalProductId"`
	Title               string    `json:"title"`
	PriceCent           int64     `json:"priceCent"`
	Status              string    `json:"status"`
	ImageURL            string    `json:"imageUrl"`
	ImageURLs           []string  `json:"imageUrls"`
	SourceJSON          string    `json:"-"`
	AutoReplyEnabled    bool      `json:"autoReplyEnabled"`
	AutoDeliveryEnabled bool      `json:"autoDeliveryEnabled"`
	AutoRelistEnabled   bool      `json:"autoRelistEnabled"`
	ReplyDelaySeconds   int       `json:"replyDelaySeconds"`
	RelistDelaySeconds  int       `json:"relistDelaySeconds"`
	RelistDescription   string    `json:"relistDescription"`
	RelistImageURLs     string    `json:"relistImageUrls"`
	RelistAddress       string    `json:"relistAddress"`
}
