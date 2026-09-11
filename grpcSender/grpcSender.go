package grpcSender

import (
	"context"
	"encoding/json"
	"fmt"
	common "httpSenderDesktop/common/structs"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
)

type GrpcSender struct {
	state
	UrlEntry, FullServiceNameEntry, DisplayEntry, ParamsEntry, RepeatEntry, DelayEntry *widget.Entry
	ScrollContainer                                                                    *container.Scroll
	ParseMethodsBtn, SendBtn, ClearResultBtn,
	ClearParametersBtn, CopyMethodDescriptionBtn, ResultCopyBtnHandlerBtn,
	SaveResultBtn *widget.Button
	SelectMethod             *widget.Select
	MethodDescriptionDisplay *widget.Label
	NotShowResultCheckbox    *widget.Check
}

func (state *state) ResetState() {
	state.Url, state.FullServiceName, state.Params, state.Method, state.ResponseData = "", "", "", "", ""
	state.MethodsDescription = make([]*methodDescription, 0)
	state.NotShowResult = false
	state.Repeat, state.Delay = 1, 200
	state.Responses = make([]*common.CustomResponse, 0)
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, refClient, err := grpcSender.getGrpcClient(ctx)
		if err != nil {
			errStr := fmt.Errorf("Failed to connect to the server: %v", err).Error()
			grpcSender.showResp(&errStr)
			return
		}
		defer conn.Close()
		defer refClient.Reset()

		grpcSender.getRepeat()
		repetitionChans := make([]chan *rpcResponseData, grpcSender.Repeat)
		for i := 0; i < grpcSender.Repeat; i++ {
			repetitionChans[i] = make(chan *rpcResponseData, 1)
		}
		var wg sync.WaitGroup
		defer wg.Wait()
		grpcSender.switchingAvailability(false)
		grpcSender.getDelay()
		for i := 0; i < grpcSender.Repeat; i++ {
			wg.Add(1)
			go func(counter int) {
				defer wg.Done()
				reqCtx, reqCancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer reqCancel()
				grpcSender.executeRpcMethod(reqCtx, conn, refClient, repetitionChans[counter], counter+1)
			}(i)
			if grpcSender.Repeat > 1 {
				time.Sleep(time.Duration(grpcSender.Delay) * time.Millisecond)
			}
		}
		grpcSender.DisplayEntry.SetPlaceHolder("Reading responses to requests...")
		for _, ch := range repetitionChans {
			resp := <-ch
			if json.Valid(resp.DataBytes) {
				grpcSender.Responses = append(grpcSender.Responses,
					&common.CustomResponse{Data: json.RawMessage(resp.DataBytes), RepeatNumber: resp.RepeatNumber},
				)
			} else {
				if resp.Error != nil {
					grpcSender.Responses = append(
						grpcSender.Responses,
						&common.CustomResponse{
							Data: json.RawMessage(
								strings.ReplaceAll(
									strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(resp.Error.Error(), `"`, ""), "\r", ""), "\n", ""),
									":", " "),
							),
							RepeatNumber: resp.RepeatNumber,
						},
					)
				} else {
					errMsg := fmt.Sprintf(
						`{"error": "Invalid JSON response", "body": %q}`,
						strings.ReplaceAll(
							strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(resp.DataBytes), `"`, ""), "\r", ""), "\n", ""),
							":", " "),
					)
					grpcSender.Responses = append(grpcSender.Responses, &common.CustomResponse{Data: json.RawMessage(errMsg), RepeatNumber: resp.RepeatNumber})
				}
			}
			close(ch)
		}

		repetitionChans = nil
		bytesData, err := json.MarshalIndent(grpcSender.Responses, "", " ")
		if err != nil {
			for _, item := range grpcSender.Responses {
				if len(item.Data) == 0 {
					item.Data = json.RawMessage(`{"error": "Empty response"}`)
					continue
				}

				if !json.Valid(item.Data) {
					item.Data = json.RawMessage(fmt.Sprintf(`"error": "%s"`, string(item.Data)))
				}
			}
			bytesData, _ = json.MarshalIndent(grpcSender.Responses, "", " ")
		}
		grpcSender.ResponseData = string(bytesData)
		grpcSender.Responses = nil
		if !grpcSender.NotShowResult {
			grpcSender.showResp(&grpcSender.ResponseData)
		}
		grpcSender.switchingAvailability(true)
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

