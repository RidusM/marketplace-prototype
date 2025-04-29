package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase реализует интерфейс Database для тестов
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Create(ctx context.Context, product *entity.Product) (uuid.UUID, error) {
	args := m.Called(ctx, product)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockDatabase) Get(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockDatabase) Update(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockDatabase) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatabase) List(ctx context.Context, offset, limit int64) ([]*entity.Product, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Product), args.Error(1)
}

// MockCache реализует интерфейс Cache для тестов
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Тесты
func TestService_Create(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("successful creation", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		req := &entity.CreateProductRequest{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999,
			Stock:       10,
		}

		dbMock.On("Create", ctx, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == req.Name &&
				p.Description == req.Description &&
				p.Price == req.Price &&
				p.Stock == req.Stock
		})).Return(uuid.New(), nil)

		cacheMock.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		id, err := svc.CreateProduct(ctx, req)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
		dbMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})

	t.Run("invalid name", func(t *testing.T) {
		svc := NewProductService(nil, nil)

		req := &entity.CreateProductRequest{
			Name:        "",
			Description: "High-performance laptop",
			Price:       999,
			Stock:       10,
		}

		_, err := svc.CreateProduct(ctx, req)

		assert.ErrorIs(t, err, entity.ErrInvalidName)
	})

	t.Run("invalid description", func(t *testing.T) {
		svc := NewProductService(nil, nil)

		req := &entity.CreateProductRequest{
			Name:        "Laptop",
			Description: "",
			Price:       999,
			Stock:       10,
		}

		_, err := svc.CreateProduct(ctx, req)

		assert.ErrorIs(t, err, entity.ErrInvalidDescription)
	})

	t.Run("invalid price", func(t *testing.T) {
		svc := NewProductService(nil, nil)

		req := &entity.CreateProductRequest{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       -999,
			Stock:       10,
		}

		_, err := svc.CreateProduct(ctx, req)

		assert.ErrorIs(t, err, entity.ErrInvalidPrice)
	})

	t.Run("invalid stock", func(t *testing.T) {
		svc := NewProductService(nil, nil)

		req := &entity.CreateProductRequest{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999,
			Stock:       -10,
		}

		_, err := svc.CreateProduct(ctx, req)

		assert.ErrorIs(t, err, entity.ErrInvalidStock)
	})

	t.Run("database error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		svc := NewProductService(dbMock, nil)

		req := &entity.CreateProductRequest{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999,
			Stock:       10,
		}

		dbMock.On("Create", ctx, mock.Anything).Return(uuid.Nil, errors.New("db error"))

		_, err := svc.CreateProduct(ctx, req)
		assert.ErrorContains(t, err, "db error")
	})

	t.Run("cache set error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)

		svc := NewProductService(dbMock, cacheMock)

		req := &entity.CreateProductRequest{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999,
			Stock:       10,
		}

		dbMock.On("Create", ctx, mock.Anything).Return(uuid.New(), nil)
		cacheMock.On("Set", ctx, mock.Anything, mock.Anything, duration).Return(errors.New("cache error"))

		_, err := svc.CreateProduct(ctx, req)
		assert.ErrorContains(t, err, "cache error")
	})

	testProduct := &entity.Product{
		Id:          testID,
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       999,
		Stock:       10,
	}

	productSerialized, err := utils.Serialize(testProduct)
	assert.NoError(t, err)

	t.Run("successful cache set", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		cacheKey := utils.GenerateCacheKey("product", testID)

		cacheMock.On("Set", ctx, cacheKey, productSerialized, duration).Return(nil)

		err = svc.cache.Set(ctx, cacheKey, productSerialized, duration)

		assert.NoError(t, err)
		cacheMock.AssertExpectations(t)
	})

	t.Run("cache set error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		cacheKey := utils.GenerateCacheKey("product", testID)

		cacheMock.On("Set", ctx, cacheKey, productSerialized, duration).Return(fmt.Errorf("cache set error"))

		err = svc.cache.Set(ctx, cacheKey, productSerialized, duration)

		assert.ErrorContains(t, err, "cache set error")
		cacheMock.AssertExpectations(t)
	})
}
func TestService_Get(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	testProduct := &entity.Product{
		Id:          testID,
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       999,
		Stock:       10,
	}

	productSerialized, err := utils.Serialize(testProduct)
	assert.NoError(t, err)

	t.Run("successful cache set", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		cacheKey := utils.GenerateCacheKey("product", testID)

		cacheMock.On("Set", ctx, cacheKey, productSerialized, duration).Return(nil)

		err = svc.cache.Set(ctx, cacheKey, productSerialized, duration)

		assert.NoError(t, err)
		cacheMock.AssertExpectations(t)
	})

	t.Run("cache set error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		cacheKey := utils.GenerateCacheKey("product", testID)

		cacheMock.On("Set", ctx, cacheKey, productSerialized, duration).Return(fmt.Errorf("cache set error"))

		err = svc.cache.Set(ctx, cacheKey, productSerialized, duration)

		assert.ErrorContains(t, err, "cache set error")
		cacheMock.AssertExpectations(t)
	})

	t.Run("successful get from cache", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		serialized, _ := utils.Serialize(testProduct)
		cacheMock.On("Get", ctx, "product:"+testID.String()).Return(serialized, nil)

		result, err := svc.GetProduct(ctx, testID)

		assert.NoError(t, err)
		assert.Equal(t, testProduct, result)
		dbMock.AssertNotCalled(t, "Get")
		cacheMock.AssertExpectations(t)
	})

	t.Run("cache miss, get from db", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		cacheMock.On("Get", ctx, "product:"+testID.String()).Return(nil, errors.New("not found"))
		dbMock.On("Get", ctx, testID).Return(testProduct, nil)

		serialized, _ := utils.Serialize(testProduct)
		cacheMock.On("Set", ctx, "product:"+testID.String(), serialized, duration).Return(nil)

		result, err := svc.GetProduct(ctx, testID)

		assert.NoError(t, err)
		assert.Equal(t, testProduct, result)
		dbMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})

	t.Run("db not found", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		cacheMock.On("Get", ctx, "product:"+testID.String()).Return(nil, errors.New("not found"))
		dbMock.On("Get", ctx, testID).Return(nil, sql.ErrNoRows)

		_, err := svc.GetProduct(ctx, testID)

		assert.ErrorIs(t, err, sql.ErrNoRows)
		cacheMock.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("db error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		cacheMock.On("Get", ctx, "product:"+testID.String()).Return(nil, errors.New("not found"))
		dbMock.On("Get", ctx, testID).Return(nil, errors.New("db error"))

		_, err := svc.GetProduct(ctx, testID)

		assert.ErrorContains(t, err, "db error")
	})
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()
	testProduct := &entity.Product{
		Id:          testID,
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       999,
		Stock:       10,
	}

	t.Run("invalid name", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		req := &entity.UpdateProductRequest{
			Id:          testID,
			Name:        "", // Invalid name
			Description: "Valid description",
			Price:       100,
			Stock:       10,
		}

		// Call the Update method
		_, err := svc.UpdateProduct(ctx, req)

		// Assertions
		assert.ErrorIs(t, err, entity.ErrInvalidName)
	})

	t.Run("invalid description", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		req := &entity.UpdateProductRequest{
			Id:          testID,
			Name:        "Valid Name",
			Description: "",
			Price:       100,
			Stock:       10,
		}

		_, err := svc.UpdateProduct(ctx, req)

		assert.ErrorIs(t, err, entity.ErrInvalidDescription)
	})

	t.Run("invalid price", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		req := &entity.UpdateProductRequest{
			Id:          testID,
			Name:        "Valid Name",
			Description: "Valid description",
			Price:       0,
			Stock:       10,
		}

		_, err := svc.UpdateProduct(ctx, req)

		assert.ErrorIs(t, err, entity.ErrInvalidPrice)
	})

	t.Run("invalid stock", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		req := &entity.UpdateProductRequest{
			Id:          testID,
			Name:        "Valid Name",
			Description: "Valid description",
			Price:       100,
			Stock:       -1,
		}

		_, err := svc.UpdateProduct(ctx, req)

		assert.ErrorIs(t, err, entity.ErrInvalidStock)
	})

	t.Run("database update error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		req := &entity.UpdateProductRequest{
			Id:          testID,
			Name:        "Valid Name",
			Description: "Valid description",
			Price:       100,
			Stock:       10,
		}

		dbMock.On("Update", ctx, mock.AnythingOfType("*entity.Product")).Return(nil, fmt.Errorf("database update error"))

		_, err := svc.UpdateProduct(ctx, req)

		assert.ErrorContains(t, err, "database update error")
		dbMock.AssertExpectations(t)
	})

	t.Run("cache set error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)

		svc := NewProductService(dbMock, cacheMock)

		req := &entity.CreateProductRequest{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999,
			Stock:       10,
		}

		dbMock.On("Create", ctx, mock.Anything).Return(uuid.New(), nil)
		cacheMock.On("Set", ctx, mock.Anything, mock.Anything, duration).Return(errors.New("cache error"))

		_, err := svc.CreateProduct(ctx, req)
		assert.ErrorContains(t, err, "cache error")
	})

	t.Run("successful cache operations", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		dbMock.On("Update", ctx, mock.AnythingOfType("*entity.Product")).Return(testProduct, nil)

		cacheKey := utils.GenerateCacheKey("product", testID)

		productSerialized, err := utils.Serialize(testProduct)
		assert.NoError(t, err)

		cacheMock.On("Set", ctx, cacheKey, productSerialized, duration).Return(nil)

		err = svc.cache.Set(ctx, cacheKey, productSerialized, duration)

		assert.NoError(t, err)
		cacheMock.AssertExpectations(t)
	})

	t.Run("successful update", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		req := &entity.UpdateProductRequest{
			Id:          testID,
			Name:        "Updated Laptop",
			Description: "Updated high-performance laptop",
			Price:       1099,
			Stock:       15,
		}

		dbMock.On("Update", ctx, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == req.Name &&
				p.Description == req.Description &&
				p.Price == req.Price &&
				p.Stock == req.Stock
		})).Return(testProduct, nil)

		cacheKey := utils.GenerateCacheKey("product", testID)
		productSerialized, _ := utils.Serialize(testProduct)
		cacheMock.On("Set", ctx, cacheKey, productSerialized, duration).Return(nil)

		result, err := svc.UpdateProduct(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, testProduct, result)
		dbMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	testID := uuid.New()

	t.Run("successful delete", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		dbMock.On("Delete", ctx, testID).Return(nil)

		cacheKey := utils.GenerateCacheKey("product", testID)
		cacheMock.On("Delete", ctx, cacheKey).Return(nil)

		deleted, err := svc.DeleteProduct(ctx, testID)

		assert.NoError(t, err)
		assert.True(t, deleted)
		dbMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})

	t.Run("database delete error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		dbMock.On("Delete", ctx, testID).Return(fmt.Errorf("database delete error"))

		deleted, err := svc.DeleteProduct(ctx, testID)

		assert.ErrorContains(t, err, "database delete error")
		assert.False(t, deleted)
		dbMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})

	t.Run("cache delete error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		dbMock.On("Delete", ctx, testID).Return(nil)

		cacheKey := utils.GenerateCacheKey("product", testID)
		cacheMock.On("Delete", ctx, cacheKey).Return(fmt.Errorf("cache delete error"))

		deleted, err := svc.DeleteProduct(ctx, testID)

		assert.ErrorContains(t, err, "cache delete error")
		assert.False(t, deleted)
		dbMock.AssertExpectations(t)
		cacheMock.AssertExpectations(t)
	})
}

func TestService_List(t *testing.T) {
	ctx := context.Background()
	offset := int64(0)
	limit := int64(10)

	t.Run("successful list", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		testProducts := []*entity.Product{
			{
				Id:          uuid.New(),
				Name:        "Laptop",
				Description: "High-performance laptop",
				Price:       999,
				Stock:       10,
			},
			{
				Id:          uuid.New(),
				Name:        "Smartphone",
				Description: "Latest model smartphone",
				Price:       699,
				Stock:       20,
			},
		}

		dbMock.On("List", ctx, offset, limit).Return(testProducts, nil)

		products, err := svc.ListProduct(ctx, offset, limit)

		assert.NoError(t, err)
		assert.Equal(t, testProducts, products)
		dbMock.AssertExpectations(t)
	})

	t.Run("database list error", func(t *testing.T) {
		dbMock := new(MockDatabase)
		cacheMock := new(MockCache)
		svc := NewProductService(dbMock, cacheMock)

		dbMock.On("List", ctx, offset, limit).Return(nil, fmt.Errorf("database list error"))

		products, err := svc.ListProduct(ctx, offset, limit)

		assert.ErrorContains(t, err, "database list error")
		assert.Empty(t, products)
		dbMock.AssertExpectations(t)
	})
}
