package service

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	ErrRpcUserAlreadyExists           = grpc.Errorf(codes.AlreadyExists, "User is already registered")
	ErrRpcUserInvalidVerificationCode = grpc.Errorf(codes.InvalidArgument, "Invalid code verification")
	ErrRpcUserAlreadyVerified         = grpc.Errorf(codes.AlreadyExists, "User is already verified")
	ErrRpcTFASessionInvalid           = grpc.Errorf(codes.InvalidArgument, "Invalid request data")
)
