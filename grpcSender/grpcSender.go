package grpcSender

import (
	"encoding/json"
	"fmt"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
)

type State struct {
	Url, FullServiceName, Params, Method, ResponseData string
	MethodsDescription                                 []*methodDescription
}

func (state *State) ResetState() {
	state.Url, state.FullServiceName, state.Params, state.Method, state.ResponseData = "", "", "", "", ""
	state.MethodsDescription = make([]*methodDescription, 0)
}

type GrpcSender struct {
	State
	UrlEntry, FullServiceNameEntry, DisplayEntry, ParamsEntry                                       *widget.Entry
	ScrollContainer                                                                                 *container.Scroll
	ParseMethodsBtn, SendBtn, ClearResultBtn, CopyBtn, ClearParametersBtn, CopyMethodDescriptionBtn *widget.Button
	SelectMethod                                                                                    *widget.Select
	MethodDescriptionDisplay                                                                        *widget.Label
}

func (grpcSender *GrpcSender) ParseMethodsBtnHandler() *widget.Button {
	return widget.NewButton("Parse methods", func() {
		grpcSender.Url = grpcSender.UrlEntry.Text
		grpcSender.FullServiceName = grpcSender.FullServiceNameEntry.Text
		grpcSender.Method = grpcSender.SelectMethod.Selected
		if grpcSender.Url == "" || grpcSender.FullServiceName == "" {
			errStr := "Check server and service name"
			grpcSender.showResp(&errStr)
			return
		}
		list, err := grpcSender.parseServerMethods()
		if err != nil {
			errStr := err.Error()
			grpcSender.showResp(&errStr)
			return
		}
		listLength := len(*list)
		if listLength > 0 {
			methodNames := make([]string, 0, listLength)
			for _, m := range *list {
				methodNames = append(methodNames, m.GetName())

				methodDesc := m.GetInputType()
				description := methodDescription{Name: methodDesc.GetName(), MethodName: m.GetName()}
				for _, f := range methodDesc.GetFields() {
					field := fieldDescription{Name: f.GetJSONName(), Type: strings.ToLower(strings.ReplaceAll((f.GetType().String()), "TYPE_", ""))}
					description.Fields = append(description.Fields, &field)
				}
				grpcSender.MethodsDescription = append(grpcSender.MethodsDescription, &description)
			}
			grpcSender.SelectMethod.Options = methodNames
			grpcSender.SelectMethod.Enable()
		}
	})
}

func (grpcSender *GrpcSender) SendBtnHandler() *widget.Button {
	return widget.NewButton("Send", func() {
		grpcSender.Params = grpcSender.ParamsEntry.Text
		if grpcSender.Url == "" || grpcSender.FullServiceName == "" || grpcSender.Method == "" || grpcSender.Params == "" {
			errStr := "Check server, service name or method"
			grpcSender.showResp(&errStr)
			return
		}
		resp := grpcSender.executeRpcMethod()
		if resp.Error != nil {
			errStr := resp.Error.Error()
			grpcSender.showResp(&errStr)
			return
		}
		strData := string(resp.DataBytes)
		grpcSender.showResp(&strData)
	})
}

func (grpcSender *GrpcSender) GetScrollDisplay() *container.Scroll {
	return container.NewVScroll(container.NewGridWithRows(
		1,
		grpcSender.DisplayEntry,
	))
}

func (grpcSender *GrpcSender) GetSelectMethod() *widget.Select {
	resp := widget.NewSelect([]string{}, func(value string) {
		grpcSender.Method = value
		if len(grpcSender.MethodsDescription) > 0 {
			for _, v := range grpcSender.MethodsDescription {
				if v.MethodName == value {
					jsonData, err := json.Marshal(v.Fields)
					if err != nil {
						grpcSender.MethodDescriptionDisplay.SetText(fmt.Sprintf("Error marshaling to JSON: %v", err))
						return
					}
					grpcSender.MethodDescriptionDisplay.SetText(string(jsonData))
					return
				}
			}
		}
	})
	resp.PlaceHolder = "Select method"
	resp.Disable()
	return resp
}

func (grpcSender *GrpcSender) showResp(data *string) {
	grpcSender.DisplayEntry.SetText(*data)
}

func (grpcSender *GrpcSender) CopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			errResp := err.Error()
			grpcSender.showResp(&errResp)
		}
		clipboard.Write(clipboard.FmtText, []byte(grpcSender.MethodDescriptionDisplay.Text))
	})
}
