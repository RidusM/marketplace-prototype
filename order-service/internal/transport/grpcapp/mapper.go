package grpcapp

import (
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/internal/entity"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/order-microservice/pkg/api/client"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapToGrpcOrder(order entity.Order) *client.Order {
	return &client.Order{OrderId: order.OrderID.String(),
		UserId:      order.UserID.String(),
		TotalAmount: uint64(order.TotalAmount),
		Status:      order.Status,
		CreatedAt:   timestamppb.New(order.CreatedAt),
		UpdatedAt:   timestamppb.New(order.UpdatedAt),
	}
}

func MapToGrpcOrdersList(orders []entity.Order) []*client.Order {
	grpcOrders := make([]*client.Order, len(orders))

	for idx, order := range orders {
		grpcOrders[idx] = MapToGrpcOrder(order)
	}

	return grpcOrders
}

func MapToGrpcItem(item entity.Item) *client.Item {
	return &client.Item{
		ItemId:    item.ItemID.String(),
		ProductId: item.ProductID.String(),
		Quantity:  int64(item.Quantity),
		Price:     int64(item.Price)}
}

func MapToGrpcItemsList(items []entity.Item) []*client.Item {
	grpcItems := make([]*client.Item, len(items))

	for idx, item := range items {
		grpcItems[idx] = MapToGrpcItem(item)
	}

	return grpcItems
}
