package rpc

import (
	"context"

	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/videocoin/cloud-accounts/datastore"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	"github.com/videocoin/cloud-api/rpc"
)

func (s *RpcServer) Create(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("owner_id", req.OwnerId)

	profile, err := s.manager.CreateAccount(ctx, req)
	if err != nil {
		logFailedTo(s.logger, "create account", err)
		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}

func (s *RpcServer) List(ctx context.Context, req *protoempty.Empty) (*v1.ListResponse, error) {
	_ = opentracing.SpanFromContext(ctx)

	profiles, err := s.manager.ListAccounts(ctx)
	if err != nil {
		logFailedTo(s.logger, "list accounts", err)
		return nil, rpc.ErrRpcInternal
	}

	return &v1.ListResponse{
		Items: profiles,
	}, nil
}

func (s *RpcServer) Key(ctx context.Context, req *v1.AccountRequest) (*v1.AccountKey, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("owner_id", req.OwnerId)

	key, err := s.manager.GetAccountKey(ctx, req.OwnerId)
	if err != nil {
		logFailedTo(s.logger, "get account key", err)
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		return nil, rpc.ErrRpcInternal
	}

	return key, nil
}

func (s *RpcServer) Get(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	profile, err := s.manager.GetAccountById(ctx, req.Id)
	if err != nil {
		logFailedTo(s.logger, "get account by id", err)
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}

func (s *RpcServer) GetByOwner(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("owner_id", req.OwnerId)

	profile, err := s.manager.GetAccountByOwner(ctx, req.OwnerId)
	if err != nil {
		logFailedTo(s.logger, "get account by owner", err)
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}

func (s *RpcServer) GetByAddress(ctx context.Context, req *v1.Address) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("address", req.Address)

	profile, err := s.manager.GetAccountByAddress(ctx, req.Address)
	if err != nil {
		logFailedTo(s.logger, "get account by address", err)
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}

func (s *RpcServer) Withdraw(ctx context.Context, req *v1.WithdrawRequest) (*protoempty.Empty, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("owner_id", req.OwnerId)
	span.SetTag("transfer_id", req.TransferId)

	key, err := s.manager.GetAccountKey(ctx, req.OwnerId)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		s.logger.WithError(err).Error("failed to get account key")
		return nil, rpc.ErrRpcInternal
	}

	s.manager.Withdraw(key, req)

	return new(protoempty.Empty), nil
}
