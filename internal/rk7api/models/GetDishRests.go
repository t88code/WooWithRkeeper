package models

import "encoding/xml"

type RK7QueryGetDishRests struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD string `xml:"CMD,attr"`
	} `xml:"RK7CMD"`
}

type RK7QueryResultGetDishRests struct {
	XMLName         xml.Name    `xml:"RK7QueryResult"`
	ServerVersion   string      `xml:"ServerVersion,attr,omitempty"`
	XmlVersion      string      `xml:"XmlVersion,attr,omitempty"`
	NetName         string      `xml:"NetName,attr,omitempty"`
	Status          string      `xml:"Status,attr,omitempty"`
	CMD             string      `xml:"CMD,attr,omitempty"`
	ErrorText       string      `xml:"ErrorText,attr,omitempty"`
	DateTime        string      `xml:"DateTime,attr,omitempty"`
	WorkTime        string      `xml:"WorkTime,attr,omitempty"`
	Processed       string      `xml:"Processed,attr,omitempty"`
	ArrivalDateTime string      `xml:"ArrivalDateTime,attr,omitempty"`
	DishRest        []*DishRest `xml:"DishRest,omitempty"`
}

type DishRest struct {
	ID         int    `xml:"id,attr,omitempty"`
	Code       int    `xml:"code,attr,omitempty"`
	Name       string `xml:"name,attr,omitempty"`
	Quantity   int    `xml:"quantity,attr,omitempty"`
	Prohibited int    `xml:"prohibited,attr,omitempty"`
}
