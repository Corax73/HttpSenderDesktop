package httpSender

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type HttpSender struct {
	Input           *widget.Entry
	Display         *widget.Label
	ScrollContainer *container.Scroll
	SendBtn         *widget.Button
}

func (httpSender *HttpSender) SendBtnHandler() *widget.Button {
	return widget.NewButton("Send", func() {
		if httpSender.Input.Text != "" {
			resp, err := http.Get(httpSender.Input.Text)
			if err == nil {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
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

func (httpSender *HttpSender) showResp(data string) {
	httpSender.Display.SetText(data)
}

func (httpSender *HttpSender) GetScrollDisplay() *container.Scroll {
	return container.NewVScroll(container.NewGridWithRows(
		1,
		httpSender.Display,
	))
}
