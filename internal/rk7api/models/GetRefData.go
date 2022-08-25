package models

import (
	"encoding/xml"
)

type RK7QueryGetRefData struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD            string `xml:"CMD,attr"`
		RefName        string `xml:"RefName,attr"`
		OnlyActive     string `xml:"OnlyActive,attr,omitempty"`
		IgnoreEnums    string `xml:"IgnoreEnums,attr,omitempty"`
		WithChildItems string `xml:"WithChildItems,attr,omitempty"`
		WithMacroProp  string `xml:"WithMacroProp,attr,omitempty"`
		PropMask       string `xml:"PropMask,attr,omitempty"`
	} `xml:"RK7CMD"`
}

type RK7QueryResultGetRefDataMenuitems struct {
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
	RK7Reference  struct {
		DataVersion  int    `xml:"DataVersion,attr"`
		ClassName    string `xml:"ClassName,attr"`
		RIChildItems string `xml:"RIChildItems"`
		Items        struct {
			Item []MenuitemItem
		} `xml:"Items"`
	} `xml:"RK7Reference"`
}

type RK7QueryResultGetRefDataCateglist struct {
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
	RK7Reference  struct {
		DataVersion  int    `xml:"DataVersion,attr"`
		ClassName    string `xml:"ClassName,attr"`
		RIChildItems string `xml:"RIChildItems"`
		Items        struct {
			Item []Categlist
		} `xml:"Items"`
	} `xml:"RK7Reference"`
}

//options
type GetRefDataOptions func(*RK7QueryGetRefData)

func OnlyActive(value string) GetRefDataOptions {
	return func(q *RK7QueryGetRefData) {
		q.RK7CMD.OnlyActive = value
	}
}

func IgnoreEnums(value string) GetRefDataOptions {
	return func(q *RK7QueryGetRefData) {
		q.RK7CMD.IgnoreEnums = value
	}
}

func WithChildItems(value string) GetRefDataOptions {
	return func(q *RK7QueryGetRefData) {
		q.RK7CMD.WithChildItems = value
	}
}

func WithMacroProp(value string) GetRefDataOptions {
	return func(q *RK7QueryGetRefData) {
		q.RK7CMD.WithMacroProp = value
	}
}

func PropMask(value string) GetRefDataOptions {
	return func(q *RK7QueryGetRefData) {
		q.RK7CMD.PropMask = value
	}
}
