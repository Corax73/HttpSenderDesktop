package grpcSender

import (
	"context"
	"fmt"
	"time"

	"github.com/jhump/protoreflect/desc"
)

func (grpcSender *GrpcSender) parseServerMethods() (*[]*desc.MethodDescriptor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, refClient, err := grpcSender.getGrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the server: %v", err)
	}
	defer conn.Close()
	defer refClient.Reset()

	serviceDesc, err := refClient.ResolveService(grpcSender.FullServiceName)
	if err != nil {
		return nil, fmt.Errorf("Service %s not found on the server: %v", grpcSender.FullServiceName, err)
	}
	list := serviceDesc.GetMethods()
	return &list, nil
}
