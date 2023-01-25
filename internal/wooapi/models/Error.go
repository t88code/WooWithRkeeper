package models

import "fmt"

type ErrorWoo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Status int `json:"status"`
		Params struct {
			Display string `json:"display"`
		} `json:"params"`
		Details struct {
			Display struct {
				Code    string      `json:"code"`
				Message string      `json:"message"`
				Data    interface{} `json:"data"`
			} `json:"display"`
		} `json:"details"`
		ResourceId int `json:"resource_id"`
	} `json:"data"`
}

func (e *ErrorWoo) Error() string {
	return fmt.Sprintf("code:%s; message:%s; status:%s; display:%s; details:%s;",
		fmt.Sprint(e.Code),
		fmt.Sprint(e.Message),
		fmt.Sprint(e.Data.Status),
		fmt.Sprint(e.Data.Params.Display),
		fmt.Sprint(e.Data.Details.Display.Message),
	)
}
