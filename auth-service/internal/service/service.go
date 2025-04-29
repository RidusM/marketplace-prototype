package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/auth"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/cache"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/email"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/hash"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/logger"
	"go.opentelemetry.io/otel/trace"
)

type (
	UserRepository interface {
		CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
		Verify(ctx context.Context, userID uuid.UUID) error
		GetByEmail(ctx context.Context, email string) (*entity.User, error)
		GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
		UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error)
		UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPass string) error
		DeleteUser(ctx context.Context, userID uuid.UUID) error
	}

	CacheRepository interface {
		Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
		Get(ctx context.Context, key string) ([]byte, error)
		Delete(ctx context.Context, key string) error
		DeleteByPattern(ctx context.Context, pattern string) error
	}

	PassHistoryRepository interface {
		GetPasswordHistory(ctx context.Context, userID uuid.UUID) ([]string, error)
		SavePasswordToHistory(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	}

	AuthService struct {
		userRepo        UserRepository
		cacheRepo       CacheRepository
		passHistoryRepo PassHistoryRepository

		tokenManager auth.TokenManager
		mailer       email.Mailer

		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		userCacheTTL    time.Duration
		verifyTTL       time.Duration

		log       *logger.Logger
		waitGroup *sync.WaitGroup
	}
)

func NewAuthService(userRepo UserRepository,
	cacheRepo CacheRepository,
	passHistoryRepo PassHistoryRepository,
	tokenManager auth.TokenManager,
	mailer email.Mailer,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	userCacheTTL time.Duration,
	verifyTTL time.Duration,
	log *logger.Logger,
	waitGroup *sync.WaitGroup,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		cacheRepo:       cacheRepo,
		passHistoryRepo: passHistoryRepo,
		mailer:          mailer,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		userCacheTTL:    userCacheTTL,
		verifyTTL:       verifyTTL,
		log:             log,
		waitGroup:       waitGroup,
	}
}

func (s *AuthService) SignUp(ctx context.Context, input entity.UserSignUpInput) (uuid.UUID, entity.Tokens, error) {
	const op = "authService.SginUp"

	deviceID := uuid.New()

	span := trace.SpanFromContext(ctx)
	span.AddEvent("registering user")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("email", input.Email),
	)

	log.Info("registering user")

	hashedPassword, err := hash.Password(input.Password)
	if err != nil {
		s.log.Error("failed to generate password hash", logger.Err(err))

		return uuid.Nil, entity.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	user := &entity.User{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
		Role:     entity.Client,
	}

	user, err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		s.log.Error("failed to create user", logger.Err(err))

		return uuid.Nil, entity.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := s.setSession(ctx, user.Role, user.ID, deviceID)
	if err != nil {
		s.log.Error("failed to set session", logger.Err(err))

		return uuid.Nil, entity.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	cacheKey := cache.GenerateCacheKey("user:", user.ID)

	userSerialized, err := cache.Serialize(user)
	if err != nil {
		s.log.Error("failed to serialize user cache", logger.Err(err))

		return uuid.Nil, tokens, fmt.Errorf("%s: %w", op, err)
	}

	err = s.cacheRepo.Set(ctx, cacheKey, userSerialized, s.userCacheTTL)
	if err != nil {
		s.log.Warn("failed to set user cache", logger.Err(err))

		return uuid.Nil, tokens, fmt.Errorf("%s: %w", op, err)
	}

	verifyToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		s.log.Error("Failed to generate verify token: %w", logger.Err(err))
		return uuid.Nil, tokens, fmt.Errorf("%s: %w", op, err)
	}

	verifyTokenSerialized, err := cache.Serialize(verifyToken)
	if err != nil {
		s.log.Error("failed to serialize verify token cache", logger.Err(err))

		return uuid.Nil, tokens, fmt.Errorf("%s: %w", op, err)
	}

	verifyCacheKey := cache.GenerateCacheKey("verify:", user.ID)
	err = s.cacheRepo.Set(ctx, verifyCacheKey, verifyTokenSerialized, s.verifyTTL)
	if err != nil {
		s.log.Error("Failed to save verify token in cache", logger.Err(err))
		return uuid.Nil, tokens, fmt.Errorf("%s: %w", op, err)
	}

	emailData := map[string]interface{}{
		"token":      verifyToken,
		"userID":     user.ID,
		"expiration": s.verifyTTL,
	}

	err = s.mailer.Send(user.Email, "user_welcome.tmpl", emailData)
	if err != nil {
		s.log.Error("Failed to send email to user email", logger.Err(err))
		return uuid.Nil, tokens, fmt.Errorf("%s: %w", op, err)
	}

	return user.ID, tokens, nil
}

