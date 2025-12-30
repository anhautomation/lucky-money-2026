package luckymoney

import "time"

type User struct {
	ID          string
	AccountName string
	Phone       string
	Registered  bool

	HasDrawn bool
	Amount   int
	DrawTime time.Time
}

type PoolItem struct {
	Amount int `json:"amount"`
	Qty    int `json:"qty"`
}

type Claim struct {
	ID          string    `json:"id"`
	AccountName string    `json:"account_name"`
	Phone       string    `json:"phone"`
	Amount      int       `json:"amount"`
	Time        time.Time `json:"time"`
}
