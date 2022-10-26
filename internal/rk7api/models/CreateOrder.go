package models

import "encoding/xml"

type RK7QueryCreateOrder struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD   string                      `xml:"CMD,attr"`
		Order *OrderInRK7QueryCreateOrder `xml:"Order"`
	} `xml:"RK7CMD"`
}

type OrderInRK7QueryCreateOrder struct {
	ExtID                string     `xml:"extID,attr"`
	ExtSource            string     `xml:"extSource,attr"`
	OpenTime             string     `xml:"openTime,attr,omitempty"`
	Duration             string     `xml:"duration,attr,omitempty"`
	PersistentComment    string     `xml:"persistentComment,omitempty,attr"`
	NonPersistentComment string     `xml:"nonPersistentComment,omitempty,attr"`
	Holder               string     `xml:"holder,attr,omitempty"`
	OrderType            *OrderType `xml:"OrderType"`
	Table                *Table     `xml:"Table"`
	Guests               struct {
		Item *[]Guest `xml:"Item"`
	} `xml:"Guests"`
	ExternalProps *ExternalProps `xml:"ExternalProps,omitempty"`
}

type RK7QueryResultCreateOrder struct {
	XMLName         xml.Name `xml:"RK7QueryResult"`
	ServerVersion   string   `xml:"ServerVersion,attr"`
	XmlVersion      string   `xml:"XmlVersion,attr"`
	NetName         string   `xml:"NetName,attr"`
	Status          string   `xml:"Status,attr"`
	CMD             string   `xml:"CMD,attr"`
	VisitID         int      `xml:"VisitID,attr"`
	OrderID         int      `xml:"OrderID,attr"`
	Guid            string   `xml:"guid,attr"`
	ErrorText       string   `xml:"ErrorText,attr"`
	DateTime        string   `xml:"DateTime,attr"`
	WorkTime        string   `xml:"WorkTime,attr"`
	Processed       string   `xml:"Processed,attr"`
	ArrivalDateTime string   `xml:"ArrivalDateTime,attr"`
	Order           *Order   `xml:"Order"`
}
