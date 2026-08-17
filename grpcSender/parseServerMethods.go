package grpcSender

import (
	"context"
	"fmt"
	"time"

	"github.com/jhump/protoreflect/desc"

	"google.golang.org/grpc"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
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

	grpcSender.decodeDescriptorProtoWithReflection(conn)
	return &list, nil
}

func (grpcSender *GrpcSender) decodeDescriptorProtoWithReflection(conn *grpc.ClientConn) {
	client := grpcSender.getServerReflectionClient(conn)
	stream, _ := client.ServerReflectionInfo(context.Background())
	stream.Send(&rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: grpcSender.FullServiceName,
		},
	})
	resp, _ := stream.Recv()
	fd := &descriptorpb.FileDescriptorProto{}
	proto.Unmarshal(resp.GetFileDescriptorResponse().FileDescriptorProto[0], fd)
	fmt.Println(fd)
}
