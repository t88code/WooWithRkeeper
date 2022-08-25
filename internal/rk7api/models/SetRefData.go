package models

import "encoding/xml"

type RK7QuerySetRefDataMenuitems struct {
	XMLName    xml.Name `xml:"RK7Query"`
	RK7Command struct {
		CMD     string `xml:"CMD,attr"`
		RefName string `xml:"RefName,attr"`
		Items   struct {
			Item []*MenuitemItem `xml:"Item"`
		} `xml:"Items"`
	} `xml:"RK7Command"`
}

type RK7QuerySetRefDataCateglist struct {
	XMLName    xml.Name `xml:"RK7Query"`
	RK7Command struct {
		CMD     string `xml:"CMD,attr"`
		RefName string `xml:"RefName,attr"`
		Items   struct {
			Item []*Categlist `xml:"Item"`
		} `xml:"Items"`
	} `xml:"RK7Command"`
}

type RK7QueryResultSetRefData struct {
	XMLName         xml.Name `xml:"RK7QueryResult"`
	ServerVersion   string   `xml:"ServerVersion,attr"`
	XmlVersion      string   `xml:"XmlVersion,attr"`
	NetName         string   `xml:"NetName,attr"`
	Status          string   `xml:"Status,attr"`
	ErrorText       string   `xml:"ErrorText,attr"`
	Processed       string   `xml:"Processed,attr"`
	ArrivalDateTime string   `xml:"ArrivalDateTime,attr"`
	CommandResult   struct {
		CMD       string `xml:"CMD,attr"`
		Status    string `xml:"Status,attr"`
		ErrorText string `xml:"ErrorText,attr"`
		DateTime  string `xml:"DateTime,attr"`
		WorkTime  string `xml:"WorkTime,attr"`
	} `xml:"CommandResult"`
}
