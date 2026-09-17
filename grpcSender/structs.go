package grpcSender

import (
	common "httpSenderDesktop/common/structs"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

type grpcState struct {
	common.State
	FullServiceName    string
	MethodsDescription []*methodDescription
	Responses          []*common.CustomResponse
}

type GrpcSender struct {
	common.Sender
	grpcState
	FullServiceNameEntry                                               *widget.Entry
	ScrollContainer                                                    *container.Scroll
	ParseMethodsBtn, CopyMethodDescriptionBtn, ResultCopyBtnHandlerBtn *widget.Button
	SelectMethod                                                       *widget.Select
	MethodDescriptionDisplay                                           *widget.Label
	NotShowResultCheckbox                                              *widget.Check
}
