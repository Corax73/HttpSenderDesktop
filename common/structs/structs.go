package common

import (
	"encoding/json"

	"fyne.io/fyne/v2/widget"
)

type CustomResponse struct {
	Data         json.RawMessage `json:"data"`
	RepeatNumber int             `json:"repeat_number"`
}

type State struct {
	ResponseData string
}

type Sender struct {
	State
	DisplayEntry  *widget.Entry
	DisplayRepeat *widget.Label
}

func (sender *Sender) GetResponseData() *string {
	return &sender.ResponseData
}

func (sender *Sender) SetResponseData(val string) {
	sender.ResponseData = val
}

func (sender *Sender) ClearResultBtnHandler() *widget.Button {
	return widget.NewButton("Clear result", func() {
		sender.DisplayEntry.SetText("")
		sender.SetResponseData("")
	})
}

func (sender *Sender) ShowResp(data *string) {
	sender.DisplayEntry.SetText(*data)
}
