package main

import (
	"httpSenderDesktop/customTheme"
	"httpSenderDesktop/httpSender"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	httpSenderApp := app.New()
	httpSenderApp.Settings().SetTheme(customTheme.NewCustomTheme())
	window := httpSenderApp.NewWindow("HttpSender")

	btnExit := widget.NewButton("Exit", func() {
		httpSenderApp.Quit()
	})

	httpSender := httpSender.HttpSender{
		UrlEntry:               widget.NewEntry(),
		DisplayEntry:           widget.NewEntry(),
		ParamsEntry:            widget.NewEntry(),
		RepeatEntry:            widget.NewEntry(),
		DelayEntry:             widget.NewEntry(),
		DisplayRepeat:          widget.NewLabel("Repeat №"),
		BasicAuthUsernameEntry: widget.NewEntry(),
		BasicAuthPasswordEntry: widget.NewPasswordEntry(),
		HeadersEntry:           widget.NewEntry(),
	}
	httpSender.ResetState()
	httpSender.ParamsEntry.MultiLine = true
	httpSender.ParamsEntry.SetPlaceHolder("Enter parameters by JSON")
	httpSender.SendBtn = httpSender.SendBtnHandler()
	httpSender.UrlEntry.SetPlaceHolder("Enter the address bar for the request")
	httpSender.RepeatEntry.SetPlaceHolder("Enter the number of repetitions, default is 1")
	httpSender.DelayEntry.SetPlaceHolder("Enter delay, default is 200 milliseconds")
	httpSender.ScrollContainer = httpSender.GetScrollDisplay()
	httpSender.ClearResultBtn = httpSender.ClearResultBtnHandler()
	httpSender.CopyBtn = httpSender.CopyBtnHandler()
	httpSender.SelectMethod = httpSender.GetSelectMethod()
	httpSender.ClearParametersBtn = httpSender.ClearParametersBtnHandler()
	httpSender.SaveResultBtn = httpSender.SaveResultBtnHandler(window)
	httpSender.NotShowResultCheckbox = httpSender.NotShowResultCheckboxHandler()
	httpSender.SetBasicAuthBtn = httpSender.SetBasicAuthBtnHandler(window)
	httpSender.SetCookieBtn = httpSender.SetDynamicCookieFormBtnHandler(window)
	httpSender.HeadersEntry.MultiLine = true
	httpSender.HeadersEntry.SetPlaceHolder("Enter headers by JSON, default is 'Content-Type', 'application/json'" +
		" and 'User-Agent', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36'")
	httpSender.SaveStateBtn = httpSender.SaveStateBtnHandler(window)
	httpSender.LoadStateBtn = httpSender.LoadStateBtnHandler(window)
	httpSender.Load()

	content := container.NewGridWithColumns(
		1,
		container.NewVBox(
			httpSender.UrlEntry,
			httpSender.HeadersEntry,
			httpSender.ParamsEntry,
		),
		container.NewVBox(
			container.NewVBox(
				container.NewGridWithColumns(
					3,
					httpSender.RepeatEntry,
					httpSender.DelayEntry,
					httpSender.SelectMethod,
				),
				container.NewGridWithColumns(
					4,
					httpSender.SetBasicAuthBtn,
					httpSender.SetCookieBtn,
					httpSender.ClearParametersBtn,
					httpSender.SendBtn,
				),
				container.NewGridWithColumns(
					2,
					httpSender.SaveStateBtn,
					httpSender.LoadStateBtn,
				),
			),
			container.NewGridWithColumns(
				2,
				httpSender.DisplayRepeat,
				httpSender.NotShowResultCheckbox,
			),
		),
		container.NewBorder(
			nil,
			btnExit,
			nil,
			container.NewBorder(
				httpSender.ClearResultBtn,
				httpSender.SaveResultBtn,
				nil,
				nil,
				httpSender.CopyBtn,
			),
			httpSender.ScrollContainer,
		),
	)
	window.SetContent(content)
	window.CenterOnScreen()
	window.Resize(fyne.NewSize(950, 600))
	window.ShowAndRun()
}
