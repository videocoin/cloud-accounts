package service

import (
	"context"
	"net"

	v1 "github.com/VideoCoin/cloud-api/accounts/v1"
	"github.com/VideoCoin/cloud-api/rpc"
	"github.com/VideoCoin/cloud-pkg/auth"
	"github.com/VideoCoin/cloud-pkg/grpcutil"
	protoempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type RpcServerOptions struct {
	Addr   string
	Secret string
	Logger *logrus.Entry
	DS     *Datastore
	EB     *EventBus
}

type RpcServer struct {
	addr   string
	secret string
	grpc   *grpc.Server
	listen net.Listener
	logger *logrus.Entry
	ds     *Datastore
	eb     *EventBus
}

func NewRpcServer(opts *RpcServerOptions) (*RpcServer, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)

	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	rpcServer := &RpcServer{
		addr:   opts.Addr,
		secret: opts.Secret,
		grpc:   grpcServer,
		listen: listen,
		logger: opts.Logger,
		ds:     opts.DS,
		eb:     opts.EB,
	}

	v1.RegisterAccountServiceServer(grpcServer, rpcServer)
	reflection.Register(grpcServer)

	return rpcServer, nil
}

func (s *RpcServer) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}

func (s *RpcServer) Health(ctx context.Context, req *protoempty.Empty) (*rpc.HealthStatus, error) {
	return &rpc.HealthStatus{Status: "OK"}, nil
}

func (s *RpcServer) Create(ctx context.Context, req *v1.CreateAccountRequest) (*v1.AccountProfile, error) {
	account, err := s.ds.Account.Create(req.OwnerID, s.secret)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	accountProfile := new(v1.AccountProfile)
	copier.Copy(accountProfile, account)

	return accountProfile, nil
}

func (s *RpcServer) Get(ctx context.Context, req *protoempty.Empty) (*v1.AccountProfile, error) {
	ctx, err := auth.AuthFromContext(ctx)
	if err != nil {
		return nil, rpc.ErrRpcUnauthenticated
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, rpc.ErrRpcUnauthenticated
	}

	account, err := s.ds.Account.Get(userID)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	accountProfile := new(v1.AccountProfile)
	err = copier.Copy(accountProfile, account)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	return accountProfile, nil
}

func (s *RpcServer) Refresh(ctx context.Context, req *protoempty.Empty) (*v1.AccountProfile, error) {
	ctx, err := auth.AuthFromContext(ctx)
	if err != nil {
		return nil, rpc.ErrRpcUnauthenticated
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, rpc.ErrRpcUnauthenticated
	}

	account, err := s.ds.Account.Get(userID)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	accountProfile := new(v1.AccountProfile)
	err = copier.Copy(accountProfile, account)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	return accountProfile, nil
}
