package handler

import (
	"log/slog"
	"os"
	"testing"

	"github.com/ai-ticket/ai-ticket/internal/auth"
	"github.com/ai-ticket/ai-ticket/internal/database"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func setupTestDB(t *testing.T) func() {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	if err := database.Init(tmpFile); err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	return func() { database.Close() }
}

func generateTestToken(t *testing.T, username string) string {
	t.Helper()
	token, err := auth.GenerateToken(username)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

func TestHandleList_NoToken(t *testing.T) {
	tools := NewTicketTools("", newTestLogger())
	_, err := tools.handleList(map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestHandleList_InvalidToken(t *testing.T) {
	tools := NewTicketTools("invalid-token", newTestLogger())
	_, err := tools.handleList(map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestHandleDetail_NoToken(t *testing.T) {
	tools := NewTicketTools("", newTestLogger())
	_, err := tools.handleDetail(map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestHandleDetail_MissingTicketID(t *testing.T) {
	token := generateTestToken(t, "testuser")
	tools := NewTicketTools(token, newTestLogger())
	_, err := tools.handleDetail(map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing ticket_id")
	}
}

func TestHandleReview_NoToken(t *testing.T) {
	tools := NewTicketTools("", newTestLogger())
	_, err := tools.handleReview(map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestHandleReview_MissingTicketID(t *testing.T) {
	token := generateTestToken(t, "testuser")
	tools := NewTicketTools(token, newTestLogger())
	_, err := tools.handleReview(map[string]interface{}{
		"action": "agree",
	})
	if err == nil {
		t.Fatal("expected error for missing ticket_id")
	}
}

func TestHandleReview_InvalidAction(t *testing.T) {
	token := generateTestToken(t, "testuser")
	tools := NewTicketTools(token, newTestLogger())
	_, err := tools.handleReview(map[string]interface{}{
		"ticket_id": "T-001",
		"action":    "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestHandleReview_Disagree_NoComment(t *testing.T) {
	token := generateTestToken(t, "testuser")
	tools := NewTicketTools(token, newTestLogger())
	_, err := tools.handleReview(map[string]interface{}{
		"ticket_id": "T-001",
		"action":    "disagree",
	})
	if err == nil {
		t.Fatal("expected error for missing comment on disagree")
	}
}
