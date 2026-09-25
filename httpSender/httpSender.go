package httpSender

import (
	"bytes"
	"encoding/json"
	"fmt"
	common "httpSenderDesktop/common/sender"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	goutilsCurl "github.com/Corax73/goUtils/curl"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (httpSender *HttpSender) ResetState() {
	httpSender.Headers, httpSender.BasicAuthUsername, httpSender.BasicAuthPassword = "", "", ""
	httpSender.CookieDefaultExpiration = 1
	httpSender.SetUrl("")
	httpSender.SetParams("")
	httpSender.SetMethod("")
	httpSender.SetResponseData("")
	httpSender.SetRepeat(1)
	httpSender.SetDelay(200)
	httpSender.SetNotShowResult(false)
	httpSender.Cookies = make([]CookieInstance, 0)
	httpSender.UrlencodeData = make([]goutilsCurl.UrlencodeData, 0)
	httpSender.SetResponses(make([]*common.CustomResponse, 0))
}

func (httpSender *HttpSender) Load() {
	httpSender.stateHistory = make(map[string]*httpState)
	httpSender.UrlEntry.OnChanged = func(content string) {
		if strings.Contains(content, "curl") {
			curlData := goutilsCurl.ParseCurlString(content)
			httpSender.UrlEntry.SetText(curlData.Url)
			httpSender.SelectMethod.SetSelected(curlData.Method)
			httpSender.ParamsEntry.SetText(curlData.Data)
			if len(curlData.UrlencodeData) != 0 {
				curlData.Headers["Content-Type"] = "application/x-www-form-urlencoded"
				httpSender.UrlencodeData = curlData.UrlencodeData
			}
			if len(curlData.Headers) != 0 {
				headersBytes, err := json.MarshalIndent(curlData.Headers, "", " ")
				if err == nil {
					httpSender.HeadersEntry.SetText(string(headersBytes))
				}
			}
			httpSender.Cookies = make([]CookieInstance, 0)
			for k, v := range curlData.Cookies {
				newCookie := CookieInstance{widget.NewEntry(), widget.NewEntry(), widget.NewEntry()}
				newCookie.CookieName.SetText(k)
				newCookie.CookieValue.SetText(v)
				httpSender.Cookies = append(httpSender.Cookies, newCookie)
			}
		}
	}
}

func (httpSender *HttpSender) SendBtnHandler() *widget.Button {
	return widget.NewButton("Send", func() {
		if httpSender.UrlEntry.Text != "" && *httpSender.GetMethod() != "" {
			httpSender.ParseRepeat()
			httpSender.SetUrl(httpSender.UrlEntry.Text)
			httpSender.SetParams(httpSender.ParamsEntry.Text)
			httpSender.Headers = httpSender.HeadersEntry.Text
			httpSender.DisplayEntry.SetText("")
			repetitionChans := make([]chan *HttpResponseData, httpSender.GetRepeat())
			for i := 0; i < httpSender.GetRepeat(); i++ {
				repetitionChans[i] = make(chan *HttpResponseData, 1)
			}
			client := &http.Client{
				Timeout: 30 * time.Second,
				Transport: &http.Transport{
					MaxIdleConnsPerHost: httpSender.GetRepeat(),
				},
			}
			var wg sync.WaitGroup
			defer wg.Wait()
			start := time.Now()
			httpSender.switchingAvailability(false)
			httpSender.ParseDelay()
			for i := 0; i < httpSender.GetRepeat(); i++ {
				wg.Add(1)
				go func(counter int) {
					defer wg.Done()
					httpSender.SendByMethod(client, repetitionChans[counter], counter+1)
				}(i)
				if httpSender.GetRepeat() > 1 {
					time.Sleep(time.Duration(httpSender.GetDelay()) * time.Millisecond)
				}
			}
			httpSender.DisplayEntry.SetPlaceHolder("Reading responses to requests...")
			for i, ch := range repetitionChans {
				httpSender.showRepeat(i+1, false, nil)
				resp := <-ch
				if json.Valid(resp.DataBytes) {
					httpSender.SetResponses(
						append(
							httpSender.GetResponses(),
							&common.CustomResponse{Data: json.RawMessage(resp.DataBytes), RepeatNumber: resp.RepeatNumber},
						))
				} else {
					if resp.Error != nil {
						httpSender.SetResponses(
							append(
								httpSender.GetResponses(),
								&common.CustomResponse{
									Data: json.RawMessage(
										strings.ReplaceAll(
											strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(resp.Error.Error(), `"`, ""), "\r", ""), "\n", ""),
											":", " "),
									),
									RepeatNumber: resp.RepeatNumber,
								},
							))
					} else {
						errMsg := fmt.Sprintf(
							`{"error": "Invalid JSON response", "body": %q}`,
							strings.ReplaceAll(
								strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(resp.DataBytes), `"`, ""), "\r", ""), "\n", ""),
								":", " "),
						)
						httpSender.SetResponses(
							append(
								httpSender.GetResponses(),
								&common.CustomResponse{Data: json.RawMessage(errMsg), RepeatNumber: resp.RepeatNumber},
							))
					}
				}
				close(ch)
			}
			httpSender.DisplayEntry.SetPlaceHolder("")
			repetitionChans = nil
			bytesData, err := json.MarshalIndent(httpSender.GetResponses(), "", " ")
			if err != nil {
				for _, item := range httpSender.GetResponses() {
					if len(item.Data) == 0 {
						item.Data = json.RawMessage(`{"error": "Empty response"}`)
						continue
					}

					if !json.Valid(item.Data) {
						item.Data = json.RawMessage(fmt.Sprintf(`"error": "%s"`, string(item.Data)))
					}
				}
				bytesData, _ = json.MarshalIndent(httpSender.GetResponses(), "", " ")
			}
			httpSender.SetResponseData(string(bytesData))
			httpSender.SetResponses(nil)
			if !httpSender.GetNotShowResult() {
				httpSender.ShowResp(httpSender.GetResponseData())
			}
			timeSpent := time.Since(start)
			httpSender.showRepeat(1, true, &timeSpent)
			httpSender.switchingAvailability(true)
		} else {
			defaultResp := "Enter the request string"
			httpSender.ShowResp(&defaultResp)
		}
	})
}

