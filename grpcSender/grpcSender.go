package grpcSender

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type State struct {
	Url, FullServiceName, Params, Method, ResponseData string
}

func (state *State) ResetState() {
	state.Url, state.FullServiceName, state.Params, state.Method, state.ResponseData = "", "", "", "", ""
}

type GrpcSender struct {
	State
	UrlEntry, FullServiceNameEntry, DisplayEntry, ParamsEntry             *widget.Entry
	ScrollContainer                                                       *container.Scroll
	ParseMethodsBtn, SendBtn, ClearResultBtn, CopyBtn, ClearParametersBtn *widget.Button
	SelectMethod                                                          *widget.Select
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
	})
	resp.PlaceHolder = "Select method"
	resp.Disable()
	return resp
}

func (grpcSender *GrpcSender) showResp(data *string) {
	grpcSender.DisplayEntry.SetText(*data)
}
