package httpSender

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.design/x/clipboard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const repetitionNumberStub = 0

type State struct {
	Method        string
	Repeat, Delay int
}

func (state *State) ResetState() {
	state.Method = ""
	state.Repeat, state.Delay = 1, 200
}

type HttpSender struct {
	State
	Input, Display, Params, RepeatEntry, DelayEntry                  *widget.Entry
	ScrollContainer                                                  *container.Scroll
	SendBtn, ClearResultBtn, CopyBtn, ClearParametersBtn, SaveResultBtn *widget.Button
	DisplayRepeat                                                    *widget.Label
	SelectMethod                                                     *widget.Select
}

func (httpSender *HttpSender) SendBtnHandler() *widget.Button {
	return widget.NewButton("Send", func() {
		if httpSender.Input.Text != "" && httpSender.Method != "" {
			httpSender.getRepeat()
			httpSender.Display.SetText("")
			for i := 0; i < httpSender.Repeat; i++ {
				httpSender.showRepeat(i + 1)
				resp, err := httpSender.SendByMethod()
				if err == nil {
					body, err := io.ReadAll(resp.Body)
					if err == nil {
						defer resp.Body.Close()
						var prettyJSON bytes.Buffer
						if err := json.Indent(&prettyJSON, []byte(body), "", "    "); err == nil {
							httpSender.showResp(prettyJSON.String(), i+1)
						} else {
							httpSender.showResp(err.Error(), i+1)
						}
					} else {
						httpSender.showResp(err.Error(), i+1)
					}
				} else {
					httpSender.showResp(err.Error(), i+1)
				}
				if httpSender.Repeat > 1 {
					httpSender.getDelay()
					time.Sleep(time.Duration(httpSender.Delay) * time.Millisecond)
				}
			}
		} else {
			httpSender.showResp("Enter the request string", repetitionNumberStub)
		}
	})
}

func (httpSender *HttpSender) SendByMethod() (*http.Response, error) {
	var resp *http.Response
	var err error
	switch httpSender.Method {
	case "GET":
		resp, err = http.Get(httpSender.Input.Text)
	case "POST":
		responseBody := httpSender.getParams()
		resp, err = http.Post(httpSender.Input.Text, "application/json", responseBody)
	case "DELETE":
		client := &http.Client{}
		var req *http.Request
		req, err = http.NewRequest(http.MethodDelete, httpSender.Input.Text, nil)
		if err == nil {
			resp, err = client.Do(req)
		}
	case "PUT":
		responseBody := httpSender.getParams()
		client := &http.Client{}
		var req *http.Request
		req, err = http.NewRequest(http.MethodPut, httpSender.Input.Text, responseBody)
		req.Header.Set("Content-Type", "application/json")
		if err == nil {
			resp, err = client.Do(req)
		}
	default:
		return resp, err
	}
	return resp, err
}

func (httpSender *HttpSender) showResp(data string, repeatNumber int) {
	var strBuilder strings.Builder
	if httpSender.Repeat > 1 {
		if httpSender.Display.Text != "" {
			strBuilder.WriteString("[")
			data := strings.Trim(httpSender.Display.Text, "[")
			data = strings.Trim(data, "]")
			strBuilder.WriteString(data)
			strBuilder.WriteString(",")
		}
		strBuilder.WriteString("{")
		strBuilder.WriteString("\n")
		strBuilder.WriteString("\"repeat_number\": ")
		strBuilder.WriteString(strconv.Itoa(repeatNumber))
		strBuilder.WriteString(",")
		strBuilder.WriteString("\n")
		strBuilder.WriteString("\"data\": \n")
		strBuilder.WriteString(data)
		strBuilder.WriteString("}")
		strBuilder.WriteString("\n")
		strBuilder.WriteString("]")
		httpSender.Display.SetText(strBuilder.String())
		strBuilder.Reset()
	} else {
		strBuilder.WriteString("[")
		strBuilder.WriteString("{")
		httpSender.Display.SetText(data)
		strBuilder.WriteString("}")
		strBuilder.WriteString("]")
	}
}

func (httpSender *HttpSender) showRepeat(repeatNumber int) {
	var strBuilder strings.Builder
	strBuilder.WriteString("Repeat №")
	strBuilder.WriteString(strconv.Itoa(repeatNumber))
	httpSender.DisplayRepeat.SetText(strBuilder.String())
	strBuilder.Reset()
}

func (httpSender *HttpSender) GetScrollDisplay() *container.Scroll {
	return container.NewVScroll(container.NewGridWithRows(
		1,
		httpSender.Display,
	))
}

func (httpSender *HttpSender) GetSelectMethod() *widget.Select {
	resp := widget.NewSelect([]string{"GET", "POST", "DELETE", "PUT"}, func(value string) {
		httpSender.Method = value
	})
	resp.PlaceHolder = "Select method"
	return resp
}

func (httpSender *HttpSender) getParams() *bytes.Buffer {
	data := make(map[string]interface{})
	str := httpSender.Params.Text
	if str == "" {
		str = "{}"
	}
	err := json.Unmarshal([]byte(str), &data)
	if err != nil {
		httpSender.showResp(err.Error(), repetitionNumberStub)
	}
	postBody, _ := json.Marshal(data)
	responseBody := bytes.NewBuffer(postBody)
	return responseBody
}

func (httpSender *HttpSender) getRepeat() {
	if httpSender.RepeatEntry.Text != "" {
		number, err := strconv.Atoi(httpSender.RepeatEntry.Text)
		if err == nil {
			httpSender.Repeat = number
		}
	}
}

func (httpSender *HttpSender) getDelay() {
	if httpSender.DelayEntry.Text != "" {
		number, err := strconv.Atoi(httpSender.DelayEntry.Text)
		if err == nil {
			httpSender.Delay = number
		}
	}
}

func (httpSender *HttpSender) ClearResultBtnHandler() *widget.Button {
	return widget.NewButton("Clear result", func() {
		httpSender.Display.SetText("")
	})
}

func (httpSender *HttpSender) CopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			httpSender.showResp(err.Error(), repetitionNumberStub)
		}
		clipboard.Write(clipboard.FmtText, []byte(httpSender.Display.Text))
	})
}

func (httpSender *HttpSender) ClearParametersBtnHandler() *widget.Button {
	return widget.NewButton("Clear all parameters", func() {
		httpSender.Input.SetText("")
		httpSender.Params.SetText("")
		httpSender.RepeatEntry.SetText("")
		httpSender.DelayEntry.SetText("")
		httpSender.SelectMethod.Selected = "Select method"
		httpSender.ResetState()
	})
}

func (httpSender *HttpSender) SaveResultBtnHandler(appWindow fyne.Window) *widget.Button {
	return widget.NewButton("Save result to file", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				_, err := writer.Write([]byte(httpSender.Display.Text))
				if err != nil {
					dialog.ShowError(err, appWindow)
				}
			}
		}, appWindow)
	})
}
