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
		Input:   widget.NewEntry(),
		Display: widget.NewEntry(),
		Params: widget.NewEntry(),
	}
	httpSender.Params.SetPlaceHolder("Enter parameters by JSON")
	httpSender.SendBtn = httpSender.SendBtnHandler()
	httpSender.Input.SetPlaceHolder("Enter the address bar for the request")
	httpSender.ScrollContainer = httpSender.GetScrollDisplay()

	content := container.NewGridWithColumns(
		1,
		container.NewGridWithRows(
			1,
			httpSender.Input,
		),
		container.NewGridWithColumns(
			3,
			httpSender.GetSelectMethod(),
			httpSender.Params,
			httpSender.SendBtn,
		),
		container.NewGridWithRows(
			1,
			httpSender.ScrollContainer,
		),
		container.NewGridWithColumns(
			1,
			btnExit,
		),
	)
	window.SetContent(content)
	window.CenterOnScreen()
	window.Resize(fyne.NewSize(500, 400))
	window.ShowAndRun()
}
