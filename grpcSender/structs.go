package grpcSender

import (
	common "httpSenderDesktop/common/structs"
)

type rpcResponseData struct {
	Error        error
	DataBytes    []byte
	RepeatNumber int
}

type methodDescription struct {
	Name       string              `json:"name"`
	MethodName string              `json:"method_name"`
	Fields     []*fieldDescription `json:"fields"`
}

type fieldDescription struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type state struct {
	Url, FullServiceName, Params, Method, ResponseData string
	Repeat, Delay                                      int
	MethodsDescription                                 []*methodDescription
	Responses                                          []*common.CustomResponse
	NotShowResult                                      bool
}
