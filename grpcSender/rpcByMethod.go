package grpcSender

import (
	"context"
	"fmt"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
)

func (grpcSender *GrpcSender) executeRpcMethod(
	ctx context.Context,
	conn *grpc.ClientConn,
	refClient *grpcreflect.Client,
	ch chan *rpcResponseData,
	repeatNumber int,
) (response *rpcResponseData) {
	response = &rpcResponseData{}
	response.RepeatNumber = repeatNumber
	serviceDesc, err := refClient.ResolveService(grpcSender.FullServiceName)
	if err != nil {
		response.Error = fmt.Errorf("Service %s not found on the server: %v", grpcSender.FullServiceName, err)
		ch <- response
		return
	}

	methodDesc := serviceDesc.FindMethodByName(grpcSender.Method)
	if methodDesc == nil {
		response.Error = fmt.Errorf("Method %s not found in the service %s", grpcSender.Method, grpcSender.FullServiceName)
		ch <- response
		return
	}

	inputMsgDesc := methodDesc.GetInputType()
	dynamicInputMsg := dynamic.NewMessage(inputMsgDesc)

	err = dynamicInputMsg.UnmarshalJSON([]byte(grpcSender.Params))
	if err != nil {
		response.Error = fmt.Errorf("Error parsing input JSON: %v", err)
		ch <- response
		return
	}

	outputMsgDesc := methodDesc.GetOutputType()
	dynamicOutputMsg := dynamic.NewMessage(outputMsgDesc)

	err = conn.Invoke(ctx,
		fmt.Sprintf("/%s/%s", grpcSender.FullServiceName, grpcSender.Method),
		dynamicInputMsg,
		dynamicOutputMsg,
	)
	if err != nil {
		response.Error = fmt.Errorf("RPC execution error: %v", err)
		ch <- response
		return
	}

	jsonOutput, err := dynamicOutputMsg.MarshalJSON()
	if err != nil {
		response.Error = fmt.Errorf("Error converting response to JSON: %v", err)
		ch <- response
		return
	}
	response.DataBytes = jsonOutput
	ch <- response
	return
}
