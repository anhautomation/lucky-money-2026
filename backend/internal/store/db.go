package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	mrand "math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"lucky-money/internal/domain/luckymoney"
	"lucky-money/internal/port"
)

func NewToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type Store struct {
	mu sync.Mutex
	db *sql.DB
}

var _ port.AdminSessionRepo = (*Store)(nil)
var _ port.IssuedIDRepo = (*Store)(nil)
var _ port.UserRepo = (*Store)(nil)
var _ port.PoolRepo = (*Store)(nil)
var _ port.ClaimRepo = (*Store)(nil)
var _ port.Locker = (*Store)(nil)

func NewStore() *Store {
	mrand.Seed(time.Now().UnixNano())

	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./data/lucky.db"
	}

	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.init(); err != nil {
		panic(err)
	}
	return s
}

func (m *Store) init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmts := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA busy_timeout = 5000;`,

		`CREATE TABLE IF NOT EXISTS admin_session (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			token TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS issued_ids (
			id TEXT PRIMARY KEY
		);`,

		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			account TEXT NOT NULL DEFAULT '',
			bank TEXT NOT NULL DEFAULT '',
			bank_no TEXT NOT NULL DEFAULT '',
			full_name TEXT NOT NULL DEFAULT '',
			registered INTEGER NOT NULL DEFAULT 0,
			has_drawn INTEGER NOT NULL DEFAULT 0,
			amount INTEGER NOT NULL DEFAULT 0,
			draw_time TEXT NOT NULL DEFAULT '' -- RFC3339
		);`,

		`CREATE TABLE IF NOT EXISTS pool_counts (
			amount INTEGER PRIMARY KEY,
			qty INTEGER NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS claims (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			account TEXT NOT NULL DEFAULT '',
			bank TEXT NOT NULL DEFAULT '',
			bank_no TEXT NOT NULL DEFAULT '',
			full_name TEXT NOT NULL DEFAULT '',
			amount INTEGER NOT NULL,
			time TEXT NOT NULL -- RFC3339
		);`,
	}

	for _, st := range stmts {
		if _, err := m.db.ExecContext(ctx, st); err != nil {
			log.Println("[ERR][init.exec]", err)
			return err
		}
	}
	return nil
}

func (m *Store) WithLock(fn func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn()
}

func (m *Store) SetSession(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token == "" {
		return errors.New("empty token")
	}

	_, err := m.db.Exec(`
		INSERT INTO admin_session(id, token)
		VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET token=excluded.token
	`, token)
	if err != nil {
		log.Println("[ERR][SetSession]", err)
	}
	return err
}

func (m *Store) ClearSession() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.db.Exec(`DELETE FROM admin_session WHERE id=1`)
	if err != nil {
		log.Println("[ERR][ClearSession]", err)
	}
	return err
}

func (m *Store) IsValidSession(token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token == "" {
		return false
	}

	var saved string
	if err := m.db.QueryRow(`SELECT token FROM admin_session WHERE id=1`).Scan(&saved); err != nil {
		return false
	}
	return token == saved
}

