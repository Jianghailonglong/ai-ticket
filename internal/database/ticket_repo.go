package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cloud-mcp/cloud-mcp/internal/model"
)

// ListTicketsByUser 查询用户相关的工单列表
// filter: "my_apply" 我发起的, "my_approve" 我审批的, "" 全部（只显示与我相关的）
func ListTicketsByUser(username, status string, page, pageSize int, filter string) ([]model.Ticket, int, error) {
	// 构建查询 - 只查询与当前用户相关的工单
	where := "WHERE (applicant = ? OR approver = ?)"
	args := []interface{}{username, username}

	// 按filter过滤
	if filter == "my_apply" {
		where = "WHERE applicant = ?"
		args = []interface{}{username}
	} else if filter == "my_approve" {
		where = "WHERE approver = ?"
		args = []interface{}{username}
	}

	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets %s", where)
	err := DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(
		`SELECT id, ticket_id, title, scene, applicant, approver, reason, status, comment, created_at, updated_at
		 FROM tickets %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		where,
	)
	args = append(args, pageSize, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tickets := []model.Ticket{}
	for rows.Next() {
		var t model.Ticket
		var createdAt, updatedAt sql.NullTime
		var comment sql.NullString
		var reason sql.NullString
		err := rows.Scan(
			&t.ID, &t.TicketID, &t.Title, &t.Scene,
			&t.Applicant, &t.Approver, &reason, &t.Status, &comment,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		t.Reason = reason.String
		t.Comment = comment.String
		if createdAt.Valid {
			t.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			t.UpdatedAt = updatedAt.Time
		}
		tickets = append(tickets, t)
	}

	return tickets, total, nil
}

// GetTicket 根据工单ID查询工单
func GetTicket(ticketID string) (*model.Ticket, error) {
	t := &model.Ticket{}
	var createdAt, updatedAt sql.NullTime
	var comment sql.NullString
	var reason sql.NullString

	err := DB.QueryRow(
		`SELECT id, ticket_id, title, scene, applicant, approver, reason, status, comment, created_at, updated_at
		 FROM tickets WHERE ticket_id = ?`,
		ticketID,
	).Scan(
		&t.ID, &t.TicketID, &t.Title, &t.Scene,
		&t.Applicant, &t.Approver, &reason, &t.Status, &comment,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.Reason = reason.String
	t.Comment = comment.String
	if createdAt.Valid {
		t.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		t.UpdatedAt = updatedAt.Time
	}
	return t, nil
}

// CreateTicket 创建工单
func CreateTicket(ticket *model.Ticket) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(
		`INSERT INTO tickets (ticket_id, title, scene, applicant, approver, reason, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ticket.TicketID, ticket.Title, ticket.Scene,
		ticket.Applicant, ticket.Approver, ticket.Reason, ticket.Status,
		now, now,
	)
	return err
}

// UpdateTicketStatus 更新工单状态
func UpdateTicketStatus(ticketID, status, comment string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := DB.Exec(
		`UPDATE tickets SET status = ?, comment = ?, updated_at = ? WHERE ticket_id = ?`,
		status, comment, now, ticketID,
	)
	return err
}

// GetNextSequence 获取当天的下一个工单序号（使用事务避免竞态）
func GetNextSequence(date string) (int, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 1, err
	}
	defer tx.Rollback()

	// 加锁查询当前最大序号
	var maxSeq int
	err = tx.QueryRow(
		`SELECT COALESCE(MAX(CAST(SUBSTR(ticket_id, -3) AS INTEGER)), 0)
		 FROM tickets WHERE ticket_id LIKE ?`,
		"T-"+date+"-%",
	).Scan(&maxSeq)
	if err != nil {
		return 1, err
	}

	if err := tx.Commit(); err != nil {
		return 1, err
	}
	return maxSeq + 1, nil
}

