package common

import (
	"encoding/json"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
)

type CustomResponse struct {
	Data         json.RawMessage `json:"data"`
	RepeatNumber int             `json:"repeat_number"`
}

type Sender struct {
	CustomSender CustomSender
	State
	DisplayEntry, UrlEntry, ParamsEntry, RepeatEntry, DelayEntry                           *widget.Entry
	DisplayRepeat                                                                          *widget.Label
	SendBtn, ClearResultBtn, ClearParametersBtn, SaveResultBtn, SaveStateBtn, LoadStateBtn *widget.Button
}

type CustomSender interface {
	SaveState(title string)
	GetStatesSelect() *widget.Select
	UseStateByTitle(title string)
}

func (sender *Sender) GetResponses() []*CustomResponse {
	return sender.Responses
}

func (sender *Sender) SetResponses(val []*CustomResponse) {
	sender.Responses = val
}

func (sender *Sender) GetUrl() *string {
	return &sender.Url
}

func (sender *Sender) SetUrl(val string) {
	sender.Url = val
}

func (sender *Sender) GetParams() *string {
	return &sender.Params
}

func (sender *Sender) SetParams(val string) {
	sender.Params = val
}

func (sender *Sender) GetMethod() *string {
	return &sender.Method
}

func (sender *Sender) SetMethod(val string) {
	sender.Method = val
}

func (sender *Sender) GetResponseData() *string {
	return &sender.ResponseData
}

func (sender *Sender) SetResponseData(val string) {
	sender.ResponseData = val
}

func (sender *Sender) GetNotShowResult() bool {
	return sender.NotShowResult
}

func (sender *Sender) SetNotShowResult(val bool) {
	sender.NotShowResult = val
}

func (sender *Sender) GetRepeat() int {
	return sender.Repeat
}

func (sender *Sender) SetRepeat(val int) {
	sender.Repeat = val
}

func (sender *Sender) GetDelay() int {
	return sender.Delay
}

func (sender *Sender) SetDelay(val int) {
	sender.Delay = val
}

func (sender *Sender) ClearResultBtnHandler() *widget.Button {
	return widget.NewButton("Clear result", func() {
		sender.DisplayEntry.SetText("")
		sender.SetResponseData("")
	})
}

func (sender *Sender) SaveResultBtnHandler(appWindow fyne.Window) *widget.Button {
	return widget.NewButton("Save result to file", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				_, err := writer.Write([]byte(*sender.GetResponseData()))
				if err != nil {
					dialog.ShowError(err, appWindow)
				}
			}
		}, appWindow)
	})
}

func (sender *Sender) ParseRepeat() {
	if sender.RepeatEntry.Text != "" {
		number, err := strconv.Atoi(sender.RepeatEntry.Text)
		if err == nil {
			sender.Repeat = number
		}
	}
}

func (sender *Sender) ParseDelay() {
	if sender.DelayEntry.Text != "" {
		number, err := strconv.Atoi(sender.DelayEntry.Text)
		if err == nil {
			sender.Delay = number
		}
	}
}

func (sender *Sender) ShowResp(data *string) {
	sender.DisplayEntry.SetText(*data)
}

func (sender *Sender) SaveStateBtnHandler(appWindow fyne.Window) *widget.Button {
	return widget.NewButton("Save state for reuse", func() {
		stateTitleForm := widget.NewForm()
		titleEntry := widget.NewEntry()
		stateTitleForm.Append("State title", titleEntry)
		dialogContent := container.NewScroll(
			container.NewVBox(
				stateTitleForm,
			),
		)

		dlg := dialog.NewCustomConfirm(
			"Set title for this state",
			"Submit",
			"Cancel",
			dialogContent,
			func(ok bool) {
				if ok && titleEntry.Text != "" {
					sender.CustomSender.SaveState(titleEntry.Text)
				}
			},
			appWindow,
		)

		dlg.Resize(fyne.NewSize(300, 170))
		dlg.Show()
	})
}

func (sender *Sender) LoadStateBtnHandler(appWindow fyne.Window) *widget.Button {
	return widget.NewButton("Load state for reuse", func() {

		selectWidget := sender.CustomSender.GetStatesSelect()

		dialogContent := container.NewScroll(
			container.NewVBox(
				selectWidget,
			),
		)
		dlg := dialog.NewCustomConfirm(
			"Set title for this state",
			"Submit",
			"Cancel",
			dialogContent,
			func(ok bool) {
				if ok {
					sender.CustomSender.UseStateByTitle(selectWidget.Selected)
				}
			},
			appWindow,
		)

		dlg.Resize(fyne.NewSize(300, 170))
		dlg.Show()
	})
}

func (sender *Sender) GetScrollDisplay() *container.Scroll {
	return container.NewVScroll(container.NewGridWithRows(
		1,
		sender.DisplayEntry,
	))
}

func (sender *Sender) ResultCopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy result to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			errResp := err.Error()
			sender.ShowResp(&errResp)
		}
		clipboard.Write(clipboard.FmtText, []byte(*sender.GetResponseData()))
	})
}

func (sender *Sender) NotShowResultCheckboxHandler() *widget.Check {
	return widget.NewCheck("Not show result(reduces the load)", func(value bool) {
		sender.SetNotShowResult(value)
	})
}
