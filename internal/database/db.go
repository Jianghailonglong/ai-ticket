package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Init 初始化数据库
func Init(dbPath string) error {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 测试连接
	if err := DB.Ping(); err != nil {
		return err
	}

	// 创建表
	if err := createTables(); err != nil {
		return err
	}

	log.Println("数据库初始化完成")
	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			display_name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tickets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticket_id TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			scene TEXT NOT NULL,
			applicant TEXT NOT NULL,
			approver TEXT NOT NULL,
			reason TEXT,
			status TEXT DEFAULT 'pending',
			comment TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭数据库
func Close() {
	if DB != nil {
		DB.Close()
	}
}
