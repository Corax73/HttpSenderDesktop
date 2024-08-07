package httpSender

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type State struct {
	Method string
}

type HttpSender struct {
	State
	Input, Display, Params *widget.Entry
	ScrollContainer        *container.Scroll
	SendBtn                *widget.Button
}

func (httpSender *HttpSender) SendBtnHandler() *widget.Button {
	return widget.NewButton("Send", func() {
		if httpSender.Input.Text != "" && httpSender.Method != "" {
			resp, err := httpSender.SendByMethod()
			if err == nil {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					defer resp.Body.Close()
					var prettyJSON bytes.Buffer
					if err := json.Indent(&prettyJSON, []byte(body), "", "    "); err == nil {
						httpSender.showResp(prettyJSON.String())
					} else {
						httpSender.showResp(err.Error())
					}
				} else {
					httpSender.showResp(err.Error())
				}
			} else {
				httpSender.showResp(err.Error())
			}
		} else {
			httpSender.showResp("Enter the request string")
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
		responseBody := httpSender.GetParams()
		resp, err = http.Post(httpSender.Input.Text, "application/json", responseBody)
		if err != nil {
			httpSender.showResp(err.Error())
		}
	default:
		return resp, err
	}
	return resp, err
}

func (httpSender *HttpSender) showResp(data string) {
	httpSender.Display.SetText(data)
}

func (httpSender *HttpSender) GetScrollDisplay() *container.Scroll {
	return container.NewVScroll(container.NewGridWithRows(
		1,
		httpSender.Display,
	))
}

func (httpSender *HttpSender) GetSelectMethod() *widget.Select {
	return widget.NewSelect([]string{"GET", "POST"}, func(value string) {
		httpSender.Method = value
	})
}

func (httpSender *HttpSender) GetParams() *bytes.Buffer {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(httpSender.Params.Text), &data)
	if err != nil {
		httpSender.showResp(err.Error())
	}
	postBody, _ := json.Marshal(data)
	responseBody := bytes.NewBuffer(postBody)
	return responseBody
}
