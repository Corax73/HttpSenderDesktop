package grpcSender

import (
	"context"
	"fmt"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func (grpcSender *GrpcSender) parseServerMethods(serverAddr, fullServiceName string) (*[]*desc.MethodDescriptor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the server: %v", err)
	}
	defer conn.Close()

	refClient := grpcreflect.NewClientV1Alpha(ctx, reflectpb.NewServerReflectionClient(conn))
	defer refClient.Reset()

	serviceDesc, err := refClient.ResolveService(fullServiceName)
	if err != nil {
		return nil, fmt.Errorf("Service %s not found on the server: %v", fullServiceName, err)
	}
	list := serviceDesc.GetMethods()
	return &list, nil
}
