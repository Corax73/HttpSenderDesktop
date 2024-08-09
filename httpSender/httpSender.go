package httpSender

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const repetitionNumberStub = 0

type State struct {
	Method        string
	Repeat, Delay int
}

type HttpSender struct {
	State
	Input, Display, Params, RepeatEntry, DelayEntry *widget.Entry
	ScrollContainer                                 *container.Scroll
	SendBtn                                         *widget.Button
	DisplayRepeat                                   *widget.Label
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
		if err != nil {
			httpSender.showResp(err.Error(), repetitionNumberStub)
		}
	default:
		return resp, err
	}
	return resp, err
}

func (httpSender *HttpSender) showResp(data string, repeatNumber int) {
	if httpSender.Repeat > 1 {
		var strBuilder strings.Builder
		strBuilder.WriteString(httpSender.Display.Text)
		strBuilder.WriteString("\n")
		strBuilder.WriteString("Repeat number: ")
		strBuilder.WriteString(strconv.Itoa(repeatNumber))
		strBuilder.WriteString("\n")
		strBuilder.WriteString("Data: \n")
		strBuilder.WriteString(data)
		httpSender.Display.SetText(strBuilder.String())
		strBuilder.Reset()
	} else {
		httpSender.Display.SetText(data)
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
	resp := widget.NewSelect([]string{"GET", "POST"}, func(value string) {
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
