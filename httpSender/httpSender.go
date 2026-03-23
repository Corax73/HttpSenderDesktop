package httpSender

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.design/x/clipboard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const repetitionNumberStub = 0

type State struct {
	Method, BasicAuthUsername, BasicAuthPassword string
	Repeat, Delay, CookieDefaultExpiration       int
	NotShowResult                                bool
	Cookies                                      []*CookieInstance
}

func (state *State) ResetState() {
	state.Method, state.BasicAuthUsername, state.BasicAuthPassword = "", "", ""
	state.Repeat, state.Delay, state.CookieDefaultExpiration = 1, 200, 1
	state.NotShowResult = false
	state.Cookies = make([]*CookieInstance, 0)

}

type CookieInstance struct {
	CookieName, CookieValue, CookieExpiration *widget.Entry
}

type HttpSender struct {
	State
	Input, Display, Params, RepeatEntry, DelayEntry, BasicAuthUsernameEntry, BasicAuthPasswordEntry    *widget.Entry
	ScrollContainer                                                                                    *container.Scroll
	SendBtn, ClearResultBtn, CopyBtn, ClearParametersBtn, SaveResultBtn, SetBasicAuthBtn, SetCookieBtn *widget.Button
	DisplayRepeat                                                                                      *widget.Label
	SelectMethod                                                                                       *widget.Select
	NotShowResultCheckbox                                                                              *widget.Check
	BasicAuthForm                                                                                      *widget.Form
}

