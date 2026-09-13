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
	Url, FullServiceName, Params, Method string
	Repeat, Delay                        int
	MethodsDescription                   []*methodDescription
	Responses                            []*common.CustomResponse
	NotShowResult                        bool
}

type GrpcSender struct {
	common.Sender
	grpcState
	UrlEntry, FullServiceNameEntry, ParamsEntry, RepeatEntry, DelayEntry *widget.Entry
	ScrollContainer                                                      *container.Scroll
	ParseMethodsBtn, SendBtn, ClearResultBtn,
	ClearParametersBtn, CopyMethodDescriptionBtn, ResultCopyBtnHandlerBtn,
	SaveResultBtn *widget.Button
	SelectMethod             *widget.Select
	MethodDescriptionDisplay *widget.Label
	NotShowResultCheckbox    *widget.Check
}
