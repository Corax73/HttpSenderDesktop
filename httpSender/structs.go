package httpSender

import (
	common "httpSenderDesktop/common/structs"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	goutilsCurl "github.com/Corax73/goUtils/curl"
)

type State struct {
	Url, Params, Headers, Method, BasicAuthUsername, BasicAuthPassword, ResponseData string
	Repeat, Delay, CookieDefaultExpiration                                           int
	NotShowResult                                                                    bool
	Cookies                                                                          []CookieInstance
	UrlencodeData                                                                    []goutilsCurl.UrlencodeData
	Responses                                                                        []*common.CustomResponse
}

type CookieInstance struct {
	CookieName, CookieValue, CookieExpiration *widget.Entry
}

type HttpSender struct {
	State
	stateHistory                                                                                                                   map[string]*State
	UrlEntry, DisplayEntry, ParamsEntry, RepeatEntry, DelayEntry, BasicAuthUsernameEntry, BasicAuthPasswordEntry, HeadersEntry     *widget.Entry
	ScrollContainer                                                                                                                *container.Scroll
	SendBtn, ClearResultBtn, CopyBtn, ClearParametersBtn, SaveResultBtn, SetBasicAuthBtn, SetCookieBtn, SaveStateBtn, LoadStateBtn *widget.Button
	DisplayRepeat                                                                                                                  *widget.Label
	SelectMethod                                                                                                                   *widget.Select
	NotShowResultCheckbox                                                                                                          *widget.Check
	BasicAuthForm                                                                                                                  *widget.Form
}

type HttpResponseData struct {
	Error        error
	DataBytes    []byte
	RepeatNumber int
}