func (httpSender *HttpSender) SendBtnHandler() *widget.Button {
	return widget.NewButton("Send", func() {
		if httpSender.Input.Text != "" && httpSender.Method != "" {
			httpSender.getRepeat()
			httpSender.Display.SetText("")
			for i := 0; i < httpSender.Repeat; i++ {
				httpSender.showRepeat(i+1, false)
				resp, err := httpSender.SendByMethod()
				if err == nil {
					body, err := io.ReadAll(resp.Body)
					if err == nil {
						defer resp.Body.Close()
						var prettyJSON bytes.Buffer
						if !httpSender.NotShowResult {
							if err := json.Indent(&prettyJSON, []byte(body), "", "    "); err == nil {
								httpSender.showResp(prettyJSON.String(), i+1)
							} else {
								httpSender.showResp(err.Error(), i+1)
							}
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
			httpSender.showRepeat(1, true)
		} else {
			httpSender.showResp("Enter the request string", repetitionNumberStub)
		}
	})
}

func (httpSender *HttpSender) SendByMethod() (*http.Response, error) {
	var req *http.Request
	var resp *http.Response
	var err error
	client := &http.Client{
		Transport: &http.Transport{},
	}
	switch httpSender.Method {
	case "GET":
		req, err = http.NewRequest(http.MethodGet, httpSender.Input.Text, nil)
		if err == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
			if httpSender.BasicAuthUsername != "" && httpSender.BasicAuthPassword != "" {
				req.SetBasicAuth(httpSender.BasicAuthUsername, httpSender.BasicAuthPassword)
			}
			httpSender.setCookies(req)
			resp, err = client.Do(req)
		}
	case "POST":
		req, err = http.NewRequest(http.MethodPost, httpSender.Input.Text, httpSender.getParams())
		if err == nil {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
			if httpSender.BasicAuthUsername != "" && httpSender.BasicAuthPassword != "" {
				req.SetBasicAuth(httpSender.BasicAuthUsername, httpSender.BasicAuthPassword)
			}
			httpSender.setCookies(req)
			resp, err = client.Do(req)
		}
	case "DELETE":
		var req *http.Request
		req, err = http.NewRequest(http.MethodDelete, httpSender.Input.Text, nil)
		if err == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
			if httpSender.BasicAuthUsername != "" && httpSender.BasicAuthPassword != "" {
				req.SetBasicAuth(httpSender.BasicAuthUsername, httpSender.BasicAuthPassword)
			}
			httpSender.setCookies(req)
			resp, err = client.Do(req)
		}
	case "PUT":
		responseBody := httpSender.getParams()
		req, err = http.NewRequest(http.MethodPut, httpSender.Input.Text, responseBody)
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
			if httpSender.BasicAuthUsername != "" && httpSender.BasicAuthPassword != "" {
				req.SetBasicAuth(httpSender.BasicAuthUsername, httpSender.BasicAuthPassword)
			}
			httpSender.setCookies(req)
			resp, err = client.Do(req)
		}
	default:
		return resp, err
	}
	return resp, err
}

func (httpSender *HttpSender) showResp(data string, repeatNumber int) {
	var strBuilder strings.Builder
	if httpSender.Repeat > 1 {
		if httpSender.Display.Text != "" {
			strBuilder.WriteString("[")
			data := strings.Trim(httpSender.Display.Text, "[")
			data = strings.Trim(data, "]")
			strBuilder.WriteString(data)
			strBuilder.WriteString(",")
		}
		strBuilder.WriteString("{")
		strBuilder.WriteString("\n")
		strBuilder.WriteString("\"repeat_number\": ")
		strBuilder.WriteString(strconv.Itoa(repeatNumber))
		strBuilder.WriteString(",")
		strBuilder.WriteString("\n")
		strBuilder.WriteString("\"data\": \n")
		strBuilder.WriteString(data)
		strBuilder.WriteString("}")
		strBuilder.WriteString("\n")
		strBuilder.WriteString("]")
		httpSender.Display.SetText(strBuilder.String())
		strBuilder.Reset()
	} else {
		strBuilder.WriteString("[")
		strBuilder.WriteString("{")
		httpSender.Display.SetText(data)
		strBuilder.WriteString("}")
		strBuilder.WriteString("]")
	}
}

func (httpSender *HttpSender) showRepeat(repeatNumber int, isEnd bool) {
	var strBuilder strings.Builder
	if !isEnd {
		strBuilder.WriteString("Repeat №")
		strBuilder.WriteString(strconv.Itoa(repeatNumber))
	} else {
		strBuilder.WriteString(httpSender.DisplayRepeat.Text)
		strBuilder.WriteString(" All repetitions completed!")
	}
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
	resp := widget.NewSelect([]string{"GET", "POST", "DELETE", "PUT"}, func(value string) {
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

func (httpSender *HttpSender) ClearResultBtnHandler() *widget.Button {
	return widget.NewButton("Clear result", func() {
		httpSender.Display.SetText("")
	})
}

func (httpSender *HttpSender) CopyBtnHandler() *widget.Button {
	return widget.NewButton("Copy to clipboard", func() {
		err := clipboard.Init()
		if err != nil {
			httpSender.showResp(err.Error(), repetitionNumberStub)
		}
		clipboard.Write(clipboard.FmtText, []byte(httpSender.Display.Text))
	})
}

func (httpSender *HttpSender) ClearParametersBtnHandler() *widget.Button {
	return widget.NewButton("Clear all parameters", func() {
		httpSender.Input.SetText("")
		httpSender.Params.SetText("")
		httpSender.RepeatEntry.SetText("")
		httpSender.DelayEntry.SetText("")
		httpSender.SelectMethod.Selected = "Select method"
		httpSender.SelectMethod.Refresh()
		httpSender.BasicAuthUsernameEntry.SetText("")
		httpSender.BasicAuthPasswordEntry.SetText("")
		httpSender.ResetState()
	})
}

func (httpSender *HttpSender) SaveResultBtnHandler(appWindow fyne.Window) *widget.Button {
	return widget.NewButton("Save result to file", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				_, err := writer.Write([]byte(httpSender.Display.Text))
				if err != nil {
					dialog.ShowError(err, appWindow)
				}
			}
		}, appWindow)
	})
}

func (httpSender *HttpSender) NotShowResultCheckboxHandler() *widget.Check {
	return widget.NewCheck("Not show result(reduces the load)", func(value bool) {
		httpSender.NotShowResult = value
	})
}

func (httpSender *HttpSender) SetBasicAuthBtnHandler(appWindow fyne.Window) *widget.Button {
	basicAuthFormSlice := []*widget.FormItem{
		widget.NewFormItem("Username", httpSender.BasicAuthUsernameEntry),
		widget.NewFormItem("Password", httpSender.BasicAuthPasswordEntry),
	}
	onSubmitFunc := func(result bool) {
		if result && httpSender.BasicAuthUsernameEntry.Text != "" && httpSender.BasicAuthPasswordEntry.Text != "" {
			httpSender.BasicAuthUsername = httpSender.BasicAuthUsernameEntry.Text
			httpSender.BasicAuthPassword = httpSender.BasicAuthPasswordEntry.Text
		} else {
			httpSender.BasicAuthUsername,
				httpSender.BasicAuthPassword,
				httpSender.BasicAuthUsernameEntry.Text,
				httpSender.BasicAuthPasswordEntry.Text =
				"", "", "", ""
		}
	}
	return widget.NewButton("Set basic auth", func() {
		dialog.ShowForm(
			"Set username and password for basic auth",
			"Apply",
			"Cancel",
			basicAuthFormSlice,
			onSubmitFunc,
			appWindow,
		)
	})
}

