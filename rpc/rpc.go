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

func (s *Server) Key(ctx context.Context, req *v1.AccountRequest) (*v1.AccountKey, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("owner_id", req.OwnerId)

	key, err := s.manager.GetAccountKey(ctx, req.OwnerId)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		logFailedTo(s.logger, "get account key", err)

		return nil, rpc.ErrRpcInternal
	}

	return key, nil
}

func (s *Server) Get(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("id", req.Id)

	profile, err := s.manager.GetAccountByID(ctx, req.Id)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		logFailedTo(s.logger, "get account by id", err)

		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}

func (s *Server) List(ctx context.Context, req *protoempty.Empty) (*v1.Accounts, error) {
	accounts, err := s.manager.GetAccounts(ctx)
	if err != nil {
		logFailedTo(s.logger, "get accounts", err)
		return nil, rpc.ErrRpcInternal
	}

	return accounts, nil
}

func (s *Server) GetByOwner(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("owner_id", req.OwnerId)

	profile, err := s.manager.GetAccountByOwner(ctx, req.OwnerId)
	if err != nil {
		if err == datastore.ErrAccountNotFound {
			return nil, rpc.ErrRpcNotFound
		}

		logFailedTo(s.logger, "get account by owner", err)

		return nil, rpc.ErrRpcInternal
	}

	return profile, nil
}
