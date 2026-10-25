package response

import "time"

type DeliveryInventory struct {
	RuleID    uint  `json:"ruleId"`
	Available int64 `json:"available"`
	Reserved  int64 `json:"reserved"`
	Consumed  int64 `json:"consumed"`
}

type DeliveryTask struct {
	ID              uint       `json:"ID"`
	CreatedAt       time.Time  `json:"CreatedAt"`
	UpdatedAt       time.Time  `json:"UpdatedAt"`
	AccountID       uint       `json:"accountId"`
	AccountLabel    string     `json:"accountLabel"`
	ExternalOrderID string     `json:"externalOrderId"`
	ProductTitle    string     `json:"productTitle"`
	BuyerUserName   string     `json:"buyerUserName"`
	Mode            string     `json:"mode"`
	Status          string     `json:"status"`
	Attempts        int        `json:"attempts"`
	NextAttemptAt   time.Time  `json:"nextAttemptAt"`
	LastError       string     `json:"lastError"`
	SentAt          *time.Time `json:"sentAt"`
	ConfirmedAt     *time.Time `json:"confirmedAt"`
}
