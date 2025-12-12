package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/repository"
	"smart-ledger-server/pkg/errcode"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository, cfg *config.Config) *UserService {
	return &UserService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	// 检查手机号是否已存在
	exists, err := s.userRepo.ExistsByPhone(ctx, req.Phone)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if exists {
		return nil, errcode.ErrUserExists
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.ErrServer
	}

	// 创建用户
	user := &model.User{
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
	}
	if user.Nickname == "" {
		user.Nickname = "用户" + req.Phone[7:] // 默认昵称
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errcode.ErrServer
	}

	// 生成Token
	return s.generateLoginResponse(user)
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 查找用户
	user, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrServer
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errcode.ErrPasswordWrong
	}

	// 更新最后登录时间
	now := time.Now()
	_ = s.userRepo.UpdateFields(ctx, user.ID, map[string]interface{}{
		"last_login_at": now,
	})
	user.LastLoginAt = &now

	// 生成Token
	return s.generateLoginResponse(user)
}

// GetProfile 获取用户信息
func (s *UserService) GetProfile(ctx context.Context, userID uint64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrServer
	}

	return s.toUserResponse(user), nil
}

// UpdateProfile 更新用户信息
func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrServer
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errcode.ErrServer
	}

	return s.toUserResponse(user), nil
}

// generateLoginResponse 生成登录响应
func (s *UserService) generateLoginResponse(user *model.User) (*dto.LoginResponse, error) {
	expiresAt := time.Now().Add(s.cfg.JWT.ExpireTime)

	token, err := s.generateToken(user.ID, expiresAt)
	if err != nil {
		return nil, errcode.ErrServer
	}

	return &dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *s.toUserResponse(user),
	}, nil
}

// generateToken 生成JWT Token
func (s *UserService) generateToken(userID uint64, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
		"iss":     s.cfg.JWT.Issuer,
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// toUserResponse 转换为用户响应
func (s *UserService) toUserResponse(user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:          user.ID,
		Phone:       user.Phone,
		Nickname:    user.Nickname,
		AvatarURL:   user.AvatarURL,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}
