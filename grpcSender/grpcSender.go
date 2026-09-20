package grpcSender

import (
	"context"
	"encoding/json"
	"fmt"
	common "httpSenderDesktop/common/sender"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
)

func (grpcSender *GrpcSender) ResetState() {
	grpcSender.FullServiceName = ""
	grpcSender.SetUrl("")
	grpcSender.SetParams("")
	grpcSender.SetMethod("")
	grpcSender.SetNotShowResult(false)
	grpcSender.SetResponseData("")
	grpcSender.SetRepeat(1)
	grpcSender.SetDelay(200)
	grpcSender.SetResponses(make([]*common.CustomResponse, 0))
}

func (grpcSender *GrpcSender) ParseMethodsBtnHandler() *widget.Button {
	return widget.NewButton("Parse methods", func() {
		grpcSender.SetUrl(grpcSender.UrlEntry.Text)
		grpcSender.FullServiceName = grpcSender.FullServiceNameEntry.Text
		grpcSender.SetMethod(grpcSender.SelectMethod.Selected)
		if *grpcSender.GetUrl() == "" || grpcSender.FullServiceName == "" {
			errStr := "Check server and service name"
			grpcSender.ShowResp(&errStr)
			return
		}
		list, err := grpcSender.parseServerMethods()
		if err != nil {
			errStr := err.Error()
			grpcSender.ShowResp(&errStr)
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
		grpcSender.SetParams(grpcSender.ParamsEntry.Text)
		if *grpcSender.GetUrl() == "" || grpcSender.FullServiceName == "" || *grpcSender.GetMethod() == "" || *grpcSender.GetParams() == "" {
			errStr := "Check server, service name or method"
			grpcSender.ShowResp(&errStr)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, refClient, err := grpcSender.getGrpcClient(ctx)
		if err != nil {
			errStr := fmt.Errorf("Failed to connect to the server: %v", err).Error()
			grpcSender.ShowResp(&errStr)
			return
		}
		defer conn.Close()
		defer refClient.Reset()

		grpcSender.ParseRepeat()
		repetitionChans := make([]chan *rpcResponseData, grpcSender.GetRepeat())
		for i := 0; i < grpcSender.GetRepeat(); i++ {
			repetitionChans[i] = make(chan *rpcResponseData, 1)
		}
		var wg sync.WaitGroup
		defer wg.Wait()
		grpcSender.switchingAvailability(false)
		grpcSender.ParseDelay()
		for i := 0; i < grpcSender.GetRepeat(); i++ {
			wg.Add(1)
			go func(counter int) {
				defer wg.Done()
				reqCtx, reqCancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer reqCancel()
				grpcSender.executeRpcMethod(reqCtx, conn, refClient, repetitionChans[counter], counter+1)
			}(i)
			if grpcSender.GetRepeat() > 1 {
				time.Sleep(time.Duration(grpcSender.GetDelay()) * time.Millisecond)
			}
		}
		grpcSender.DisplayEntry.SetPlaceHolder("Reading responses to requests...")
		for _, ch := range repetitionChans {
			resp := <-ch
			if json.Valid(resp.DataBytes) {
				grpcSender.SetResponses(
					append(
						grpcSender.GetResponses(),
						&common.CustomResponse{Data: json.RawMessage(resp.DataBytes), RepeatNumber: resp.RepeatNumber},
					))
			} else {
				if resp.Error != nil {
					grpcSender.SetResponses(
						append(
							grpcSender.GetResponses(),
							&common.CustomResponse{
								Data: json.RawMessage(
									strings.ReplaceAll(
										strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(resp.Error.Error(), `"`, ""), "\r", ""), "\n", ""),
										":", " "),
								),
								RepeatNumber: resp.RepeatNumber,
							},
						))
				} else {
					errMsg := fmt.Sprintf(
						`{"error": "Invalid JSON response", "body": %q}`,
						strings.ReplaceAll(
							strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(resp.DataBytes), `"`, ""), "\r", ""), "\n", ""),
							":", " "),
					)
					grpcSender.SetResponses(
						append(
							grpcSender.GetResponses(),
							&common.CustomResponse{Data: json.RawMessage(errMsg), RepeatNumber: resp.RepeatNumber},
						))
				}
			}
			close(ch)
		}

		repetitionChans = nil
		bytesData, err := json.MarshalIndent(grpcSender.GetResponses(), "", " ")
		if err != nil {
			for _, item := range grpcSender.GetResponses() {
				if len(item.Data) == 0 {
					item.Data = json.RawMessage(`{"error": "Empty response"}`)
					continue
				}

				if !json.Valid(item.Data) {
					item.Data = json.RawMessage(fmt.Sprintf(`"error": "%s"`, string(item.Data)))
				}
			}
			bytesData, _ = json.MarshalIndent(grpcSender.GetResponses(), "", " ")
		}
		grpcSender.SetResponseData(string(bytesData))
		grpcSender.SetResponses(nil)
		if !grpcSender.GetNotShowResult() {
			grpcSender.ShowResp(grpcSender.GetResponseData())
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
		grpcSender.SetMethod(value)
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

func (grpcSender *GrpcSender) MethodDescriptionCopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy description to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			errResp := err.Error()
			grpcSender.ShowResp(&errResp)
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
		grpcSender.SelectMethod.Options = nil
		grpcSender.SelectMethod.Selected = "Select method"
		grpcSender.SelectMethod.Refresh()
		grpcSender.DelayEntry.SetText("")
		grpcSender.RepeatEntry.SetText("")
		grpcSender.ResetState()
	})
}

