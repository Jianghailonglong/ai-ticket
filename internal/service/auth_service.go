package service

import (
	"fmt"

	"github.com/ai-ticket/ai-ticket/internal/auth"
	"github.com/ai-ticket/ai-ticket/internal/database"
	"github.com/ai-ticket/ai-ticket/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register 注册
func (s *AuthService) Register(req model.LoginRequest, displayName string) error {
	// 检查用户是否已存在
	exists, err := database.UserExists(req.Username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("username already exists")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		DisplayName:  displayName,
	}

	return database.CreateUser(user)
}

// Login 登录
func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	// 查询用户
	user, err := database.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 生成Token
	token, err := auth.GenerateToken(user.Username)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:       token,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	}, nil
}
