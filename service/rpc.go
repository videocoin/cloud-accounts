package service

import (
	"context"
	"fmt"
	"net"

	v1 "github.com/VideoCoin/cloud-api/accounts/v1"
	"github.com/VideoCoin/cloud-api/rpc"
	"github.com/VideoCoin/cloud-pkg/grpcutil"
	"github.com/VideoCoin/go-videocoin/common"
	"github.com/VideoCoin/go-videocoin/ethclient"
	protoempty "github.com/gogo/protobuf/types"
	"github.com/jinzhu/copier"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type RpcServerOptions struct {
	Addr         string
	NodeHTTPAddr string
	Secret       string
	Logger       *logrus.Entry
	DS           *Datastore
	EB           *EventBus
}

type RpcServer struct {
	addr   string
	secret string
	grpc   *grpc.Server
	listen net.Listener
	logger *logrus.Entry
	ds     *Datastore
	eb     *EventBus
	ec     *ethclient.Client
}

func NewRpcServer(opts *RpcServerOptions) (*RpcServer, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)

	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(opts.NodeHTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial eth client: %s", err.Error())
	}

	rpcServer := &RpcServer{
		addr:   opts.Addr,
		secret: opts.Secret,
		grpc:   grpcServer,
		listen: listen,
		logger: opts.Logger,
		ds:     opts.DS,
		eb:     opts.EB,
		ec:     ethClient,
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "Create")
	defer span.Finish()

	span.LogFields(
		log.String("owner_id", req.OwnerId),
	)

	account, err := s.ds.Account.Create(ctx, req.OwnerId, s.secret)
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	accounts, err := s.ds.Account.List(ctx)
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "Key")
	defer span.Finish()

	span.LogFields(
		log.String("owner_id", req.OwnerId),
	)

	account, err := s.ds.Account.GetByOwner(ctx, req.OwnerId)
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.LogFields(
		log.String("id", req.Id),
	)

	account, err := s.ds.Account.Get(ctx, req.Id)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	address := common.HexToAddress(account.Address)
	balanceWei, err := s.ec.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}

	balanceVdc, err := wei2Vdc(balanceWei)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	balance, _ := balanceVdc.Float64()
	if err = s.ds.Account.UpdateBalance(ctx, account, balance); err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	accountProfile := new(v1.AccountProfile)
	if err := copier.Copy(accountProfile, account); err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	return accountProfile, nil
}

func (s *RpcServer) GetByOwner(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetByOwner")
	defer span.Finish()

	span.LogFields(
		log.String("id", req.Id),
	)

	account, err := s.ds.Account.GetByOwner(ctx, req.OwnerId)
	if err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	accountProfile := new(v1.AccountProfile)
	if err := copier.Copy(accountProfile, account); err != nil {
		s.logger.Error(err)
		return nil, rpc.ErrRpcInternal
	}

	return accountProfile, nil
}

func (s *RpcServer) GetByAddress(ctx context.Context, req *v1.Address) (*v1.AccountProfile, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetByAddress")
	defer span.Finish()

	span.LogFields(
		log.String("address", req.Address),
	)

	account, err := s.ds.Account.GetByAddress(ctx, req.Address)
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
