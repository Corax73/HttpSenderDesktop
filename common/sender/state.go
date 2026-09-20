package common

type State struct {
	Url, Params, Method, ResponseData string
	Repeat, Delay                     int
	NotShowResult                     bool
	Responses                         []*CustomResponse
}
