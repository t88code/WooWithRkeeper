package models

import "encoding/xml"

type RK7QueryGetOrder struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD  string `xml:"CMD,attr"`
		Guid string `xml:"guid,attr"`
	} `xml:"RK7CMD"`
}

type RK7QueryResultGetOrder struct {
	XMLName         xml.Name `xml:"RK7QueryResult"`
	ServerVersion   string   `xml:"ServerVersion,attr"`
	XmlVersion      string   `xml:"XmlVersion,attr"`
	NetName         string   `xml:"NetName,attr"`
	Status          string   `xml:"Status,attr"`
	CMD             string   `xml:"CMD,attr"`
	ErrorText       string   `xml:"ErrorText,attr"`
	DateTime        string   `xml:"DateTime,attr"`
	WorkTime        string   `xml:"WorkTime,attr"`
	Processed       string   `xml:"Processed,attr"`
	ArrivalDateTime string   `xml:"ArrivalDateTime,attr"`
	Order           *Order   `xml:"Order"`
}
