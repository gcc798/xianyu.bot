package response

import "time"

type ReplyRule struct {
	ID           uint      `json:"ID"`
	CreatedAt    time.Time `json:"CreatedAt"`
	UpdatedAt    time.Time `json:"UpdatedAt"`
	ProductID    uint      `json:"productId"`
	ProductTitle string    `json:"productTitle"`
	Type         string    `json:"type"`
	Keyword      string    `json:"keyword"`
	Priority     int       `json:"priority"`
	Enabled      bool      `json:"enabled"`
	Content      string    `json:"content"`
}
