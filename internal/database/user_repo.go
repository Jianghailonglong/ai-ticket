package database

import (
	"database/sql"

	"github.com/ai-ticket/ai-ticket/internal/model"
)

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	_, err := DB.Exec(
		`INSERT INTO users (username, password_hash, display_name) VALUES (?, ?, ?)`,
		user.Username, user.PasswordHash, user.DisplayName,
	)
	return err
}

// GetUserByUsername 根据用户名查询用户
func GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	var createdAt sql.NullTime
	err := DB.QueryRow(
		`SELECT id, username, password_hash, display_name, created_at FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName, &createdAt)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	return user, nil
}

// UserExists 检查用户是否存在
func UserExists(username string) (bool, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
