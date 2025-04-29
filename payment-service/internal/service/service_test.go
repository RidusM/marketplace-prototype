package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ursulgwopp/payment-microservice/internal/entity"
	mock_repository "github.com/ursulgwopp/payment-microservice/internal/repository/mock"
	"go.uber.org/mock/gomock"
)

func TestProcessPayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Используем сгенерированные моки
	mockDB := mock_repository.NewMockDatabase(ctrl)
	mockCache := mock_repository.NewMockCache(ctrl)

	service := New(mockDB, mockCache)

	ctx := context.Background()
	payment := &entity.Payment{
		PaymentId: "payment123",
		OrderId:   "order456",
		Amount:    10000,
		Status:    "completed",
		CreatedAt: time.Now(),
	}

	// Успешный сценарий
	mockDB.EXPECT().ProcessPayment(ctx, *payment).Return(nil)

	success, err := service.ProcessPayment(ctx, payment)

	assert.NoError(t, err)
	assert.True(t, success)

	// Ошибка при обработке платежа
	mockDB.EXPECT().ProcessPayment(ctx, *payment).Return(errors.New("database error"))

	success, err = service.ProcessPayment(ctx, payment)

	assert.Error(t, err)
	assert.False(t, success)
	assert.Equal(t, "database error", err.Error())

	// Проверка с пустым платежом
	emptyPayment := &entity.Payment{}
	mockDB.EXPECT().ProcessPayment(ctx, *emptyPayment).Return(errors.New("invalid payment"))

	success, err = service.ProcessPayment(ctx, emptyPayment)

	assert.Error(t, err)
	assert.False(t, success)
	assert.Equal(t, "invalid payment", err.Error())
}
