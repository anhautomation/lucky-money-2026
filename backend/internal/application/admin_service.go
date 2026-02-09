package application

import (
	"lucky-money/internal/domain/luckymoney"
	"lucky-money/internal/port"
	"lucky-money/internal/store"
)

type AdminService struct {
	pool      port.PoolRepo
	issued    port.IssuedIDRepo
	claims    port.ClaimRepo
	sessions  port.AdminSessionRepo
	users     port.UserRepo
	adminUser string
	adminPass string
}

func NewAdminService(
	pool port.PoolRepo,
	issued port.IssuedIDRepo,
	claims port.ClaimRepo,
	sessions port.AdminSessionRepo,
	users port.UserRepo,
	adminUser string,
	adminPass string,
) *AdminService {
	return &AdminService{
		pool:      pool,
		issued:    issued,
		claims:    claims,
		sessions:  sessions,
		users:     users,
		adminUser: adminUser,
		adminPass: adminPass,
	}
}

func (s *AdminService) Login(user, pass string) (string, error) {
	if user != s.adminUser || pass != s.adminPass {
		return "", luckymoney.ErrInvalidCredential
	}
	token := store.NewToken()
	if err := s.sessions.SetSession(token); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AdminService) Logout(token string) error {
	_ = token
	return s.sessions.ClearSession()
}

func (s *AdminService) IssueIDs(ids []string) error {
	return s.issued.IssueIDs(ids)
}

func (s *AdminService) SetPool(items []luckymoney.PoolItem) error {
	return s.pool.SetPool(items)
}

func (s *AdminService) GetPool() []luckymoney.PoolItem {
	return s.pool.GetPoolCounts()
}

func (s *AdminService) GetClaims() []luckymoney.Claim {
	return s.claims.ListClaims()
}

func (s *AdminService) IsSessionValid(token string) bool {
	return s.sessions.IsValidSession(token)
}

func (s *AdminService) ListIssuedIDs() []string {
	return s.issued.ListIssued()
}

func (s *AdminService) DeleteIssuedID(id string) error {
	if l, ok := s.issued.(port.Locker); ok && l != nil {
		var err error
		l.WithLock(func() {
			err = s.issued.DeleteIssued(id)
		})
		return err
	}

	return s.issued.DeleteIssued(id)
}
