package rpc

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/datastore"
	"github.com/videocoin/cloud-accounts/ebus"
	"github.com/videocoin/cloud-accounts/manager"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerOptions struct {
	Addr    string
	DS      *datastore.Datastore
	EB      *ebus.EventBus
	Manager *manager.Manager

	ClientSecret string

	Logger *logrus.Entry
}

type Server struct {
	addr   string
	grpc   *grpc.Server
	listen net.Listener

	ds      *datastore.Datastore
	eb      *ebus.EventBus
	manager *manager.Manager

	clientSecret string

	logger *logrus.Entry
}

func NewServer(opts *ServerOptions) (*Server, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)
	healthService := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	rpcServer := &Server{
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

func (s *Server) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}
