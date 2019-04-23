package service

import (
	"context"
	"net"

	v1 "github.com/VideoCoin/cloud-api/accounts/v1"
	"github.com/VideoCoin/cloud-api/rpc"
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

func (s *RpcServer) Create(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	account, err := s.ds.Account.Create(req.OwnerID, s.secret)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	accountProfile := new(v1.AccountProfile)
	err = copier.Copy(accountProfile, account)
	if err != nil {
		return nil, rpc.ErrRpcInternal
	}

	return accountProfile, nil
}

func (s *RpcServer) List(ctx context.Context, req *protoempty.Empty) (*v1.ListResponse, error) {
	accounts, err := s.ds.Account.List()
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	accountProfiles := make([]*v1.AccountProfile, 0)
	err = copier.Copy(&accountProfiles, &accounts)
	if err != nil {
		return nil, rpc.ErrRpcInternal
	}

	return &v1.ListResponse{
		Items: accountProfiles,
	}, nil
}

func (s *RpcServer) Key(ctx context.Context, req *v1.AccountRequest) (*v1.AccountKey, error) {
	account, err := s.ds.Account.Get(req.OwnerID)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	accountKey := new(v1.AccountKey)
	err = copier.Copy(accountKey, account)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	return accountKey, nil
}

func (s *RpcServer) Get(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	err := req.Validate()
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	account, err := s.ds.Account.Get(req.OwnerID)
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

func (s *RpcServer) GetByAddress(ctx context.Context, req *v1.Address) (*v1.AccountProfile, error) {
	err := req.Validate()
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	account, err := s.ds.Account.GetByAddress(req.Address)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	accountProfile := new(v1.AccountProfile)
	err = copier.Copy(accountProfile, account)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	return accountProfile, nil
}

func (s *RpcServer) Refresh(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	err := req.Validate()
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	account, err := s.ds.Account.Get(req.OwnerID)
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
