package server

import (
	"Order/entities"
	pb "Order/grpc"
	"context"
	"errors"
	"fmt"
	"strconv"
)

func (server *Server) CreateOrder(ctx context.Context, in *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {

	order := &entities.Order{
		ClientID: uint(in.UserID),
		Amount:   float32(in.Amount),
		Status:   pb.Status_CREATED.String(),
	}

	server.mutex.Lock()
	order, err := server.storage.AddNewOrder(order)
	server.mutex.Unlock()
	if err != nil {
		return nil, err
	}

	server.mutex.Lock()
	err = server.kafkaConnection.SendMessageOrder(*order)
	server.mutex.Unlock()
	if err != nil {
		return nil, err
	}

	order.Status = <-server.kafkaConnection.GetMessages[order.ID]
	delete(server.kafkaConnection.GetMessages, order.ClientID)

	server.mutex.Lock()
	err = server.storage.UpdateStatus(order)
	server.mutex.Unlock()
	if err != nil {
		return nil, err
	}

	if order.Status == "CANCELED" {
		return nil, errors.New("CANCELED")
	}
	return &pb.CreateOrderResponse{TransactionID: fmt.Sprintf("%d", order.ID)}, nil
}

func (server *Server) GetOrder(ctx context.Context, in *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {

	id, err := strconv.Atoi(in.GetTransactionID())
	if err != nil {
		return nil, fmt.Errorf("INCORRECT ID, %d", err)
	}

	order, err := server.storage.GetOrderById(uint(id))
	if err != nil {
		return nil, err
	}

	status, err := Strtostat(order.Status)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrderResponse{
		ID:       uint64(order.ID),
		ClientID: uint64(order.ClientID),
		Amount:   float64(order.Amount),
		Status:   status,
	}, nil
}

func Strtostat(str string) (pb.Status, error) {
	switch str {
	case "CREATED":
		return pb.Status_CREATED, nil
	case "PAID":
		return pb.Status_PAID, nil
	case "CANCELED":
		return pb.Status_CANCELED, nil
	default:
		return 0, errors.New("STATUS NOT EXIST")
	}
}
