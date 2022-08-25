package models

import "time"

type ProductList struct {
	Result           []*Product `json:"result"`
	Total            int        `json:"total"`
	ErrorText        string     `json:"error"`
	ErrorDescription string     `json:"error_description"`
	Time             struct {
		Start      float64   `json:"start"`
		Finish     float64   `json:"finish"`
		Duration   float64   `json:"duration"`
		Processing float64   `json:"processing"`
		DateStart  time.Time `json:"date_start"`
		DateFinish time.Time `json:"date_finish"`
		Operating  float64   `json:"operating"`
	} `json:"time"`
}