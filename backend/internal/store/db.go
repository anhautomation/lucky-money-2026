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
	drawSem chan struct{}

	db *sql.DB
}

var _ port.AdminSessionRepo = (*Store)(nil)
var _ port.IssuedIDRepo = (*Store)(nil)
var _ port.UserRepo = (*Store)(nil)
var _ port.PoolRepo = (*Store)(nil)
var _ port.ClaimRepo = (*Store)(nil)
var _ port.Locker = (*Store)(nil)

const (
	dbOpTimeout = 2 * time.Second

	drawOpTimeout = 4 * time.Second

	drawQueueTimeout = 5 * time.Second
)

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
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s := &Store{
		db:      db,
		drawSem: make(chan struct{}, 1),
	}

	s.drawSem <- struct{}{}

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
			draw_time TEXT NOT NULL DEFAULT ''
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
			time TEXT NOT NULL
		);`,
	}

	for _, st := range stmts {
		if _, err := m.db.ExecContext(ctx, st); err != nil {
			return err
		}
	}
	return nil
}

func (m *Store) withTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

func (m *Store) WithLock(fn func()) {
	timer := time.NewTimer(drawQueueTimeout)
	defer timer.Stop()

	select {
	case <-m.drawSem:
	case <-timer.C:
		log.Println("[DRAW][QUEUE_TIMEOUT]")
		panic(luckymoney.ErrDrawTimeout)
	}

	defer func() {
		m.drawSem <- struct{}{}
	}()

	defer func() {
		if r := recover(); r != nil {
			log.Println("[LOCK][PANIC]", r)
			panic(r)
		}
	}()

	fn()
}

func (m *Store) SetSession(token string) error {
	if token == "" {
		return errors.New("empty token")
	}
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	_, err := m.db.ExecContext(ctx, `
		INSERT INTO admin_session(id, token)
		VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET token=excluded.token
	`, token)
	return err
}

func (m *Store) IsValidSession(token string) bool {
	if token == "" {
		return false
	}
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	var saved string
	if err := m.db.QueryRowContext(ctx, `SELECT token FROM admin_session WHERE id=1`).Scan(&saved); err != nil {
		return false
	}
	return token == saved
}

func (m *Store) ClearSession() error {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	_, err := m.db.ExecContext(ctx, `DELETE FROM admin_session WHERE id=1`)
	return err
}

func (m *Store) IssueIDs(ids []string) error {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO issued_ids(id) VALUES (?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (m *Store) IsIssued(id string) bool {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	var x string
	err := m.db.QueryRowContext(ctx, `SELECT id FROM issued_ids WHERE id=?`, id).Scan(&x)
	return err == nil
}

func (m *Store) ListIssued() []string {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	rows, err := m.db.QueryContext(ctx, `SELECT id FROM issued_ids ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		_ = rows.Scan(&id)
		out = append(out, id)
	}
	return out
}

func (m *Store) DeleteIssued(id string) error {
	if id == "" {
		return nil
	}
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var drawn int
	err = tx.QueryRowContext(ctx, `SELECT has_drawn FROM users WHERE id=?`, id).Scan(&drawn)
	if err == nil && drawn == 1 {
		return luckymoney.ErrIDAlreadyDrawn
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM issued_ids WHERE id=?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (m *Store) GetByID(id string) (*luckymoney.User, bool) {
	if id == "" {
		return nil, false
	}
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	var u luckymoney.User
	var reg, drawn int
	var drawTime string

	err := m.db.QueryRowContext(ctx, `
		SELECT id, account, bank, bank_no, full_name, registered, has_drawn, amount, draw_time
		FROM users WHERE id=?`,
		id,
	).Scan(
		&u.ID, &u.Account, &u.Bank, &u.BankNo, &u.FullName,
		&reg, &drawn, &u.Amount, &drawTime,
	)
	if err != nil {
		return nil, false
	}

	u.Registered = reg == 1
	u.HasDrawn = drawn == 1
	if drawTime != "" {
		u.DrawTime, _ = time.Parse(time.RFC3339, drawTime)
	}
	return &u, true
}

func (m *Store) Save(u *luckymoney.User) error {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

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

	_, err := m.db.ExecContext(ctx, `
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
			draw_time=excluded.draw_time`,
		u.ID, u.Account, u.Bank, u.BankNo, u.FullName,
		reg, drawn, u.Amount, drawTime,
	)
	return err
}

func (m *Store) SetPool(items []luckymoney.PoolItem) error {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM pool_counts`); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO pool_counts(amount, qty) VALUES(?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, it := range items {
		if it.Amount > 0 && it.Qty > 0 {
			_, _ = stmt.ExecContext(ctx, it.Amount, it.Qty)
		}
	}
	return tx.Commit()
}

func (m *Store) GetPoolCounts() []luckymoney.PoolItem {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	rows, err := m.db.QueryContext(ctx, `SELECT amount, qty FROM pool_counts ORDER BY amount`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []luckymoney.PoolItem
	for rows.Next() {
		var it luckymoney.PoolItem
		_ = rows.Scan(&it.Amount, &it.Qty)
		out = append(out, it)
	}
	return out
}

func (m *Store) Remaining() int {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	rows, err := m.db.QueryContext(ctx, `SELECT qty FROM pool_counts`)
	if err != nil {
		return 0
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var q int
		_ = rows.Scan(&q)
		total += q
	}
	return total
}

func (m *Store) DrawOne() (int, bool) {
	ctx, cancel := m.withTimeout(drawOpTimeout)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT amount, qty FROM pool_counts WHERE qty > 0 ORDER BY amount`)
	if err != nil {
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
			return 0, false
		}
		list = append(list, r)
		total += r.qty
	}
	if total == 0 {
		return 0, false
	}

	rn := mrand.Intn(total)
	chosen := 0
	for _, it := range list {
		if rn < it.qty {
			chosen = it.amount
			break
		}
		rn -= it.qty
	}
	if chosen <= 0 {
		return 0, false
	}

	res, err := tx.ExecContext(ctx, `UPDATE pool_counts SET qty = qty - 1 WHERE amount = ? AND qty > 0`, chosen)
	if err != nil {
		return 0, false
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return 0, false
	}

	return chosen, tx.Commit() == nil
}

