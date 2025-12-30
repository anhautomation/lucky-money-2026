package store

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	mrand "math/rand"
	"sync"
	"time"

	"lucky-money/internal/domain/luckymoney"
	"lucky-money/internal/port"
)

type database struct {
	Issued map[string]bool
	Users  map[string]*luckymoney.User
	Pool   []int
	Claims []luckymoney.Claim
}

func loadDatabase() database {
	// t1, _ := time.Parse(time.RFC3339, "2025-01-01T10:00:00Z")

	return database{
		Issued: map[string]bool{
			// "298e50bed4ddd674": true,
			// "3e53959825d08a61": true,
		},

		Pool: []int{
			// 50, 50, 100, 200,
		},

		Users: map[string]*luckymoney.User{
			// "298e50bed4ddd674": {
			// 	ID:          "298e50bed4ddd674",
			// 	AccountName: "Nguyen Van A",
			// 	Phone:       "0909123456",
			// 	HasDrawn:    true,
			// 	Amount:      100,
			// 	DrawTime:    t1,
			// },
		},

		// CLAIM HISTORY
		Claims: []luckymoney.Claim{
			// {
			// 	ID:          "298e50bed4ddd674",
			// 	AccountName: "Nguyen Van A",
			// 	Phone:       "0909123456",
			// 	Amount:      100,
			// 	Time:        t1,
			// },
		},
	}
}

type MemoryStore struct {
	mu         sync.Mutex
	adminToken string

	issued map[string]bool
	users  map[string]*luckymoney.User
	pool   []int
	claims []luckymoney.Claim
}

var _ port.AdminSessionRepo = (*MemoryStore)(nil)
var _ port.IssuedIDRepo = (*MemoryStore)(nil)
var _ port.UserRepo = (*MemoryStore)(nil)
var _ port.PoolRepo = (*MemoryStore)(nil)
var _ port.ClaimRepo = (*MemoryStore)(nil)
var _ port.Locker = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	mrand.Seed(time.Now().UnixNano())

	db := loadDatabase()

	return &MemoryStore{
		issued: db.Issued,
		users:  db.Users,
		pool:   db.Pool,
		claims: db.Claims,
	}
}

func (m *MemoryStore) WithLock(fn func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn()
}

func (m *MemoryStore) SetSession(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adminToken = token
	return nil
}

func (m *MemoryStore) IsValidSession(token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return token != "" && token == m.adminToken
}

func NewToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *MemoryStore) IssueIDs(ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		if id == "" {
			continue
		}
		m.issued[id] = true
	}
	return nil
}

func (m *MemoryStore) IsIssued(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.issued[id]
}

func (m *MemoryStore) ListIssued() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0, len(m.issued))
	for id := range m.issued {
		out = append(out, id)
	}
	return out
}

func (m *MemoryStore) DeleteIssued(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.issued, id)
	delete(m.users, id)
	return nil
}

func (m *MemoryStore) GetByID(id string) (*luckymoney.User, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

func (m *MemoryStore) Save(u *luckymoney.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *MemoryStore) SetPool(items []luckymoney.PoolItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pool = []int{}
	for _, it := range items {
		for i := 0; i < it.Qty; i++ {
			m.pool = append(m.pool, it.Amount)
		}
	}
	return nil
}

func (m *MemoryStore) GetPoolCounts() []luckymoney.PoolItem {
	m.mu.Lock()
	defer m.mu.Unlock()

	counter := map[int]int{}
	for _, a := range m.pool {
		counter[a]++
	}

	out := []luckymoney.PoolItem{}
	for amount, qty := range counter {
		out = append(out, luckymoney.PoolItem{Amount: amount, Qty: qty})
	}
	return out
}

func (m *MemoryStore) Remaining() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.pool)
}

func (m *MemoryStore) DrawOne() (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Println("[DRAW] pool size =", len(m.pool)) // 👈 THÊM DÒNG NÀY

	if len(m.pool) == 0 {
		return 0, false
	}
	idx := mrand.Intn(len(m.pool))
	amount := m.pool[idx]
	m.pool = append(m.pool[:idx], m.pool[idx+1:]...)
	return amount, true
}

func (m *MemoryStore) AppendClaim(c luckymoney.Claim) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claims = append(m.claims, c)
	return nil
}

func (m *MemoryStore) ListClaims() []luckymoney.Claim {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]luckymoney.Claim, len(m.claims))
	copy(out, m.claims)
	return out
}