func (grpcSender *GrpcSender) MethodDescriptionCopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy description to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			errResp := err.Error()
			grpcSender.showResp(&errResp)
		}
		clipboard.Write(clipboard.FmtText, []byte(grpcSender.MethodDescriptionDisplay.Text))
	})
}

func (grpcSender *GrpcSender) ClearParametersBtnHandler() *widget.Button {
	return widget.NewButton("Clear all parameters", func() {
		grpcSender.UrlEntry.SetText("")
		grpcSender.ParamsEntry.SetText("")
		grpcSender.FullServiceNameEntry.SetText("")
		grpcSender.MethodDescriptionDisplay.SetText("")
		grpcSender.SelectMethod.Selected = "Select method"
		grpcSender.SelectMethod.Refresh()
		grpcSender.DelayEntry.SetText("")
		grpcSender.RepeatEntry.SetText("")
		grpcSender.ResetState()
	})
}

func (grpcSender *GrpcSender) ClearResultBtnHandler() *widget.Button {
	return widget.NewButton("Clear result", func() {
		grpcSender.DisplayEntry.SetText("")
		grpcSender.ResponseData = ""
	})
}

func (grpcSender *GrpcSender) ResultCopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy result to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			errResp := err.Error()
			grpcSender.showResp(&errResp)
		}
		clipboard.Write(clipboard.FmtText, []byte(grpcSender.ResponseData))
	})
}

func (grpcSender *GrpcSender) SaveResultBtnHandler(appWindow fyne.Window) *widget.Button {
	return widget.NewButton("Save result to file", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				_, err := writer.Write([]byte(grpcSender.ResponseData))
				if err != nil {
					dialog.ShowError(err, appWindow)
				}
			}
		}, appWindow)
	})
}

func (grpcSender *GrpcSender) NotShowResultCheckboxHandler() *widget.Check {
	return widget.NewCheck("Not show result(reduces the load)", func(value bool) {
		grpcSender.NotShowResult = value
	})
}

func (grpcSender *GrpcSender) getRepeat() {
	if grpcSender.RepeatEntry.Text != "" {
		number, err := strconv.Atoi(grpcSender.RepeatEntry.Text)
		if err == nil {
			grpcSender.Repeat = number
		}
	}
}

func (grpcSender *GrpcSender) getDelay() {
	if grpcSender.DelayEntry.Text != "" {
		number, err := strconv.Atoi(grpcSender.DelayEntry.Text)
		if err == nil {
			grpcSender.Delay = number
		}
	}
}

func (grpcSender *GrpcSender) switchingAvailability(isOn bool) {
	if isOn {
		grpcSender.UrlEntry.Enable()
		grpcSender.FullServiceNameEntry.Enable()
		grpcSender.DisplayEntry.Enable()
		grpcSender.ParamsEntry.Enable()
		grpcSender.RepeatEntry.Enable()
		grpcSender.DelayEntry.Enable()
		grpcSender.SendBtn.Enable()
		grpcSender.SendBtn.SetText("Send")
		grpcSender.ClearResultBtn.Enable()
		grpcSender.ClearParametersBtn.Enable()
		grpcSender.SaveResultBtn.Enable()
		grpcSender.SelectMethod.Enable()
		grpcSender.NotShowResultCheckbox.Enable()
		grpcSender.CopyMethodDescriptionBtn.Enable()
		grpcSender.ResultCopyBtnHandlerBtn.Enable()
		grpcSender.ParseMethodsBtn.Enable()
	} else {
		grpcSender.UrlEntry.Disable()
		grpcSender.FullServiceNameEntry.Disable()
		grpcSender.DisplayEntry.Disable()
		grpcSender.ParamsEntry.Disable()
		grpcSender.RepeatEntry.Disable()
		grpcSender.DelayEntry.Disable()
		grpcSender.SendBtn.Disable()
		grpcSender.ClearResultBtn.Disable()
		grpcSender.ClearParametersBtn.Disable()
		grpcSender.SaveResultBtn.Disable()
		grpcSender.SelectMethod.Disable()
		grpcSender.NotShowResultCheckbox.Disable()
		grpcSender.SendBtn.SetText("Sending...")
		grpcSender.CopyMethodDescriptionBtn.Disable()
		grpcSender.ResultCopyBtnHandlerBtn.Disable()
		grpcSender.ParseMethodsBtn.Disable()
	}
}
