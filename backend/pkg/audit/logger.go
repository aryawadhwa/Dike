package audit

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDB initializes the SQLite database for audit logging.
func InitDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home dir: %w", err)
	}

	pulseDir := filepath.Join(homeDir, ".pulse")
	if err := os.MkdirAll(pulseDir, 0755); err != nil {
		return fmt.Errorf("could not create .pulse dir: %w", err)
	}

	dbPath := filepath.Join(pulseDir, "audit.db")

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	db = database

	createTableSQL := `CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		username TEXT,
		command TEXT,
		risk_level TEXT,
		decision TEXT,
		preview_summary TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("could not create table: %w", err)
	}

	return nil
}

// LogDecision records a command attempt in the audit log.
func LogDecision(command string, riskLevel string, decision string, previewSummary string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	username := os.Getenv("USER")
	if username == "" {
		username = "unknown"
	}

	insertSQL := `INSERT INTO audit_log (username, command, risk_level, decision, preview_summary) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, username, command, riskLevel, decision, previewSummary)
	return err
}

// Close gracefully closes the database.
func Close() {
	if db != nil {
		db.Close()
	}
}

type AuditEntry struct {
	ID             int    `json:"id"`
	Timestamp      string `json:"timestamp"`
	Username       string `json:"username"`
	Command        string `json:"command"`
	RiskLevel      string `json:"risk_level"`
	Decision       string `json:"decision"`
	PreviewSummary string `json:"preview_summary"`
}

type Stats struct {
	Total    int `json:"total"`
	Applied  int `json:"applied"`
	Rejected int `json:"rejected"`
	Critical int `json:"critical"`
}

// GetAuditLogs returns the last 50 audit entries.
func GetAuditLogs() ([]AuditEntry, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query("SELECT id, timestamp, username, command, risk_level, decision, preview_summary FROM audit_log ORDER BY id DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Username, &e.Command, &e.RiskLevel, &e.Decision, &e.PreviewSummary); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// GetStats returns summary statistics.
func GetStats() (*Stats, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var stats Stats
	db.QueryRow("SELECT COUNT(*) FROM audit_log").Scan(&stats.Total)
	db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE decision = 'APPLIED'").Scan(&stats.Applied)
	db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE decision = 'REJECTED'").Scan(&stats.Rejected)
	db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE risk_level = 'CRITICAL'").Scan(&stats.Critical)

	return &stats, nil
}