func (m *Store) IssueIDs(ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, err := m.db.Begin()
	if err != nil {
		log.Println("[ERR][IssueIDs.begin]", err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO issued_ids(id) VALUES (?)`)
	if err != nil {
		log.Println("[ERR][IssueIDs.prepare]", err)
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, err := stmt.Exec(id); err != nil {
			log.Println("[ERR][IssueIDs.exec]", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("[ERR][IssueIDs.commit]", err)
		return err
	}
	return nil
}

func (m *Store) IsIssued(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		return false
	}

	var x string
	err := m.db.QueryRow(`SELECT id FROM issued_ids WHERE id=?`, id).Scan(&x)
	return err == nil && x == id
}

func (m *Store) ListIssued() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows, err := m.db.Query(`SELECT id FROM issued_ids ORDER BY id`)
	if err != nil {
		log.Println("[ERR][ListIssued.query]", err)
		return nil
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Println("[ERR][ListIssued.scan]", err)
			continue
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		log.Println("[ERR][ListIssued.rows]", err)
	}
	return out
}

func (m *Store) DeleteIssued(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		return nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		log.Println("[ERR][DeleteIssued.begin]", err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM issued_ids WHERE id=?`, id); err != nil {
		log.Println("[ERR][DeleteIssued.ids]", err)
		return err
	}

	if _, err := tx.Exec(`DELETE FROM users WHERE id=?`, id); err != nil {
		log.Println("[ERR][DeleteIssued.users]", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println("[ERR][DeleteIssued.commit]", err)
		return err
	}
	return nil
}

func (m *Store) GetByID(id string) (*luckymoney.User, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		return nil, false
	}

	var u luckymoney.User
	var reg, drawn int
	var drawTime string

	err := m.db.QueryRow(`
		SELECT id, account, bank, bank_no, full_name, registered, has_drawn, amount, draw_time
		FROM users WHERE id=?
	`, id).Scan(
		&u.ID, &u.Account, &u.Bank, &u.BankNo, &u.FullName,
		&reg, &drawn, &u.Amount, &drawTime,
	)
	if err != nil {
		return nil, false
	}

	u.Registered = reg == 1
	u.HasDrawn = drawn == 1
	if drawTime != "" {
		if t, e := time.Parse(time.RFC3339, drawTime); e == nil {
			u.DrawTime = t
		}
	}

	cp := u
	return &cp, true
}

func (m *Store) Save(u *luckymoney.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if u == nil || u.ID == "" {
		return errors.New("invalid user id")
	}

	reg := 0
	if u.Registered {
		reg = 1
	}
	drawn := 0
	if u.HasDrawn {
		drawn = 1
	}
	drawTime := ""
	if !u.DrawTime.IsZero() {
		drawTime = u.DrawTime.Format(time.RFC3339)
	}

	_, err := m.db.Exec(`
		INSERT INTO users(id, account, bank, bank_no, full_name, registered, has_drawn, amount, draw_time)
		VALUES(?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			account=excluded.account,
			bank=excluded.bank,
			bank_no=excluded.bank_no,
			full_name=excluded.full_name,
			registered=excluded.registered,
			has_drawn=excluded.has_drawn,
			amount=excluded.amount,
			draw_time=excluded.draw_time
	`,
		u.ID, u.Account, u.Bank, u.BankNo, u.FullName,
		reg, drawn, u.Amount, drawTime,
	)
	if err != nil {
		log.Println("[ERR][User.Save]", err)
	}
	return err
}

func (m *Store) SetPool(items []luckymoney.PoolItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, err := m.db.Begin()
	if err != nil {
		log.Println("[ERR][SetPool.begin]", err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM pool_counts`); err != nil {
		log.Println("[ERR][SetPool.delete]", err)
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO pool_counts(amount, qty) VALUES(?, ?)`)
	if err != nil {
		log.Println("[ERR][SetPool.prepare]", err)
		return err
	}
	defer stmt.Close()

	for _, it := range items {
		if it.Amount <= 0 || it.Qty < 0 {
			continue
		}
		if _, err := stmt.Exec(it.Amount, it.Qty); err != nil {
			log.Println("[ERR][SetPool.insert]", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("[ERR][SetPool.commit]", err)
		return err
	}
	return nil
}

func (m *Store) GetPoolCounts() []luckymoney.PoolItem {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows, err := m.db.Query(`SELECT amount, qty FROM pool_counts ORDER BY amount`)
	if err != nil {
		log.Println("[ERR][GetPoolCounts.query]", err)
		return nil
	}
	defer rows.Close()

	out := []luckymoney.PoolItem{}
	for rows.Next() {
		var it luckymoney.PoolItem
		if err := rows.Scan(&it.Amount, &it.Qty); err != nil {
			log.Println("[ERR][GetPoolCounts.scan]", err)
			continue
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		log.Println("[ERR][GetPoolCounts.rows]", err)
	}
	return out
}

func (m *Store) remainingNoLock() int {
	rows, err := m.db.Query(`SELECT qty FROM pool_counts`)
	if err != nil {
		log.Println("[ERR][Remaining.query]", err)
		return 0
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var q int
		if err := rows.Scan(&q); err != nil {
			log.Println("[ERR][Remaining.scan]", err)
			continue
		}
		total += q
	}
	if err := rows.Err(); err != nil {
		log.Println("[ERR][Remaining.rows]", err)
	}
	return total
}

func (m *Store) Remaining() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.remainingNoLock()
}

func (m *Store) DrawOne() (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Println("[DRAW] remaining =", m.remainingNoLock())

	tx, err := m.db.Begin()
	if err != nil {
		log.Println("[ERR][DrawOne.begin]", err)
		return 0, false
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.Query(`SELECT amount, qty FROM pool_counts WHERE qty > 0 ORDER BY amount`)
	if err != nil {
		log.Println("[ERR][DrawOne.select]", err)
		return 0, false
	}
	defer rows.Close()

	type row struct {
		amount int
		qty    int
	}
	var list []row
	total := 0
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.amount, &r.qty); err != nil {
			log.Println("[ERR][DrawOne.scan]", err)
			continue
		}
		list = append(list, r)
		total += r.qty
	}
	if err := rows.Err(); err != nil {
		log.Println("[ERR][DrawOne.rows]", err)
		return 0, false
	}
	if total <= 0 {
		return 0, false
	}

	r := mrand.Intn(total)
	chosen := 0
	for _, it := range list {
		if r < it.qty {
			chosen = it.amount
			break
		}
		r -= it.qty
	}
	if chosen == 0 {
		return 0, false
	}

	res, err := tx.Exec(`UPDATE pool_counts SET qty = qty - 1 WHERE amount = ? AND qty > 0`, chosen)
	if err != nil {
		log.Println("[ERR][DrawOne.update]", err)
		return 0, false
	}
	aff, err := res.RowsAffected()
	if err != nil {
		log.Println("[ERR][DrawOne.affected]", err)
		return 0, false
	}
	if aff == 0 {
		return 0, false
	}

	if err := tx.Commit(); err != nil {
		log.Println("[ERR][DrawOne.commit]", err)
		return 0, false
	}
	return chosen, true
}

func (m *Store) AppendClaim(c luckymoney.Claim) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t := c.Time
	if t.IsZero() {
		t = time.Now()
	}

	_, err := m.db.Exec(`
		INSERT INTO claims(user_id, account, bank, bank_no, full_name, amount, time)
		VALUES(?,?,?,?,?,?,?)
	`,
		c.ID, c.Account, c.Bank, c.BankNo, c.FullName, c.Amount, t.Format(time.RFC3339),
	)
	if err != nil {
		log.Println("[ERR][AppendClaim]", err)
	}
	return err
}

func (m *Store) ListClaims() []luckymoney.Claim {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows, err := m.db.Query(`
		SELECT user_id, account, bank, bank_no, full_name, amount, time
		FROM claims
		ORDER BY time DESC
	`)
	if err != nil {
		log.Println("[ERR][ListClaims.query]", err)
		return nil
	}
	defer rows.Close()

	out := []luckymoney.Claim{}
	for rows.Next() {
		var c luckymoney.Claim
		var ts string
		if err := rows.Scan(&c.ID, &c.Account, &c.Bank, &c.BankNo, &c.FullName, &c.Amount, &ts); err != nil {
			log.Println("[ERR][ListClaims.scan]", err)
			continue
		}
		if t, e := time.Parse(time.RFC3339, ts); e == nil {
			c.Time = t
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		log.Println("[ERR][ListClaims.rows]", err)
	}
	return out
}
