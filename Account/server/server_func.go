package server

import (
	pb "Account/grpc"
	"context"
)

func (server *Server) CreateAccount(ctx context.Context, in *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	server.storage.MU.Lock()
	defer server.storage.MU.Unlock()
	account, err := server.storage.CreateAccount(in.GetDeposit())
	if err != nil {
		return nil, err
	}
	return &pb.CreateAccountResponse{
		ID: uint32(account),
	}, nil
}

func (server *Server) GetDeposit(ctx context.Context, in *pb.GetDepositRequest) (*pb.GetDepositResponse, error) {

	deposit, err := server.storage.GetDeposit(uint(in.GetID()))
	if err != nil {
		return nil, err
	}
	return &pb.GetDepositResponse{
		Deposit: deposit,
	}, nil
}

func (server *Server) IncreaseDeposit(ctx context.Context, in *pb.IncreaseDepositRequest) (*pb.IncreaseDepositResponse, error) {
	server.storage.MU.Lock()
	defer server.storage.MU.Unlock()
	success := server.storage.IncreaseDeposit(uint(in.GetID()), in.GetDeposit())
	return &pb.IncreaseDepositResponse{
		Success: success,
	}, nil
}
