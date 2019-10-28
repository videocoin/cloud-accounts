package rpc

import (
	"context"
	"net"

	protoempty "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/datastore"
	"github.com/videocoin/cloud-accounts/ebus"
	"github.com/videocoin/cloud-accounts/manager"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	"github.com/videocoin/cloud-api/rpc"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type RpcServerOptions struct {
	Addr    string
	DS      *datastore.Datastore
	EB      *ebus.EventBus
	Manager *manager.Manager

	ClientSecret string

	Logger *logrus.Entry
}

type RpcServer struct {
	addr   string
	grpc   *grpc.Server
	listen net.Listener

	ds      *datastore.Datastore
	eb      *ebus.EventBus
	manager *manager.Manager

	clientSecret string

	logger *logrus.Entry
}

func NewRpcServer(opts *RpcServerOptions) (*RpcServer, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)

	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	rpcServer := &RpcServer{
		addr:         opts.Addr,
		grpc:         grpcServer,
		listen:       listen,
		ds:           opts.DS,
		eb:           opts.EB,
		manager:      opts.Manager,
		clientSecret: opts.ClientSecret,
		logger:       opts.Logger,
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
