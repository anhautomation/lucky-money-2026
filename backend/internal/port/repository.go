package port

import "lucky-money/internal/domain/luckymoney"

type AdminSessionRepo interface {
	SetSession(token string) error
	IsValidSession(token string) bool
}

type IssuedIDRepo interface {
	IssueIDs(ids []string) error
	IsIssued(id string) bool
	ListIssued() []string
	DeleteIssued(id string) error
}

type UserRepo interface {
	GetByID(id string) (*luckymoney.User, bool)
	Save(u *luckymoney.User) error
}

type PoolRepo interface {
	SetPool(items []luckymoney.PoolItem) error
	GetPoolCounts() []luckymoney.PoolItem
	DrawOne() (amount int, ok bool)
	Remaining() int
}

type ClaimRepo interface {
	AppendClaim(c luckymoney.Claim) error
	ListClaims() []luckymoney.Claim
}

type Locker interface {
	WithLock(fn func())
}
