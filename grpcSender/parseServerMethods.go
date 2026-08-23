package grpcSender

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	methodsSlice := make([]*methodDescription, 0)
	for _, m := range fd.MessageType {
		if strings.Contains(*m.Name, "Request") {
			description := methodDescription{Name: *m.Name}
			for _, f := range m.Field {
				field := fieldDescription{Name: f.GetJsonName(), Type: strings.ToLower(strings.ReplaceAll((f.GetType().String()), "TYPE_", ""))}
				description.Fields = append(description.Fields, &field)
			}
			methodsSlice = append(methodsSlice, &description)
		}
	}
	bytes, _ := json.Marshal(methodsSlice)
	fmt.Println(string(bytes))
}
