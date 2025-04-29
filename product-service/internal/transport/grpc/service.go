package grpcServer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/product-microservice/pkg/api/product"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	CreateProduct(ctx context.Context, input *entity.CreateProductRequest) (uuid.UUID, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	UpdateProduct(ctx context.Context, input *entity.UpdateProductRequest) (*entity.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) (bool, error)
	ListProduct(ctx context.Context, offset, limit int64) ([]*entity.Product, error)
}

type ProductService struct {
	product.UnimplementedProductServiceServer
	service Service
}

func NewProductService(service Service) *ProductService {
	return &ProductService{service: service}
}

func (t *ProductService) CreateProduct(ctx context.Context, input *product.CreateProductRequest) (*product.CreateProductResponse, error) {
	const op = "Service.CreateProduct"

	data := &entity.CreateProductRequest{
		Name:        input.GetName(),
		Description: input.GetDescription(),
		Price:       input.GetPrice(),
		Stock:       input.GetStock(),
	}

	id, err := t.service.CreateProduct(ctx, data)
	if err != nil {
		return &product.CreateProductResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return &product.CreateProductResponse{
		Id: id.String(),
	}, nil
}

func (t *ProductService) GetProduct(ctx context.Context, input *product.GetProductRequest) (*product.GetProductResponse, error) {
	const op = "Service.CreateProduct"

	id, err := ParseUUID(input.GetId())
	if err != nil {
		return &product.GetProductResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	products, err := t.service.GetProduct(ctx, id)
	if err != nil {
		return &product.GetProductResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return &product.GetProductResponse{
		Product: &product.Product{
			Id:          input.Id,
			Name:        products.Name,
			Description: products.Description,
			Price:       products.Price,
			Stock:       products.Stock,
			CreatedAt:   timestamppb.New(products.CreatedAt),
			UpdatedAt:   timestamppb.New(products.UpdatedAt),
		},
	}, nil
}

func (t *ProductService) UpdateProduct(ctx context.Context, input *product.UpdateProductRequest) (*product.UpdateProductResponse, error) {
	const op = "Service.CreateProduct"

	id, err := ParseUUID(input.GetId())
	if err != nil {
		return &product.UpdateProductResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	data := &entity.UpdateProductRequest{
		Id:          id,
		Name:        input.GetName(),
		Description: input.GetDescription(),
		Price:       input.GetPrice(),
		Stock:       input.GetStock(),
	}

	products, err := t.service.UpdateProduct(ctx, data)
	if err != nil {
		return &product.UpdateProductResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return &product.UpdateProductResponse{
		Product: &product.Product{
			Id:          products.Id.String(),
			Name:        products.Name,
			Description: products.Description,
			Price:       products.Price,
			Stock:       products.Stock,
			CreatedAt:   timestamppb.New(products.CreatedAt),
			UpdatedAt:   timestamppb.New(products.UpdatedAt),
		},
	}, nil
}

func (t *ProductService) DeleteProduct(ctx context.Context, input *product.DeleteProductRequest) (*product.DeleteProductResponse, error) {
	const op = "Service.CreateProduct"

	id, err := ParseUUID(input.GetId())
	if err != nil {
		return &product.DeleteProductResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	success, err := t.service.DeleteProduct(ctx, id)
	if err != nil {
		return &product.DeleteProductResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return &product.DeleteProductResponse{
		Success: success,
	}, nil
}

func (t *ProductService) ListProducts(ctx context.Context, input *product.ListProductsRequest) (*product.ListProductsResponse, error) {
	const op = "Service.CreateProduct"

	products, err := t.service.ListProduct(ctx, input.GetOffset(), input.GetLimit())
	if err != nil {
		return &product.ListProductsResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	var protobufProducts []*product.Product
	for _, products := range products {
		protobufProducts = append(protobufProducts, &product.Product{
			Id:          products.Id.String(),
			Name:        products.Name,
			Description: products.Description,
			Price:       products.Price,
			Stock:       products.Stock,
			CreatedAt:   timestamppb.New(products.CreatedAt),
			UpdatedAt:   timestamppb.New(products.UpdatedAt),
		})
	}

	return &product.ListProductsResponse{
		Products: protobufProducts,
	}, nil

}

func ParseUUID(id string) (uuid.UUID, error) {
	const op = "ParseUUID"

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return parsedUUID, nil
}
