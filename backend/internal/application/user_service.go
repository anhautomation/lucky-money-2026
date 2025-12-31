package application

import (
	"strings"
	"time"

	"lucky-money/internal/domain/luckymoney"
	"lucky-money/internal/port"
)

type UserService struct {
	issued port.IssuedIDRepo
	users  port.UserRepo
	pool   port.PoolRepo
	claims port.ClaimRepo
	lock   port.Locker
}

func NewUserService(
	issued port.IssuedIDRepo,
	users port.UserRepo,
	pool port.PoolRepo,
	claims port.ClaimRepo,
) *UserService {
	return &UserService{
		issued: issued,
		users:  users,
		pool:   pool,
		claims: claims,
	}
}

func (s *UserService) Register(id, account, bank, bankno, fullname string) error {
	if strings.TrimSpace(id) == "" {
		return luckymoney.ErrIDNotIssued
	}

	if !s.issued.IsIssued(id) {
		return luckymoney.ErrIDNotIssued
	}

	u, ok := s.users.GetByID(id)
	if ok && u.Registered {
		return luckymoney.ErrAlreadyRegistered
	}

	if !ok {
		u = &luckymoney.User{ID: id}
	}

	u.Account = account
	u.Bank = bank
	u.BankNo = bankno
	u.FullName = fullname
	u.Registered = true
	u.HasDrawn = false
	u.Amount = 0
	u.DrawTime = time.Time{}

	return s.users.Save(u)
}

func (s *UserService) SubmitAndDraw(id, account, bank, bankno, fullname string) (int, error) {
	if strings.TrimSpace(id) == "" {
		return 0, luckymoney.ErrIDNotIssued
	}
	return s.submitAndDrawNoLock(id, account, bank, bankno, fullname)
}

func (s *UserService) submitAndDrawNoLock(id, account, bank, bankno, fullname string) (int, error) {
	if strings.TrimSpace(id) == "" {
		return 0, luckymoney.ErrIDNotIssued
	}

	if !s.issued.IsIssued(id) {
		return 0, luckymoney.ErrIDNotIssued
	}

	u, ok := s.users.GetByID(id)
	if !ok {
		u = &luckymoney.User{ID: id}
	}

	if !u.Registered {
		u.Account = account
		u.Bank = bank
		u.BankNo = bankno
		u.FullName = fullname
		u.Registered = true
		_ = s.users.Save(u)
	} else {
		if u.Account != account || u.FullName != fullname {
			return 0, luckymoney.ErrInfoMismatch
		}
	}

	if u.HasDrawn {
		return 0, luckymoney.ErrAlreadyDrawn
	}

	amount, okDraw := s.pool.DrawOne()
	if !okDraw {
		return 0, luckymoney.ErrPoolEmpty
	}

	u.HasDrawn = true
	u.Amount = amount
	u.DrawTime = time.Now()
	_ = s.users.Save(u)

	_ = s.claims.AppendClaim(luckymoney.Claim{
		ID:       u.ID,
		Account:  u.Account,
		Bank:     u.Bank,
		BankNo:   u.BankNo,
		FullName: u.FullName,
		Amount:   amount,
		Time:     u.DrawTime,
	})

	return amount, nil
}