func (httpSender *HttpSender) setCookies(req *http.Request) {
	for i, cookie := range httpSender.Cookies {
		name := cookie.CookieName.Text
		value := cookie.CookieValue.Text
		if name != "" && value != "" {
			expirationStr := cookie.CookieExpiration.Text
			expirationInt, err := strconv.Atoi(expirationStr)
			if err != nil || expirationInt <= 0 {
				expirationInt = httpSender.CookieDefaultExpiration
			}
			expiration := time.Now().Add(time.Duration(expirationInt) * time.Hour)
			cookie := http.Cookie{
				Name:     name,
				Value:    value,
				Expires:  expiration,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			req.AddCookie(&cookie)
		} else {
			httpSender.Cookies = slices.Delete(httpSender.Cookies, i, i+1)
		}
	}
}

func (httpSender *HttpSender) SetDynamicCookieFormBtnHandler(appWindow fyne.Window) *widget.Button {
	return widget.NewButton("Set cookies", func() {
		httpSender.showDynamicCookieFormDialog(appWindow)
	})
}

func (httpSender *HttpSender) showDynamicCookieFormDialog(appWindow fyne.Window) *widget.Button {
	cookieForm := widget.NewForm()
	if len(httpSender.Cookies) == 0 {
		newCookie := CookieInstance{widget.NewEntry(), widget.NewEntry(), widget.NewEntry()}
		httpSender.Cookies = append(httpSender.Cookies, &newCookie)
	}
	for _, cookie := range httpSender.Cookies {
		cookieForm.Append("Cookie name", cookie.CookieName)
		cookieForm.Append("Cookie value", cookie.CookieValue)
		cookieForm.Append("Cookie expiration", cookie.CookieExpiration)
		cookieForm.Append("Delete", httpSender.deleteCookieBtnHandler(cookie, cookieForm))
	}
	addButton := httpSender.newCookieBtnHandler(cookieForm)

	dialogContent := container.NewVBox(
		cookieForm,
		addButton,
	)

	dlg := dialog.NewCustomConfirm(
		"Set name, value and expiration time for cookies",
		"Submit",
		"Clear all cookies",
		dialogContent,
		func(ok bool) {
			if ok {
				for i, cookie := range httpSender.Cookies {
					name := cookie.CookieName.Text
					value := cookie.CookieValue.Text
					if name == "" || value == "" {
						httpSender.Cookies = slices.Delete(httpSender.Cookies, i, i+1)
					}
				}
			} else {
				httpSender.Cookies = make([]*CookieInstance, 0)
			}
		},
		appWindow,
	)

	dlg.Resize(fyne.NewSize(300, 250))
	dlg.Show()
	return addButton
}

func (httpSender *HttpSender) newCookieBtnHandler(cookieForm *widget.Form) *widget.Button {
	return widget.NewButton("Add new cookie", func() {
		newCookie := CookieInstance{widget.NewEntry(), widget.NewEntry(), widget.NewEntry()}
		httpSender.Cookies = append(httpSender.Cookies, &newCookie)
		cookieForm.Append("Cookie name", newCookie.CookieName)
		cookieForm.Append("Cookie value", newCookie.CookieValue)
		cookieForm.Append("Cookie expiration", newCookie.CookieExpiration)
		cookieForm.Append("Delete", httpSender.deleteCookieBtnHandler(&newCookie, cookieForm))
		cookieForm.Refresh()
	})
}

func (httpSender *HttpSender) deleteCookieBtnHandler(newCookie *CookieInstance, cookieForm *widget.Form) *widget.Button {
	return widget.NewButton(
		"Delete this cookie",
		func() {
			for i, cookie := range httpSender.Cookies {
				if cookie == newCookie {
					httpSender.Cookies = slices.Delete(httpSender.Cookies, i, i+1)
				}
			}
			cookieForm.Items = make([]*widget.FormItem, 0)
			for _, cookie := range httpSender.Cookies {
				cookieForm.Append("Cookie name", cookie.CookieName)
				cookieForm.Append("Cookie value", cookie.CookieValue)
				cookieForm.Append("Cookie expiration", cookie.CookieExpiration)
				cookieForm.Append("Delete", httpSender.deleteCookieBtnHandler(cookie, cookieForm))
			}
			cookieForm.Refresh()
		},
	)
}
