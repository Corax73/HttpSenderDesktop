package grpcSender

import (
	"context"

	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func (grpcSender *GrpcSender) getGrpcClient(ctx context.Context) (conn *grpc.ClientConn, client *grpcreflect.Client, err error) {
	conn, err = grpc.NewClient(grpcSender.Url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client = grpcreflect.NewClientV1Alpha(ctx, grpcSender.getServerReflectionClient(conn))
	return
}

func (grpcSender *GrpcSender) getServerReflectionClient(conn *grpc.ClientConn) (clientInst reflectpb.ServerReflectionClient) {
	clientInst = reflectpb.NewServerReflectionClient(conn)
	return clientInst
}
