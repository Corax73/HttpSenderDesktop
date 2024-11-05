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
		Input:         widget.NewEntry(),
		Display:       widget.NewEntry(),
		Params:        widget.NewEntry(),
		RepeatEntry:   widget.NewEntry(),
		DelayEntry:    widget.NewEntry(),
		DisplayRepeat: widget.NewLabel("Repeat №"),
	}
	httpSender.ResetState()
	httpSender.Params.MultiLine = true
	httpSender.Params.SetPlaceHolder("Enter parameters by JSON")
	httpSender.SendBtn = httpSender.SendBtnHandler()
	httpSender.Input.SetPlaceHolder("Enter the address bar for the request")
	httpSender.RepeatEntry.SetPlaceHolder("Enter the number of repetitions, default is 1")
	httpSender.DelayEntry.SetPlaceHolder("Enter delay, default is 200 milliseconds")
	httpSender.ScrollContainer = httpSender.GetScrollDisplay()
	httpSender.ClearResultBtn = httpSender.ClearResultBtnHandler()
	httpSender.CopyBtn = httpSender.CopyBtnHandler()
	httpSender.SelectMethod = httpSender.GetSelectMethod()
	httpSender.ClearParametersBtn = httpSender.ClearParametersBtnHandler()
	httpSender.SaveResultBtn = httpSender.SaveResultBtnHandler(window)
	httpSender.NotShowResultCheckbox = httpSender.NotShowResultCheckboxHandler()

	content := container.NewGridWithColumns(
		1,
		container.NewBorder(
			httpSender.Input,
			nil,
			nil,
			nil,
			httpSender.Params,
		),
		container.NewGridWithRows(
			3,
			container.NewBorder(
				nil,
				nil,
				nil,
				nil,
				container.NewGridWithColumns(
					2,
					httpSender.RepeatEntry,
					httpSender.DelayEntry,
				),
			),
			container.NewBorder(
				nil,
				nil,
				nil,
				nil,
				container.NewGridWithColumns(
					3,
					httpSender.SelectMethod,
					httpSender.SendBtn,
					httpSender.ClearParametersBtn,
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
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}
