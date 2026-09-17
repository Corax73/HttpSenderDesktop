package httpSender

import (
	common "httpSenderDesktop/common/structs"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	goutilsCurl "github.com/Corax73/goUtils/curl"
)

type httpState struct {
	common.State
	Headers, BasicAuthUsername, BasicAuthPassword string
	CookieDefaultExpiration                       int
	Cookies                                       []CookieInstance
	UrlencodeData                                 []goutilsCurl.UrlencodeData
	Responses                                     []*common.CustomResponse
}

type CookieInstance struct {
	CookieName, CookieValue, CookieExpiration *widget.Entry
}

type HttpSender struct {
	common.Sender
	httpState
	stateHistory                                                 map[string]*httpState
	BasicAuthUsernameEntry, BasicAuthPasswordEntry, HeadersEntry *widget.Entry
	ScrollContainer                                              *container.Scroll
	CopyBtn, SetBasicAuthBtn, SetCookieBtn                       *widget.Button
	SelectMethod                                                 *widget.Select
	NotShowResultCheckbox                                        *widget.Check
	BasicAuthForm                                                *widget.Form
}

type HttpResponseData struct {
	Error        error
	DataBytes    []byte
	RepeatNumber int
}