func (httpSender *HttpSender) SendByMethod(client *http.Client, ch chan *HttpResponseData, repeatNumber int) {
	jsonData, err := httpSender.getParams()
	if err != nil {
		ch <- &HttpResponseData{Error: err, RepeatNumber: repeatNumber}
		return
	}
	req, err := http.NewRequest(*httpSender.GetMethod(), httpSender.UrlEntry.Text, jsonData)
	if err != nil {
		ch <- &HttpResponseData{Error: err, RepeatNumber: repeatNumber}
		return
	}

	err = httpSender.setHeadersCookiesAndAuth(req)
	if err != nil {
		ch <- &HttpResponseData{Error: err, RepeatNumber: repeatNumber}
		return
	}
	httpSender.applyUrlencodeData(req)

	resp, err := client.Do(req)
	if err != nil {
		ch <- &HttpResponseData{Error: err, RepeatNumber: repeatNumber}
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- &HttpResponseData{Error: err, RepeatNumber: repeatNumber}
		return
	}

	ch <- &HttpResponseData{DataBytes: body, RepeatNumber: repeatNumber}
}

func (httpSender *HttpSender) showRepeat(repeatNumber int, isEnd bool, timeSpent *time.Duration) {
	var strBuilder strings.Builder
	if !isEnd {
		strBuilder.WriteString("Repeat №")
		strBuilder.WriteString(strconv.Itoa(repeatNumber))
	} else {
		strBuilder.WriteString(httpSender.DisplayRepeat.Text)
		strBuilder.WriteString(" All repetitions completed! Time spent ")
		strBuilder.WriteString(timeSpent.String())
	}
	httpSender.DisplayRepeat.SetText(strBuilder.String())
	strBuilder.Reset()
}

func (httpSender *HttpSender) GetSelectMethod() *widget.Select {
	resp := widget.NewSelect([]string{"GET", "POST", "DELETE", "PUT"}, func(value string) {
		httpSender.SetMethod(value)
	})
	resp.PlaceHolder = "Select method"
	return resp
}

func (httpSender *HttpSender) getParams() (*bytes.Buffer, error) {
	data := make(map[string]any)
	str := httpSender.ParamsEntry.Text
	if str == "" {
		str = "{}"
	}
	err := json.Unmarshal([]byte(str), &data)
	if err != nil {
		errResp := err.Error()
		httpSender.ShowResp(&errResp)
		return nil, err
	}
	postBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	responseBody := bytes.NewBuffer(postBody)
	return responseBody, nil
}

