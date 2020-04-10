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
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerOptions struct {
	Logger  *logrus.Entry
	Addr    string
	DS      *datastore.Datastore
	EB      *ebus.EventBus
	Manager *manager.Manager
}

type Server struct {
	logger  *logrus.Entry
	addr    string
	grpc    *grpc.Server
	listen  net.Listener
	ds      *datastore.Datastore
	manager *manager.Manager
}

func NewServer(opts *ServerOptions) (*Server, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcServer := grpc.NewServer(grpcOpts...)

	healthService := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthService)

	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}

	rpcServer := &Server{
		logger:  opts.Logger,
		addr:    opts.Addr,
		grpc:    grpcServer,
		listen:  listen,
		ds:      opts.DS,
		manager: opts.Manager,
	}

	v1.RegisterAccountServiceServer(grpcServer, rpcServer)
	reflection.Register(grpcServer)

	return rpcServer, nil
}

func (s *Server) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}
