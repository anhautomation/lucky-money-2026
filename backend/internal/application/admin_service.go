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
	_ = s.sessions.SetSession(token)
	return token, nil
}

func (s *AdminService) Logout(token string) error {
	if !s.sessions.IsValidSession(token) {
		return luckymoney.ErrUnauthorized
	}
	_ = s.sessions.SetSession("")
	return nil
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
	u, ok := s.users.GetByID(id)
	if ok && u.HasDrawn {
		return luckymoney.ErrIDAlreadyDrawn
	}
	return s.issued.DeleteIssued(id)
}
