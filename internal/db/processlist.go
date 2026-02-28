package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/go-sql-driver/mysql"

	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/model"
)

// OpenDB creates a connection pool for the given instance. The pool is meant to
// be long-lived; callers should create one pool per instance at startup and
// reuse it across requests.
func OpenDB(inst config.Instance) (*sql.DB, error) {
	cfg := mysql.Config{
		User:                 inst.User,
		Passwd:               inst.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", inst.Host, inst.Port),
		DBName:               inst.Database,
		AllowNativePasswords: true,
	}
	return sql.Open("mysql", cfg.FormatDSN())
}

func GetProcessList(db *sql.DB) ([]model.ProcessWithTransaction, string, error) {
	rows, err := db.Query("SHOW PROCESSLIST")
	if err != nil {
		return nil, "", fmt.Errorf("show processlist: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, "", fmt.Errorf("getting columns: %w", err)
	}
	hasProgress := len(cols) >= 9

	var processes []model.ProcessWithTransaction
	for rows.Next() {
		var p model.ProcessWithTransaction
		var dbName, info sql.NullString

		if hasProgress {
			err = rows.Scan(&p.ID, &p.User, &p.Host, &dbName, &p.Command, &p.Time, &p.State, &info, &p.Progress)
		} else {
			err = rows.Scan(&p.ID, &p.User, &p.Host, &dbName, &p.Command, &p.Time, &p.State, &info)
		}
		if err != nil {
			return nil, "", fmt.Errorf("scanning row: %w", err)
		}

		p.DB = dbName.String
		p.Info = info.String
		processes = append(processes, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterating rows: %w", err)
	}

	var typ, name, status string
	err = db.QueryRow("SHOW ENGINE INNODB STATUS").Scan(&typ, &name, &status)
	if err != nil {
		slog.Warn("failed to get innodb status, transaction data will be unavailable", "error", err)
		status = ""
	}

	txMap := parseInnoDBStatus(status)
	for i := range processes {
		if tx, ok := txMap[processes[i].ID]; ok {
			processes[i].Transaction = tx
		}
	}

	return processes, status, nil
}

func KillProcess(db *sql.DB, id int64) error {
	// Use fmt.Sprintf rather than a placeholder because MariaDB/MySQL may not
	// support parameter binding for KILL statements across all versions.
	// id is validated as int64 by the caller so there is no injection risk.
	_, err := db.Exec(fmt.Sprintf("KILL %d", id))
	if err != nil {
		return fmt.Errorf("killing process %d: %w", id, err)
	}

	return nil
}