func (grpcSender *GrpcSender) ResultCopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy result to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			errResp := err.Error()
			grpcSender.ShowResp(&errResp)
		}
		clipboard.Write(clipboard.FmtText, []byte(*grpcSender.GetResponseData()))
	})
}

func (grpcSender *GrpcSender) NotShowResultCheckboxHandler() *widget.Check {
	return widget.NewCheck("Not show result(reduces the load)", func(value bool) {
		grpcSender.SetNotShowResult(value)
	})
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

func (grpcSender *GrpcSender) SaveState(title string) {
	if title != "" {
		grpcSender.ParseRepeat()
		grpcSender.ParseDelay()
		grpcSender.stateHistory[title] = &grpcState{
			common.State{
				Url:           grpcSender.UrlEntry.Text,
				Params:        grpcSender.ParamsEntry.Text,
				Method:        *grpcSender.GetMethod(),
				ResponseData:  "",
				Repeat:        grpcSender.GetRepeat(),
				Delay:         grpcSender.GetDelay(),
				NotShowResult: grpcSender.GetNotShowResult(),
				Responses:     grpcSender.GetResponses(),
			},
			grpcSender.FullServiceNameEntry.Text,
			grpcSender.MethodsDescription,
		}
	}
}

func (grpcSender *GrpcSender) GetStatesSelect() *widget.Select {
	var keys []string
	for k := range grpcSender.stateHistory {
		keys = append(keys, k)
	}
	return widget.NewSelect(keys, func(value string) {})
}

func (grpcSender *GrpcSender) UseStateByTitle(title string) {
	state, ok := grpcSender.stateHistory[title]
	if ok {
		grpcSender.UrlEntry.SetText(state.Url)
		grpcSender.ParamsEntry.SetText(state.Params)
		grpcSender.MethodsDescription = state.MethodsDescription
		methodNames := make([]string, 0, len(state.MethodsDescription))
		for _, m := range state.MethodsDescription {
			methodNames = append(methodNames, m.MethodName)
		}
		grpcSender.SelectMethod.Options = methodNames
		grpcSender.SelectMethod.SetSelected(state.Method)
		grpcSender.RepeatEntry.SetText(strconv.Itoa(state.Repeat))
		grpcSender.DelayEntry.SetText(strconv.Itoa(state.Delay))
		grpcSender.FullServiceNameEntry.SetText(state.FullServiceName)
		grpcSender.NotShowResultCheckbox.SetChecked(state.NotShowResult)
		grpcSender.SetResponses(state.Responses)
	}
}

func (grpcSender *GrpcSender) Load() {
	grpcSender.stateHistory = make(map[string]*grpcState)
	grpcSender.MethodsDescription = make([]*methodDescription, 0)
}
