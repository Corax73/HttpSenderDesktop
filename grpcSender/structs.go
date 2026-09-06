package grpcSender

import "encoding/json"

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

type CustomResponse struct {
	Data         json.RawMessage `json:"data"`
	RepeatNumber int             `json:"repeat_number"`
}
