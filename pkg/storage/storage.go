package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"github.com/yourusername/fmcl/pkg/parser"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS financial_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			previous_value TEXT,
			forecast_value TEXT,
			actual_value TEXT,
			timestamp DATETIME
		)
	`); err != nil {
		return nil, err
	}

	return &DB{Conn: conn}, nil
}

func (db *DB) Save(data *parser.FinancialData) error {
	_, err := db.Conn.Exec(
		"INSERT INTO financial_data (previous_value, forecast_value, actual_value, timestamp) VALUES (?, ?, ?, ?)",
		data.PreviousValue, data.ForecastValue, data.ActualValue, data.Timestamp,
	)
	return err
}
