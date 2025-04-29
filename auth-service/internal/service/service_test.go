package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/config"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/entity"
	mock_repository "gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/repository/mock"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/service"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/auth"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/cache"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/email"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/hash"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/logger"
	"go.uber.org/mock/gomock"
)

type signUpTestedInput struct {
	user entity.UserSignUpInput
}

type signUpExpectedOutput struct {
	tokens *entity.Tokens
	err    error
}

func TestAuthService_SignUp(t *testing.T) {
	ctx := context.Background()
	userName := gofakeit.Name()
	userEmail := gofakeit.Email()
	userPassword := "super-secret-password"

	hashedPassword, _ := hash.Password(userPassword)

	userInput := &entity.UserSignUpInput{
		Username: userName,
		Email:    userEmail,
		Password: userPassword,
	}

	userOut := &entity.User{
		ID:        uuid.New(),
		Username:  userName,
		Email:     userEmail,
		Password:  hashedPassword,
		Role:      entity.Client,
		Verified:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tokensOutput := &signUpExpectedOutput{
		tokens: &entity.Tokens{
			AccessToken:  gofakeit.UUID(),
			RefreshToken: gofakeit.UUID(),
		},
	}

	testCases := []struct {
		desc  string
		mocks func(
			userRepo *mock_repository.MockUserRepository,
			cache *mock_repository.MockCacheRepository,
		)
		input    signUpTestedInput
		expected signUpExpectedOutput
	}{
		{
			desc: "Success",
			mocks: func(
				userRepo *mock_repository.MockUserRepository,
				cache *mock_repository.MockCacheRepository,
			) {
				userRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.AssignableToTypeOf(&entity.User{})).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						assert.Equal(t, userInput.Username, user.Username)
						assert.Equal(t, userInput.Email, user.Email)
						assert.Equal(t, entity.Client, user.Role)
						return userOut, nil
					})

				cache.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				cache.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

			},
			input: signUpTestedInput{
				user: *userInput,
			},
			expected: signUpExpectedOutput{
				tokens: tokensOutput.tokens,
				err:    nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создание моков
			userRepo := mock_repository.NewMockUserRepository(ctrl)
			cacheRepo := mock_repository.NewMockCacheRepository(ctrl)

			// Создание зависимостей
			authManager, _ := auth.New("super-secret-key")
			mailer := email.New(&config.EmailConfig{})
			logger := logger.New("local", nil)

			// Настройка моков
			tc.mocks(userRepo, cacheRepo)

			// Создание сервиса
			userService := service.NewAuthService(userRepo, cacheRepo, nil, authManager, mailer, 0, 0, 0, 0, logger, &sync.WaitGroup{})

			// Вызов тестируемого метода
			userid, tokens, err := userService.SignUp(ctx, tc.input.user)

			// Проверка результатов
			assert.Equal(t, tc.expected.err, err, "Error mismatch")

			// Проверяем, что токены не пустые
			assert.NotEmpty(t, userid, "userid should not be empty")
			assert.NotEmpty(t, tokens.AccessToken, "AccessToken should not be empty")
			assert.NotEmpty(t, tokens.RefreshToken, "RefreshToken should not be empty")
		})
	}
}

type verifyTestedInput struct {
	userID      uuid.UUID
	verifyToken []byte
}

type verifyExpectedOutput struct {
	err error
}

