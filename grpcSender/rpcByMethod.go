package grpcSender

import (
	"context"
	"fmt"
	"time"

	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
)

func (grpcSender *GrpcSender) executeRpcMethod() (response *rpcResponseData) {
	response = &rpcResponseData{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, refClient, err := grpcSender.getGrpcClient(ctx)
	if err != nil {
		response.Error = fmt.Errorf("Failed to connect to the server: %v", err)
		return
	}
	defer conn.Close()
	defer refClient.Reset()

	serviceDesc, err := refClient.ResolveService(grpcSender.FullServiceName)
	if err != nil {
		response.Error = fmt.Errorf("Service %s not found on the server: %v", grpcSender.FullServiceName, err)
		return
	}

	methodDesc := serviceDesc.FindMethodByName(grpcSender.Method)
	if methodDesc == nil {
		response.Error = fmt.Errorf("Method %s not found in the service %s", grpcSender.Method, grpcSender.FullServiceName)
		return
	}

	inputMsgDesc := methodDesc.GetInputType()
	dynamicInputMsg := dynamic.NewMessage(inputMsgDesc)

	err = dynamicInputMsg.UnmarshalJSON([]byte(grpcSender.Params))
	if err != nil {
		response.Error = fmt.Errorf("Error parsing input JSON: %v", err)
		return
	}

	outputMsgDesc := methodDesc.GetOutputType()
	dynamicOutputMsg := dynamic.NewMessage(outputMsgDesc)

	err = grpc.Invoke(ctx,
		fmt.Sprintf("/%s/%s", grpcSender.FullServiceName, grpcSender.Method),
		dynamicInputMsg,
		dynamicOutputMsg,
		conn,
	)
	if err != nil {
		response.Error = fmt.Errorf("RPC execution error: %v", err)
		return
	}

	jsonOutput, err := dynamicOutputMsg.MarshalJSON()
	if err != nil {
		response.Error = fmt.Errorf("Error converting response to JSON: %v", err)
		return
	}
	response.DataBytes = jsonOutput
	return
}
