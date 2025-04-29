package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/utils"
)

const duration = time.Second * 30

type Database interface {
	Create(ctx context.Context, product *entity.Product) (uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) (*entity.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int64) ([]*entity.Product, error)
}

type Cache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type ProductService struct {
	db    Database
	cache Cache
}

func NewProductService(db Database, cache Cache) *ProductService {
	return &ProductService{db: db, cache: cache}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *entity.CreateProductRequest) (uuid.UUID, error) {
	const op = "ProductService.Create"

	if req.Name == "" || len(req.Name) > 99 {
		return uuid.Nil, fmt.Errorf("%s: %w", op, entity.ErrInvalidName)
	}

	if req.Description == "" || len(req.Description) > 9999 {
		return uuid.Nil, fmt.Errorf("%s: %w", op, entity.ErrInvalidDescription)
	}

	if req.Price < 1 {
		return uuid.Nil, fmt.Errorf("%s: %w", op, entity.ErrInvalidPrice)
	}

	if req.Stock < 0 {
		return uuid.Nil, fmt.Errorf("%s: %w", op, entity.ErrInvalidStock)
	}

	id := entity.GenerateID()
	now := time.Now()

	product := &entity.Product{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	createdId, err := s.db.Create(ctx, product)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	cacheKey := utils.GenerateCacheKey("product", id)
	productSerialized, err := utils.Serialize(product)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.cache.Set(ctx, cacheKey, productSerialized, duration)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return createdId, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	const op = "ProductService.Get"

	product := &entity.Product{}

	cacheKey := utils.GenerateCacheKey("product", id)
	cachedProduct, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		err := utils.Deserialize(cachedProduct, product)
		if err != nil {
			return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
		}

		return product, nil
	}

	product, err = s.db.Get(ctx, id)
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	cacheKey = utils.GenerateCacheKey("product", product.Id)
	productSerialized, err := utils.Serialize(product)
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	err = s.cache.Set(ctx, cacheKey, productSerialized, duration)
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	return product, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *entity.UpdateProductRequest) (*entity.Product, error) {
	const op = "ProductService.Update"

	if req.Name == "" || len(req.Name) > 99 {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, entity.ErrInvalidName)
	}

	if req.Description == "" || len(req.Description) > 9999 {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, entity.ErrInvalidDescription)
	}

	if req.Price < 1 {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, entity.ErrInvalidPrice)
	}

	if req.Stock < 0 {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, entity.ErrInvalidStock)
	}

	now := time.Now()

	product, err := s.db.Update(ctx, &entity.Product{
		Id:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		UpdatedAt:   now,
	})
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	cacheKey := utils.GenerateCacheKey("product", req.Id)
	productSerialized, err := utils.Serialize(product)
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	err = s.cache.Set(ctx, cacheKey, productSerialized, duration)
	if err != nil {
		return &entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) (bool, error) {
	const op = "ProductService.Delete"

	err := s.db.Delete(ctx, id)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	cacheKey := utils.GenerateCacheKey("product", id)
	err = s.cache.Delete(ctx, cacheKey)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *ProductService) ListProduct(ctx context.Context, offset, limit int64) ([]*entity.Product, error) {
	const op = "ProductService.List"

	products, err := s.db.List(ctx, offset, limit)
	if err != nil {
		return []*entity.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}
