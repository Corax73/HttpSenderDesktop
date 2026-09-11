package common

import "encoding/json"

type CustomResponse struct {
	Data         json.RawMessage `json:"data"`
	RepeatNumber int             `json:"repeat_number"`
}
