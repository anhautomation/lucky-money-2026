package application

import (
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

func NewUserService(issued port.IssuedIDRepo, users port.UserRepo, pool port.PoolRepo, claims port.ClaimRepo) *UserService {
	return &UserService{
		issued: issued,
		users:  users,
		pool:   pool,
		claims: claims,
		lock:   nil,
	}
}

func (s *UserService) Register(id, name, phone string) error {
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

	u.AccountName = name
	u.Phone = phone
	u.Registered = true
	u.HasDrawn = false
	u.Amount = 0
	u.DrawTime = time.Time{}

	return s.users.Save(u)
}

func (s *UserService) SubmitAndDraw(id, name, phone string) (int, error) {
	return s.submitAndDrawNoLock(id, name, phone)
}

func (s *UserService) submitAndDrawNoLock(id, name, phone string) (int, error) {
	u, ok := s.users.GetByID(id)
	if !ok {
		u = &luckymoney.User{ID: id}
	}

	if !u.Registered {
		u.AccountName = name
		u.Phone = phone
		u.Registered = true
		_ = s.users.Save(u)
	} else {
		if u.AccountName != name || u.Phone != phone {
			return 0, luckymoney.ErrInfoMismatch
		}
	}

	if u.HasDrawn {
		return u.Amount, luckymoney.ErrAlreadyDrawn
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
		ID:          u.ID,
		AccountName: u.AccountName,
		Phone:       u.Phone,
		Amount:      amount,
		Time:        u.DrawTime,
	})

	return amount, nil
}
