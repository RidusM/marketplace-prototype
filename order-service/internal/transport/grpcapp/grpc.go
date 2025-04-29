package grpcapp

import (
	"context"

	"github.com/google/uuid"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/api/client"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service interface {
	CreateOrder(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*entity.Order, error)
	UpdateOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, status string, total uint64) (*entity.Order, error)
	DeleteOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error
	ListOrdersByUser(ctx context.Context, userID uuid.UUID) ([]entity.Order, error)
	AddItemOrder(ctx context.Context, orderID uuid.UUID, productID uuid.UUID, productPrice uint64, quantity int) (uuid.UUID, error)
	DeleteItemOrder(ctx context.Context, itemID uuid.UUID) error
	UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error
	ListItemsByOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) ([]entity.Item, error)
}

type OrderService struct {
	client.UnimplementedOrderServiceServer
	service Service
}

func NewOrderService(svc Service) *OrderService {
	return &OrderService{service: svc}
}

func (s *OrderService) AddItemToOrder(ctx context.Context, req *client.AddItemRequest) (*client.AddItemResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	productID, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	itemID, err := s.service.AddItemOrder(ctx, orderID, productID, req.GetProductPrice(), int(req.GetQuantity()))
	if err != nil {
		return nil, HandleErrors(err)
	}

	return &client.AddItemResponse{ItemId: itemID.String()}, nil
}

func (s *OrderService) RemoveItemFromOrder(ctx context.Context, req *client.RemoveItemRequest) (*emptypb.Empty, error) {
	const op = "grpcapp.RemoveItemFromOrder"

	itemID, err := uuid.Parse(req.GetItemId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	if err = s.service.DeleteItemOrder(ctx, itemID); err != nil {
		return nil, HandleErrors(err)
	}

	return &emptypb.Empty{}, nil
}
func (s *OrderService) UpdateItemInOrder(ctx context.Context, req *client.UpdateItemRequest) (*emptypb.Empty, error) {
	const op = "grpcapp.UpdateItemInOrder"

	itemID, err := uuid.Parse(req.GetItemId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	if err = s.service.UpdateItem(ctx, itemID, int(req.GetQuantity())); err != nil {
		return nil, HandleErrors(err)
	}

	return &emptypb.Empty{}, nil
}
func (s *OrderService) ListItemsFromOrder(ctx context.Context, req *client.ListItemsRequest) (*client.ListItemsResponse, error) {
	const op = "grpcapp.ListItemsFromOrder"

	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {

		return nil, HandleErrors(err)
	}

	items, err := s.service.ListItemsByOrder(ctx, orderID, userID)
	if err != nil {

		return nil, HandleErrors(err)
	}

	return &client.ListItemsResponse{Items: MapToGrpcItemsList(items)}, nil
}

func (s *OrderService) AddItemOrder(ctx context.Context, req *client.AddItemRequest) (*client.AddItemResponse, error) {
	const op = "grpcapp.ListItemsFromOrder"

	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	productID, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	itemID, err := s.service.AddItemOrder(ctx, orderID, productID, req.GetProductPrice(), int(req.GetQuantity()))
	if err != nil {

		return nil, HandleErrors(err)
	}

	return &client.AddItemResponse{ItemId: itemID.String()}, nil
}
func (s *OrderService) CancelOrder(ctx context.Context, req *client.CancelOrderRequest) (*emptypb.Empty, error) {
	const op = "grpcapp.CancelOrder"

	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	err = s.service.DeleteOrder(ctx, userID, orderID)
	if err != nil {
		return nil, HandleErrors(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req *client.CreateOrderRequest) (*client.CreateOrderResponse, error) {
	const op = "grpcapp.CreateOrder"

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	orderID, err := s.service.CreateOrder(ctx, userID)
	if err != nil {
		return nil, HandleErrors(err)
	}

	return &client.CreateOrderResponse{OrderId: orderID.String()}, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *client.UpdateOrderStatusRequest) (*client.UpdateOrderStatusResponse, error) {
	const op = "grpcapp.updateOrderStatus"

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	order, err := s.service.GetOrder(ctx, userID, orderID)
	if err != nil {
		return nil, HandleErrors(err)
	}

	updOrder, err := s.service.UpdateOrder(ctx, userID, orderID, req.GetStatus(), order.TotalAmount)
	if err != nil {
		return nil, HandleErrors(err)
	}

	return &client.UpdateOrderStatusResponse{Order: MapToGrpcOrder(*updOrder)}, nil
}

func (s *OrderService) UpdateOrderTotal(ctx context.Context, req *client.UpdateOrderTotalRequest) (*client.UpdateOrderTotalResponse, error) {
	const op = "grpcapp.UpdateOrderTotal"

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	order, err := s.service.GetOrder(ctx, userID, orderID)
	if err != nil {
		return nil, HandleErrors(err)
	}

	updOrder, err := s.service.UpdateOrder(ctx, userID, orderID, order.Status, uint64(req.GetNewTotal()))
	if err != nil {
		return nil, HandleErrors(err)
	}

	return &client.UpdateOrderTotalResponse{Order: MapToGrpcOrder(*updOrder)}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *client.GetOrderRequest) (*client.GetOrderResponse, error) {
	const op = "grpcapp.GetOrder"

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	order, err := s.service.GetOrder(ctx, userID, orderID)
	if err != nil {
		return nil, HandleErrors(err)
	}

	return &client.GetOrderResponse{Order: MapToGrpcOrder(*order)}, nil
}

func (s *OrderService) ListOrdersByUser(ctx context.Context, req *client.ListOrdersRequest) (*client.ListOrdersResponse, error) {
	const op = "grpcapp.ListOrdersByUser"

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, HandleErrors(err)
	}

	orders, err := s.service.ListOrdersByUser(ctx, userID)
	if err != nil {
		return nil, HandleErrors(err)
	}

	return &client.ListOrdersResponse{Orders: MapToGrpcOrdersList(orders)}, nil
}
