package db

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/model"
)

func openDB(inst config.Instance) (*sql.DB, error) {
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

func GetProcessList(inst config.Instance) ([]model.ProcessWithTransaction, string, error) {
	db, err := openDB(inst)
	if err != nil {
		return nil, "", fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

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

func KillProcess(inst config.Instance, id int64) error {
	db, err := openDB(inst)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("KILL ?", id)
	if err != nil {
		return fmt.Errorf("killing process %d: %w", id, err)
	}

	return nil
}