func TestAuthService_Verify(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	verifyToken := []byte("verify-token")

	cacheKey := cache.GenerateCacheKey("verify:", userID)

	verifyInput := &verifyTestedInput{
		userID:      userID,
		verifyToken: verifyToken,
	}

	verifyExpected := &verifyExpectedOutput{
		err: nil,
	}

	testCases := []struct {
		desc     string
		mocks    func(cache *mock_repository.MockCacheRepository, userRepo *mock_repository.MockUserRepository)
		input    verifyTestedInput
		expected verifyExpectedOutput
	}{
		{
			desc: "Success",
			mocks: func(cache *mock_repository.MockCacheRepository, userRepo *mock_repository.MockUserRepository) {
				cache.EXPECT().Get(gomock.Any(), cacheKey).Return(verifyToken, nil)
				userRepo.EXPECT().Verify(gomock.Any(), userID).Return(nil)
			},
			input:    *verifyInput,
			expected: *verifyExpected,
		},
		{
			desc: "Token mismatch",
			mocks: func(cache *mock_repository.MockCacheRepository, userRepo *mock_repository.MockUserRepository) {
				cache.EXPECT().Get(gomock.Any(), cacheKey).Return([]byte("wrong-token"), nil)
			},
			input:    *verifyInput,
			expected: *verifyExpected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cacheRepo := mock_repository.NewMockCacheRepository(ctrl)
			userRepo := mock_repository.NewMockUserRepository(ctrl)

			mailer := email.New(&config.EmailConfig{})

			tc.mocks(cacheRepo, userRepo)

			authService := service.NewAuthService(
				userRepo,
				cacheRepo,
				nil,
				nil,
				mailer,
				0, 0, 0, 0,
				logger.New("local", nil),
				&sync.WaitGroup{},
			)

			err := authService.Verify(ctx, tc.input.userID, tc.input.verifyToken)

			if tc.expected.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type signInExpectedOutput struct {
	tokens *entity.Tokens
	err    error
}

func TestAuthService_SignIn(t *testing.T) {
	ctx := context.Background()
	userEmail := gofakeit.Email()
	userPassword := "super-secret-password"
	hashedPassword, _ := hash.Password(userPassword)

	testCases := []struct {
		desc     string
		mocks    func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository)
		input    entity.UserSignInInput
		expected signInExpectedOutput
	}{
		{
			desc: "Success",
			mocks: func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository) {
				user := &entity.User{
					ID:       uuid.New(),
					Email:    userEmail,
					Password: hashedPassword,
					Role:     entity.Client,
				}
				userRepo.EXPECT().GetByEmail(gomock.Any(), userEmail).Return(user, nil)
				cacheRepo.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil) // Настроен вызов Set
			},
			input: entity.UserSignInInput{
				Email:    userEmail,
				Password: userPassword,
			},
			expected: signInExpectedOutput{
				tokens: &entity.Tokens{},
				err:    nil,
			},
		},
		{
			desc: "User not found",
			mocks: func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository) {
				userRepo.EXPECT().GetByEmail(gomock.Any(), "notfound@example.com").Return(nil, entity.ErrDataNotFound)
			},
			input: entity.UserSignInInput{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			expected: signInExpectedOutput{
				tokens: nil,
				err:    entity.ErrDataNotFound,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock_repository.NewMockUserRepository(ctrl)
			cacheRepo := mock_repository.NewMockCacheRepository(ctrl)

			manager, _ := auth.New("super-secret-key")

			tc.mocks(userRepo, cacheRepo)

			authService := service.NewAuthService(
				userRepo,
				cacheRepo,
				nil,
				manager,
				email.Mailer{},
				0, 0, 0, 0,
				logger.New("local", nil),
				&sync.WaitGroup{},
			)

			tokens, err := authService.SignIn(ctx, tc.input)

			if tc.expected.err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expected.err.Error()) // Проверка текста ошибки
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokens)
			}
		})
	}
}

func TestAuthService_GetUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	testCases := []struct {
		desc     string
		mocks    func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository)
		expected error
	}{
		{
			desc: "Success",
			mocks: func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository) {
				user := &entity.User{
					ID: userID,
				}
				cacheRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, entity.ErrDataNotFound)
				userRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
				cacheRepo.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil) // Настроен вызов Set
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock_repository.NewMockUserRepository(ctrl)
			cacheRepo := mock_repository.NewMockCacheRepository(ctrl)
			tc.mocks(userRepo, cacheRepo)

			authService := service.NewAuthService(
				userRepo,
				cacheRepo,
				nil,
				nil,
				email.Mailer{},
				0, 0, 0, 0,
				logger.New("local", nil),
				&sync.WaitGroup{},
			)

			_, err := authService.GetUser(ctx, userID)

			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAuthService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	testCases := []struct {
		desc     string
		mocks    func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository)
		input    *entity.User
		expected error
	}{
		{
			desc: "Success",
			mocks: func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository) {
				userRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(&entity.User{
					ID:       userID,
					Username: "oldUsername",
					Email:    "old@example.com",
					Role:     entity.Client,
				}, nil)
				userRepo.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(&entity.User{
					ID:       userID,
					Username: "newUsername",
					Email:    "new@example.com",
					Role:     entity.Client,
				}, nil)
				cacheRepo.EXPECT().DeleteByPattern(gomock.Any(), gomock.Any()).Return(nil)
				cacheRepo.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			input: &entity.User{
				ID:       userID,
				Username: "newUsername",
				Email:    "new@example.com",
				Password: "newPassword",
				Role:     entity.Client,
			},
			expected: nil,
		},
		{
			desc: "User not found",
			mocks: func(userRepo *mock_repository.MockUserRepository, cacheRepo *mock_repository.MockCacheRepository) {
				userRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, entity.ErrDataNotFound)
			},
			input: &entity.User{
				ID:       userID,
				Username: "newUsername",
				Email:    "new@example.com",
				Password: "newPassword",
				Role:     entity.Client,
			},
			expected: entity.ErrDataNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock_repository.NewMockUserRepository(ctrl)
			cacheRepo := mock_repository.NewMockCacheRepository(ctrl)
			tc.mocks(userRepo, cacheRepo)

			authService := service.NewAuthService(
				userRepo,
				cacheRepo,
				nil,
				nil,
				email.Mailer{},
				0, 0, 0, 0,
				logger.New("local", nil),
				&sync.WaitGroup{},
			)

			_, err := authService.UpdateUser(ctx, tc.input)

			if tc.expected != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expected.Error()) // Проверка текста ошибки
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
