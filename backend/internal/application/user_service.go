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
	lock port.Locker,
) *UserService {
	return &UserService{
		issued: issued,
		users:  users,
		pool:   pool,
		claims: claims,
		lock:   lock,
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

type poolAtomicDrawCap interface {
	DrawAndCommit(u luckymoney.User) (int, error)
}

type poolRefundCap interface {
	RefundOne(amount int) error
}

func (s *UserService) SubmitAndDraw(id, account, bank, bankno, fullname string) (int, error) {
	if strings.TrimSpace(id) == "" {
		return 0, luckymoney.ErrIDNotIssued
	}

	var (
		amount int
		err    error
	)

	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && e == luckymoney.ErrDrawTimeout {
				amount = 0
				err = luckymoney.ErrDrawTimeout
				return
			}
			panic(r)
		}
	}()

	if s.lock != nil {
		s.lock.WithLock(func() {
			amount, err = s.submitAndDrawNoLock(id, account, bank, bankno, fullname)
		})
		return amount, err
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
		if err := s.users.Save(u); err != nil {
			return 0, err
		}
	} else {
		if u.Account != account || u.FullName != fullname {
			return 0, luckymoney.ErrInfoMismatch
		}
	}

	if u.HasDrawn {
		return 0, luckymoney.ErrAlreadyDrawn
	}

	if atomic, ok := s.pool.(poolAtomicDrawCap); ok {
		now := time.Now()
		snap := luckymoney.User{
			ID:         u.ID,
			Account:    u.Account,
			Bank:       u.Bank,
			BankNo:     u.BankNo,
			FullName:   u.FullName,
			Registered: true,
			HasDrawn:   true,
			Amount:     0,
			DrawTime:   now,
		}

		amt, err := atomic.DrawAndCommit(snap)
		if err != nil {
			if err == luckymoney.ErrPoolEmpty {
				return 0, luckymoney.ErrPoolEmpty
			}
			return 0, err
		}
		return amt, nil
	}

	amount, okDraw := s.pool.DrawOne()
	if !okDraw {
		return 0, luckymoney.ErrPoolEmpty
	}

	u.HasDrawn = true
	u.Amount = amount
	u.DrawTime = time.Now()

	if err := s.users.Save(u); err != nil {
		if rb, ok := s.pool.(poolRefundCap); ok {
			_ = rb.RefundOne(amount)
		}
		return 0, err
	}

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
