package models

import "encoding/xml"

type RK7QueryGetRefList struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD string `xml:"CMD,attr"`
	} `xml:"RK7CMD"`
}

type RK7QueryResultGetRefList struct {
	XMLName       xml.Name `xml:"RK7QueryResult"`
	ServerVersion string   `xml:"ServerVersion,attr"`
	XmlVersion    string   `xml:"XmlVersion,attr"`
	NetName       string   `xml:"NetName,attr"`
	Status        string   `xml:"Status,attr"`
	CMD           string   `xml:"CMD,attr"`
	ErrorText     string   `xml:"ErrorText,attr"`
	DateTime      string   `xml:"DateTime,attr"`
	WorkTime      string   `xml:"WorkTime,attr"`
	Processed     string   `xml:"Processed,attr"`
	RK7RefList    struct {
		Count        string `xml:"Count,attr"`
		RK7Reference []struct {
			RefName     string `xml:"RefName,attr"`
			Count       string `xml:"Count,attr"`
			DataVersion int    `xml:"DataVersion,attr"`
		} `xml:"RK7Reference"`
	} `xml:"RK7RefList"`
}