func (m *Store) RefundOne(amount int) error {
	if amount <= 0 {
		return nil
	}
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `UPDATE pool_counts SET qty = qty + 1 WHERE amount = ?`, amount)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO pool_counts(amount, qty) VALUES(?, 1)`, amount); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (m *Store) AppendClaim(c luckymoney.Claim) error {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	t := c.Time
	if t.IsZero() {
		t = time.Now()
	}

	_, err := m.db.ExecContext(ctx, `
		INSERT INTO claims(user_id, account, bank, bank_no, full_name, amount, time)
		VALUES(?,?,?,?,?,?,?)`,
		c.ID, c.Account, c.Bank, c.BankNo, c.FullName, c.Amount,
		t.Format(time.RFC3339),
	)
	return err
}

func (m *Store) ListClaims() []luckymoney.Claim {
	ctx, cancel := m.withTimeout(dbOpTimeout)
	defer cancel()

	rows, err := m.db.QueryContext(ctx, `
		SELECT user_id, account, bank, bank_no, full_name, amount, time
		FROM claims
		ORDER BY time DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []luckymoney.Claim
	for rows.Next() {
		var c luckymoney.Claim
		var ts string
		_ = rows.Scan(&c.ID, &c.Account, &c.Bank, &c.BankNo, &c.FullName, &c.Amount, &ts)
		c.Time, _ = time.Parse(time.RFC3339, ts)
		out = append(out, c)
	}
	return out
}

func (m *Store) DrawAndCommit(u luckymoney.User) (int, error) {
	if u.ID == "" {
		return 0, errors.New("empty user id")
	}

	t := u.DrawTime
	if t.IsZero() {
		t = time.Now()
	}

	ctx, cancel := m.withTimeout(drawOpTimeout)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT amount, qty FROM pool_counts WHERE qty > 0 ORDER BY amount`)
	if err != nil {
		return 0, err
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
			return 0, err
		}
		list = append(list, r)
		total += r.qty
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, luckymoney.ErrPoolEmpty
	}

	rn := mrand.Intn(total)
	chosen := 0
	for _, it := range list {
		if rn < it.qty {
			chosen = it.amount
			break
		}
		rn -= it.qty
	}
	if chosen <= 0 {
		return 0, luckymoney.ErrPoolEmpty
	}

	res, err := tx.ExecContext(ctx, `UPDATE pool_counts SET qty = qty - 1 WHERE amount = ? AND qty > 0`, chosen)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return 0, luckymoney.ErrPoolEmpty
	}

	reg := 0
	if u.Registered {
		reg = 1
	}
	drawn := 1
	drawTime := t.Format(time.RFC3339)

	_, err = tx.ExecContext(ctx, `
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
			draw_time=excluded.draw_time`,
		u.ID, u.Account, u.Bank, u.BankNo, u.FullName,
		reg, drawn, chosen, drawTime,
	)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO claims(user_id, account, bank, bank_no, full_name, amount, time)
		VALUES(?,?,?,?,?,?,?)`,
		u.ID, u.Account, u.Bank, u.BankNo, u.FullName, chosen, drawTime,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return chosen, nil
}
