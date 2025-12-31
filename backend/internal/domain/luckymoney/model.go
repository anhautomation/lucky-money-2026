package luckymoney

import "time"

type User struct {
	ID         string
	Account    string
	Bank       string
	BankNo     string
	FullName   string
	Registered bool

	HasDrawn bool
	Amount   int
	DrawTime time.Time
}

type PoolItem struct {
	Amount int `json:"amount"`
	Qty    int `json:"qty"`
}

type Claim struct {
	ID       string    `json:"id"`
	Account  string    `json:"account"`
	Bank     string    `json:"bank"`
	BankNo   string    `json:"bankno"`
	FullName string    `json:"name"`
	Amount   int       `json:"amount"`
	Time     time.Time `json:"time"`
}
