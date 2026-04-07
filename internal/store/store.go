package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Estimates struct {
	ID string `json:"id"`
	ClientName string `json:"client_name"`
	ClientEmail string `json:"client_email"`
	ClientPhone string `json:"client_phone"`
	Title string `json:"title"`
	Description string `json:"description"`
	Total float64 `json:"total"`
	ValidUntil string `json:"valid_until"`
	Status string `json:"status"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

type LineItems struct {
	ID string `json:"id"`
	EstimateId string `json:"estimate_id"`
	Description string `json:"description"`
	Quantity float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Total float64 `json:"total"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "estimate.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS estimates(id TEXT PRIMARY KEY, client_name TEXT NOT NULL, client_email TEXT DEFAULT '', client_phone TEXT DEFAULT '', title TEXT NOT NULL, description TEXT DEFAULT '', total REAL DEFAULT 0, valid_until TEXT DEFAULT '', status TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS line_items(id TEXT PRIMARY KEY, estimate_id TEXT NOT NULL, description TEXT NOT NULL, quantity REAL DEFAULT 0, unit_price REAL DEFAULT 0, total REAL DEFAULT 0, created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL, record_id TEXT NOT NULL, data TEXT NOT NULL DEFAULT '{}', PRIMARY KEY(resource, record_id))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateEstimates(e *Estimates) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO estimates(id, client_name, client_email, client_phone, title, description, total, valid_until, status, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.ClientName, e.ClientEmail, e.ClientPhone, e.Title, e.Description, e.Total, e.ValidUntil, e.Status, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetEstimates(id string) *Estimates {
	var e Estimates
	if d.db.QueryRow(`SELECT id, client_name, client_email, client_phone, title, description, total, valid_until, status, notes, created_at FROM estimates WHERE id=?`, id).Scan(&e.ID, &e.ClientName, &e.ClientEmail, &e.ClientPhone, &e.Title, &e.Description, &e.Total, &e.ValidUntil, &e.Status, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListEstimates() []Estimates {
	rows, _ := d.db.Query(`SELECT id, client_name, client_email, client_phone, title, description, total, valid_until, status, notes, created_at FROM estimates ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Estimates
	for rows.Next() { var e Estimates; rows.Scan(&e.ID, &e.ClientName, &e.ClientEmail, &e.ClientPhone, &e.Title, &e.Description, &e.Total, &e.ValidUntil, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateEstimates(e *Estimates) error {
	_, err := d.db.Exec(`UPDATE estimates SET client_name=?, client_email=?, client_phone=?, title=?, description=?, total=?, valid_until=?, status=?, notes=? WHERE id=?`, e.ClientName, e.ClientEmail, e.ClientPhone, e.Title, e.Description, e.Total, e.ValidUntil, e.Status, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteEstimates(id string) error {
	_, err := d.db.Exec(`DELETE FROM estimates WHERE id=?`, id)
	return err
}

func (d *DB) CountEstimates() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM estimates`).Scan(&n); return n
}

func (d *DB) SearchEstimates(q string, filters map[string]string) []Estimates {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (client_name LIKE ? OR client_email LIKE ? OR client_phone LIKE ? OR title LIKE ? OR description LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["status"]; ok && v != "" { where += " AND status=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, client_name, client_email, client_phone, title, description, total, valid_until, status, notes, created_at FROM estimates WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Estimates
	for rows.Next() { var e Estimates; rows.Scan(&e.ID, &e.ClientName, &e.ClientEmail, &e.ClientPhone, &e.Title, &e.Description, &e.Total, &e.ValidUntil, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateLineItems(e *LineItems) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO line_items(id, estimate_id, description, quantity, unit_price, total, created_at) VALUES(?, ?, ?, ?, ?, ?, ?)`, e.ID, e.EstimateId, e.Description, e.Quantity, e.UnitPrice, e.Total, e.CreatedAt)
	return err
}

func (d *DB) GetLineItems(id string) *LineItems {
	var e LineItems
	if d.db.QueryRow(`SELECT id, estimate_id, description, quantity, unit_price, total, created_at FROM line_items WHERE id=?`, id).Scan(&e.ID, &e.EstimateId, &e.Description, &e.Quantity, &e.UnitPrice, &e.Total, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListLineItems() []LineItems {
	rows, _ := d.db.Query(`SELECT id, estimate_id, description, quantity, unit_price, total, created_at FROM line_items ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []LineItems
	for rows.Next() { var e LineItems; rows.Scan(&e.ID, &e.EstimateId, &e.Description, &e.Quantity, &e.UnitPrice, &e.Total, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateLineItems(e *LineItems) error {
	_, err := d.db.Exec(`UPDATE line_items SET estimate_id=?, description=?, quantity=?, unit_price=?, total=? WHERE id=?`, e.EstimateId, e.Description, e.Quantity, e.UnitPrice, e.Total, e.ID)
	return err
}

func (d *DB) DeleteLineItems(id string) error {
	_, err := d.db.Exec(`DELETE FROM line_items WHERE id=?`, id)
	return err
}

func (d *DB) CountLineItems() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM line_items`).Scan(&n); return n
}

func (d *DB) SearchLineItems(q string, filters map[string]string) []LineItems {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (estimate_id LIKE ? OR description LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	rows, _ := d.db.Query(`SELECT id, estimate_id, description, quantity, unit_price, total, created_at FROM line_items WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []LineItems
	for rows.Next() { var e LineItems; rows.Scan(&e.ID, &e.EstimateId, &e.Description, &e.Quantity, &e.UnitPrice, &e.Total, &e.CreatedAt); o = append(o, e) }
	return o
}

// GetExtras returns the JSON extras blob for a record. Returns "{}" if none.
func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(`SELECT data FROM extras WHERE resource=? AND record_id=?`, resource, recordID).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

// SetExtras stores the JSON extras blob for a record.
func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?) ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`, resource, recordID, data)
	return err
}

// DeleteExtras removes extras when a record is deleted.
func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(`DELETE FROM extras WHERE resource=? AND record_id=?`, resource, recordID)
	return err
}

// AllExtras returns all extras for a resource type as a map of record_id → JSON string.
func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(`SELECT record_id, data FROM extras WHERE resource=?`, resource)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