func (s *AuthService) Verify(ctx context.Context, userID uuid.UUID, verifyToken []byte) error {
	const op = "authService.Verify"

	span := trace.SpanFromContext(ctx)
	span.AddEvent("verify user")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("user_id", userID.String()),
	)
	log.Info("verify user")

	cacheKey := cache.GenerateCacheKey("verify:", userID)

	expectedToken, err := s.cacheRepo.Get(ctx, cacheKey)
	if err != nil {
		s.log.Error("failed to get expected cache token", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if !bytes.Equal(expectedToken, verifyToken) {
		s.log.Info("verification token does not match expected")

		return nil
	}

	err = s.userRepo.Verify(ctx, userID)
	if err != nil {
		s.log.Error("failed to verify user", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *AuthService) SignIn(ctx context.Context, input entity.UserSignInInput) (entity.Tokens, error) {
	const op = "authService.SignIn"

	span := trace.SpanFromContext(ctx)
	span.AddEvent("attempting to login user")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("email", input.Email),
	)

	log.Info("attempting to login user")

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, entity.ErrDataNotFound) {
			s.log.Warn("user not found", logger.Err(err))

			return entity.Tokens{}, fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", logger.Err(err))

		return entity.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	err = hash.ComparePassword(input.Password, user.Password)
	if err != nil {
		s.log.Info("invalid credentials", logger.Err(err))

		return entity.Tokens{}, fmt.Errorf("%s: %w", op, entity.ErrInvalidCredentials)
	}

	deviceID := uuid.New()

	return s.setSession(ctx, user.Role, user.ID, deviceID)
}

func (s *AuthService) setSession(ctx context.Context, userRole entity.UserRole, userID, deviceID uuid.UUID) (entity.Tokens, error) {
	const op = "authService.setSession"

	var (
		res entity.Tokens
		err error
	)

	span := trace.SpanFromContext(ctx)
	span.AddEvent("set user session")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("user_id", userID.String()),
		slog.String("device_id", deviceID.String()),
	)

	log.Info("set user session")

	cacheKey := cache.GenerateCacheKey("session:"+deviceID.String(), userID)

	res.AccessToken, err = s.tokenManager.NewJWT(userID, deviceID, userRole, s.accessTokenTTL)
	if err != nil {
		s.log.Error("failed to generate access token", logger.Err(err))

		return res, fmt.Errorf("%s: %w", op, err)
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		s.log.Error("failed to generate refresh token", logger.Err(err))

		return res, fmt.Errorf("%s: %w", op, err)
	}

	session := entity.Session{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}

	sessionSerialized, err := cache.Serialize(session)
	if err != nil {
		s.log.Error("failed to serialize user cache", logger.Err(err))

		return res, fmt.Errorf("%s: %w", op, err)
	}

	err = s.cacheRepo.Set(ctx, cacheKey, sessionSerialized, time.Duration(session.ExpiresAt.Unix()))
	if err != nil {
		s.log.Error("failed to set session", logger.Err(err))

		return res, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (s *AuthService) RefreshSession(ctx context.Context, userRole entity.UserRole, userID, deviceID uuid.UUID) (entity.Tokens, error) {
	const op = "authService.RefreshSession"

	span := trace.SpanFromContext(ctx)
	span.AddEvent("refresh user session")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("user_id", userID.String()),
		slog.String("device_id", deviceID.String()),
	)

	log.Info("refresh user session")

	cacheKey := cache.GenerateCacheKey("session:"+deviceID.String(), userID)

	_, err := s.cacheRepo.Get(ctx, cacheKey)
	if err != nil {
		s.log.Error("failed to get refresh token", logger.Err(err))

		return entity.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	session, err := s.setSession(ctx, userRole, userID, deviceID)
	if err != nil {
		s.log.Error("failed to set session", logger.Err(err))

		return entity.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	return session, nil
}

func (s *AuthService) LogOut(ctx context.Context, tokens entity.Tokens) error {
	const op = "authService.LogOut"

	var data *auth.ParsedTokenData

	span := trace.SpanFromContext(ctx)
	span.AddEvent("user log out")

	data, err := s.tokenManager.Parse(tokens.AccessToken)
	if err != nil {
		s.log.Error("failed to parse user data", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("userID", data.UserID.String()),
		slog.String("deviceID", data.DeviceID.String()),
	)

	log.Info("user log out")

	sessionCacheKey := cache.GenerateCacheKey("session:"+data.DeviceID.String(), data.UserID)

	err = s.cacheRepo.Delete(ctx, sessionCacheKey)
	if err != nil {
		s.log.Info("failed to delete user session", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, input entity.UserChangePasswordInput) error {
	const op = "authService.ChangePassword"

	var existingUser *entity.User

	span := trace.SpanFromContext(ctx)
	span.AddEvent("change user password")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("email", input.Email),
	)

	log.Info("change user password")

	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, entity.ErrDataNotFound) {
			s.log.Warn("user not found", logger.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := hash.ComparePassword(input.NewPassword, existingUser.Password); err != nil {
		s.log.Info("new password must not be null and the same as the previous one")

		return fmt.Errorf("%s: %w", op, err)
	}

	oldPasswords, err := s.passHistoryRepo.GetPasswordHistory(ctx, existingUser.ID)
	if err != nil {
		s.log.Error("failed to get password history")

		return fmt.Errorf("%s: %w", op, err)
	}

	for _, oldPass := range oldPasswords {
		if err := hash.ComparePassword(input.NewPassword, oldPass); err != nil {
			s.log.Info("ew password must not be null and the same as the previous one")

			return fmt.Errorf("%s: %w", op, err)
		}
	}

	s.waitGroup.Add(1)
	go func() {
		defer s.waitGroup.Done()
		defer func() {
			if p := recover(); p != nil {
				s.log.Warn("failed to sending email")
			}
		}()

		verifyToken, err := s.tokenManager.NewRefreshToken()
		if err != nil {
			s.log.Error("Failed to generate verify token", logger.Err(err))
			return
		}

		verifyCacheKey := cache.GenerateCacheKey("verify:", existingUser.ID)
		err = s.cacheRepo.Set(ctx, verifyCacheKey, []byte(verifyToken), s.verifyTTL)
		if err != nil {
			s.log.Error("Failed to save verify token in cache", logger.Err(err))
			return
		}

		data := map[string]interface{}{
			"name":       existingUser.Username,
			"token":      verifyToken,
			"userID":     existingUser.ID,
			"expiration": s.verifyTTL,
		}

		err = s.mailer.Send(input.Email, "password_reset.tmpl", data)
		if err != nil {
			s.log.Error("Error sending email", logger.Err(err))
		}

		s.log.Info("Email sent to %s")
	}()

	return nil
}

func (s *AuthService) ConfirmChangePassword(ctx context.Context, userID uuid.UUID, verifyToken []byte, newPassword string) error {
	const op = "authService.ConfirmChangePassword"

	span := trace.SpanFromContext(ctx)
	span.AddEvent("confirm user password change")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("user_id", userID.String()),
	)

	log.Info("confirm user password change")

	cacheKey := cache.GenerateCacheKey("verify:", userID)

	expectedToken, err := s.cacheRepo.Get(ctx, cacheKey)
	if err != nil {
		s.log.Error("failed to get expected cache token", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if !bytes.Equal(expectedToken, verifyToken) {
		s.log.Info("verification token does not match expected")

		return nil
	}

	hashedPassword, err := hash.Password(newPassword)
	if err != nil {
		s.log.Error("failed to hash new password", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.userRepo.UpdateUserPassword(ctx, userID, hashedPassword)
	if err != nil {
		s.log.Error("failed to update user password", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	if err = s.passHistoryRepo.SavePasswordToHistory(ctx, userID, hashedPassword); err != nil {
		s.log.Error("failed to save password to history", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	sessionCacheKey := cache.GenerateCacheKey("session:*", userID)

	err = s.cacheRepo.DeleteByPattern(ctx, sessionCacheKey)
	if err != nil {
		s.log.Warn("failed to delete all user sessions", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	const op = "authService.GetUser"

	var user *entity.User

	span := trace.SpanFromContext(ctx)
	span.AddEvent("get user")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("user_id", userID.String()),
	)

	log.Info("get user")

	cacheKey := cache.GenerateCacheKey("user:", userID)

	cachedUser, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil {
		err = cache.Deserialize(cachedUser, &user)
		if err != nil {
			s.log.Warn("failed to deserialize user cache", logger.Err(err))

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return user, nil
	}

	user, err = s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrDataNotFound) {
			s.log.Warn("user not found", logger.Err(err))

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", logger.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userSerialized, err := cache.Serialize(user)
	if err != nil {
		s.log.Warn("failed to serialize user cache", logger.Err(err))

		return user, fmt.Errorf("%s: %w", op, err)
	}

	err = s.cacheRepo.Set(ctx, cacheKey, userSerialized, 0)
	if err != nil {
		s.log.Warn("failed to set user cache", logger.Err(err))

		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	const op = "authService.UpdateUser"

	var hashedPassword string

	span := trace.SpanFromContext(ctx)
	span.AddEvent("update user")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("user_id", user.ID.String()),
	)

	log.Info("update user")

	existingUser, err := s.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, entity.ErrDataNotFound) {
			s.log.Warn("user not found", logger.Err(err))

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", logger.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	emptyData := user.Username == "" &&
		user.Email == "" &&
		user.Password == "" &&
		user.Role == ""

	sameData := existingUser.Username == user.Username &&
		existingUser.Email == user.Email &&
		existingUser.Role == user.Role

	if emptyData || sameData {
		s.log.Warn("empty data or same data", logger.Err(err))

		return nil, fmt.Errorf("%s: %w", op, entity.ErrNoUpdatedData)
	}

	if user.Password == "" {
		s.log.Error("empty password", logger.Err(err))

		return nil, fmt.Errorf("%s: %w", op, entity.ErrEmptyPassword)
	}

	hashedPassword, err = hash.Password(user.Password)
	if err != nil {
		s.log.Error("failed to hash password", logger.Err(err))

		return nil, fmt.Errorf("%s: %w", op, entity.ErrEmptyPassword)
	}

	user.Password = hashedPassword

	user, err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, entity.ErrConflictingData) {
			s.log.Warn("new data must not be null", logger.Err(err))

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to update user", logger.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userCacheKeyPattern := cache.GenerateCacheKey("user:*", user.ID)

	err = s.cacheRepo.DeleteByPattern(ctx, userCacheKeyPattern)
	if err != nil {
		s.log.Info("failed to delete user cache", logger.Err(err))

		return user, fmt.Errorf("%s: %w", op, err)
	}

	cacheKey := cache.GenerateCacheKey("user:", user.ID)

	userSerialized, err := cache.Serialize(user)
	if err != nil {
		s.log.Warn("failed to serialize user cache", logger.Err(err))

		return user, fmt.Errorf("%s: %w", op, err)
	}

	err = s.cacheRepo.Set(ctx, cacheKey, userSerialized, s.userCacheTTL)
	if err != nil {
		s.log.Warn("failed to set user cache", logger.Err(err))

		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "authService.DeleteUser"

	span := trace.SpanFromContext(ctx)
	span.AddEvent("delete user")

	log := s.log.With(
		slog.String("op", op),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("user_id", userID.String()),
	)

	log.Info("delete user")

	sessionCacheKeyPattern := cache.GenerateCacheKey("session:*", userID)

	userCacheKeyPattern := cache.GenerateCacheKey("user:*", userID)

	_ = s.userRepo.DeleteUser(ctx, userID)

	err := s.cacheRepo.DeleteByPattern(ctx, sessionCacheKeyPattern)
	if err != nil {
		s.log.Error("failed to delete session cache", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.cacheRepo.Delete(ctx, userCacheKeyPattern)
	if err != nil {
		s.log.Error("failed to delete user cache", logger.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