func (httpSender *HttpSender) ClearParametersBtnHandler() *widget.Button {
	return widget.NewButton("Clear all parameters", func() {
		httpSender.UrlEntry.SetText("")
		httpSender.ParamsEntry.SetText("")
		httpSender.RepeatEntry.SetText("")
		httpSender.DelayEntry.SetText("")
		httpSender.SelectMethod.Selected = "Select method"
		httpSender.SelectMethod.Refresh()
		httpSender.BasicAuthUsernameEntry.SetText("")
		httpSender.BasicAuthPasswordEntry.SetText("")
		httpSender.HeadersEntry.SetText("")
		httpSender.ResetState()
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
	var validCookies []CookieInstance
	for _, cookie := range httpSender.Cookies {
		name := cookie.CookieName.Text
		value := cookie.CookieValue.Text
		if name != "" && value != "" {
			expirationStr := cookie.CookieExpiration.Text
			expirationInt, err := strconv.Atoi(expirationStr)
			if err != nil || expirationInt <= 0 {
				expirationInt = httpSender.CookieDefaultExpiration
			}
			expiration := time.Now().Add(time.Duration(expirationInt) * time.Hour)
			validCookies = append(validCookies, cookie)
			reqCookie := http.Cookie{
				Name:    name,
				Value:   value,
				Expires: expiration,
				Path:    "/",
			}
			req.AddCookie(&reqCookie)
		}
	}
	httpSender.Cookies = validCookies
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
		httpSender.Cookies = append(httpSender.Cookies, newCookie)
	}
	for _, cookie := range httpSender.Cookies {
		cookieForm.Append("Cookie name", cookie.CookieName)
		cookieForm.Append("Cookie value", cookie.CookieValue)
		cookieForm.Append("Cookie expiration", cookie.CookieExpiration)
		cookieForm.Append("Delete", httpSender.deleteCookieBtnHandler(&cookie, cookieForm))
	}
	addButton := httpSender.newCookieBtnHandler(cookieForm)

	dialogContent := container.NewScroll(
		container.NewVBox(
			cookieForm,
			addButton,
		),
	)

	dlg := dialog.NewCustomConfirm(
		"Set name, value and expiration time for cookies",
		"Submit",
		"Clear all cookies",
		dialogContent,
		func(ok bool) {
			if ok {
				var validCookies []CookieInstance
				for _, cookie := range httpSender.Cookies {
					name := cookie.CookieName.Text
					value := cookie.CookieValue.Text
					if name != "" || value != "" {
						validCookies = append(validCookies, cookie)
					}
				}
				httpSender.Cookies = validCookies
			} else {
				httpSender.Cookies = make([]CookieInstance, 0)
			}
		},
		appWindow,
	)

	dlg.Resize(fyne.NewSize(300, float32((len(httpSender.Cookies)+1)*170)))
	dlg.Show()
	return addButton
}

func (httpSender *HttpSender) newCookieBtnHandler(cookieForm *widget.Form) *widget.Button {
	return widget.NewButton("Add new cookie", func() {
		newCookie := CookieInstance{widget.NewEntry(), widget.NewEntry(), widget.NewEntry()}
		httpSender.Cookies = append(httpSender.Cookies, newCookie)
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
			var validCookies []CookieInstance
			for _, cookie := range httpSender.Cookies {
				if cookie != *newCookie {
					validCookies = append(validCookies, cookie)
				}
			}
			httpSender.Cookies = validCookies
			cookieForm.Items = make([]*widget.FormItem, 0)
			for _, cookie := range httpSender.Cookies {
				cookieForm.Append("Cookie name", cookie.CookieName)
				cookieForm.Append("Cookie value", cookie.CookieValue)
				cookieForm.Append("Cookie expiration", cookie.CookieExpiration)
				cookieForm.Append("Delete", httpSender.deleteCookieBtnHandler(&cookie, cookieForm))
			}
			cookieForm.Refresh()
		},
	)
}

func (httpSender *HttpSender) setHeadersCookiesAndAuth(req *http.Request) (err error) {
	if httpSender.HeadersEntry.Text == "" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	} else {
		headers := make(map[string]string)
		err = json.Unmarshal([]byte(httpSender.HeadersEntry.Text), &headers)
		if err != nil {
			return
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	if httpSender.BasicAuthUsername != "" && httpSender.BasicAuthPassword != "" {
		req.SetBasicAuth(httpSender.BasicAuthUsername, httpSender.BasicAuthPassword)
	}
	httpSender.setCookies(req)
	return
}

func (httpSender *HttpSender) SaveState(title string) {
	if title != "" {
		httpSender.ParseRepeat()
		httpSender.ParseDelay()
		httpSender.stateHistory[title] = &httpState{
			common.State{
				Url:           httpSender.UrlEntry.Text,
				Params:        httpSender.ParamsEntry.Text,
				Method:        *httpSender.GetMethod(),
				ResponseData:  "",
				Repeat:        httpSender.GetRepeat(),
				Delay:         httpSender.GetDelay(),
				NotShowResult: httpSender.GetNotShowResult(),
				Responses:     httpSender.GetResponses(),
			},
			httpSender.HeadersEntry.Text,
			httpSender.BasicAuthUsernameEntry.Text,
			httpSender.BasicAuthPasswordEntry.Text,
			httpSender.CookieDefaultExpiration,
			httpSender.Cookies,
			httpSender.UrlencodeData,
		}
	}
}

func (httpSender *HttpSender) GetStatesSelect() *widget.Select {
	var keys []string
	for k := range httpSender.stateHistory {
		keys = append(keys, k)
	}
	return widget.NewSelect(keys, func(value string) {})
}

func (httpSender *HttpSender) UseStateByTitle(title string) {
	state, ok := httpSender.stateHistory[title]
	if ok {
		httpSender.UrlEntry.SetText(state.Url)
		httpSender.ParamsEntry.SetText(state.Params)
		httpSender.HeadersEntry.SetText(state.Headers)
		httpSender.SelectMethod.SetSelected(state.Method)
		httpSender.BasicAuthUsernameEntry.SetText(state.BasicAuthUsername)
		httpSender.BasicAuthPasswordEntry.SetText(state.BasicAuthPassword)
		httpSender.RepeatEntry.SetText(strconv.Itoa(state.Repeat))
		httpSender.DelayEntry.SetText(strconv.Itoa(state.Delay))
		httpSender.NotShowResultCheckbox.SetChecked(state.NotShowResult)
		httpSender.Cookies = state.Cookies
		httpSender.UrlencodeData = state.UrlencodeData
		httpSender.SetResponses(state.Responses)
	}
}

func (httpSender *HttpSender) applyUrlencodeData(req *http.Request) {
	if len(httpSender.UrlencodeData) != 0 {
		q := req.URL.Query()
		for _, data := range httpSender.UrlencodeData {
			q.Add(data.Key, data.Value)
		}
		req.URL.RawQuery = q.Encode()
	}
}

func (httpSender *HttpSender) switchingAvailability(isOn bool) {
	if isOn {
		httpSender.UrlEntry.Enable()
		httpSender.DisplayEntry.Enable()
		httpSender.ParamsEntry.Enable()
		httpSender.RepeatEntry.Enable()
		httpSender.DelayEntry.Enable()
		httpSender.BasicAuthUsernameEntry.Enable()
		httpSender.BasicAuthPasswordEntry.Enable()
		httpSender.HeadersEntry.Enable()
		httpSender.SendBtn.Enable()
		httpSender.SendBtn.SetText("Send")
		httpSender.ClearResultBtn.Enable()
		httpSender.CopyBtn.Enable()
		httpSender.ClearParametersBtn.Enable()
		httpSender.SaveResultBtn.Enable()
		httpSender.SetBasicAuthBtn.Enable()
		httpSender.SetCookieBtn.Enable()
		httpSender.SaveStateBtn.Enable()
		httpSender.LoadStateBtn.Enable()
		httpSender.SelectMethod.Enable()
		httpSender.NotShowResultCheckbox.Enable()
	} else {
		httpSender.UrlEntry.Disable()
		httpSender.DisplayEntry.Disable()
		httpSender.ParamsEntry.Disable()
		httpSender.RepeatEntry.Disable()
		httpSender.DelayEntry.Disable()
		httpSender.BasicAuthUsernameEntry.Disable()
		httpSender.BasicAuthPasswordEntry.Disable()
		httpSender.HeadersEntry.Disable()
		httpSender.SendBtn.Disable()
		httpSender.ClearResultBtn.Disable()
		httpSender.CopyBtn.Disable()
		httpSender.ClearParametersBtn.Disable()
		httpSender.SaveResultBtn.Disable()
		httpSender.SetBasicAuthBtn.Disable()
		httpSender.SetCookieBtn.Disable()
		httpSender.SaveStateBtn.Disable()
		httpSender.LoadStateBtn.Disable()
		httpSender.SelectMethod.Disable()
		httpSender.NotShowResultCheckbox.Disable()
		httpSender.SendBtn.SetText("Sending...")
	}
}
