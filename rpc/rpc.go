package rpc

import (
	"context"

	protoempty "github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/videocoin/cloud-accounts/datastore"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	"github.com/videocoin/cloud-api/rpc"
)

func (s *Server) Create(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("owner_id", req.OwnerId)

	profile, err := s.manager.CreateAccount(ctx, req)
	if err != nil {
		logFailedTo(s.logger, "create account", err)
		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}

func (s *Server) List(ctx context.Context, req *protoempty.Empty) (*v1.ListResponse, error) {
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

func (s *Server) Key(ctx context.Context, req *v1.AccountRequest) (*v1.AccountKey, error) {
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

func (s *Server) Keys(ctx context.Context, req *protoempty.Empty) (*v1.AccountKeys, error) {
	_ = opentracing.SpanFromContext(ctx)

	keys, err := s.manager.GetAccountKeys(ctx)
	if err != nil {
		logFailedTo(s.logger, "get account keys", err)
		return nil, rpc.ErrRpcInternal
	}

	return &v1.AccountKeys{
		Items: keys,
	}, nil
}

func (s *Server) Get(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	profile, err := s.manager.GetAccountByID(ctx, req.Id)
	if err != nil {
		logFailedTo(s.logger, "get account by id", err)
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}

func (s *Server) GetByOwner(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
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

func (s *Server) GetByAddress(ctx context.Context, req *v1.Address) (*v1.AccountProfile, error) {
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

func (s *Server) CreateTransfer(ctx context.Context, req *v1.CreateTransferRequest) (*v1.TransferResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("user_id", req.UserId)
	span.SetTag("to_address", req.ToAddress)

	transfer, err := s.manager.CreateTransfer(ctx, req)
	if err != nil {
		s.logger.WithError(err).Error("failed to create transfer")
		return nil, rpc.ErrRpcInternal
	}

	return transfer, nil
}

func (s *Server) GetTransfer(ctx context.Context, req *v1.TransferRequest) (*v1.TransferResponse, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	transfer, err := s.manager.GetTransfer(ctx, req)
	if err != nil {
		if err == datastore.ErrTransferNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		s.logger.WithError(err).Error("failed to get transfer")
		return nil, rpc.ErrRpcInternal
	}

	return transfer, nil
}

func (s *Server) ExecuteTransfer(ctx context.Context, req *v1.ExecuteTransferRequest) (*protoempty.Empty, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)
	span.SetTag("user_id", req.UserId)
	span.SetTag("user_email", req.UserEmail)

	key, err := s.manager.GetAccountKey(ctx, req.UserId)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}
		s.logger.WithError(err).Error("failed to get account key")
		return nil, rpc.ErrRpcInternal
	}

	s.manager.ExecuteTransfer(ctx, key, req)

	return new(protoempty.Empty), nil
}
