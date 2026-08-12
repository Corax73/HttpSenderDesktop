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
		list, err := grpcSender.parseServerMethods(grpcSender.UrlEntry.Text, grpcSender.FullServiceNameEntry.Text)
		if err != nil {
			errStr := err.Error()
			grpcSender.showResp(&errStr)
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
