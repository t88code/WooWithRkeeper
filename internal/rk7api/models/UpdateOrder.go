package models

import "encoding/xml"

type RK7QueryUpdateOrder struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD   string `xml:"CMD,attr"`
		Order struct {
			Guid string `xml:"guid,attr"`
		} `xml:"Order"`
		Table         *Table         `xml:"Table,omitempty"`
		Waiter        *Waiter        `xml:"Waiter,omitempty"`
		OrderType     *OrderType     `xml:"OrderType,omitempty"`
		Guests        *Guests        `xml:"Guests,omitempty"`
		ExternalProps *ExternalProps `xml:"ExternalProps,omitempty"`
	} `xml:"RK7CMD"`
}

type RK7QueryResultUpdateOrder struct {
	XMLName         xml.Name `xml:"RK7QueryResult"`
	ServerVersion   string   `xml:"ServerVersion,attr"`
	XmlVersion      int      `xml:"XmlVersion,attr"`
	NetName         string   `xml:"NetName,attr"`
	Status          string   `xml:"Status,attr"`
	CMD             string   `xml:"CMD,attr"`
	ErrorText       string   `xml:"ErrorText,attr"`
	DateTime        string   `xml:"DateTime,attr"`
	WorkTime        string   `xml:"WorkTime,attr"`
	Processed       string   `xml:"Processed,attr"`
	ArrivalDateTime string   `xml:"ArrivalDateTime,attr"`
}

type FieldUpdateOrder func(*RK7QueryUpdateOrder)

func ExternalProp(name, value string) FieldUpdateOrder {
	return func(RK7QueryUpdateOrder *RK7QueryUpdateOrder) {
		if RK7QueryUpdateOrder.RK7CMD.ExternalProps == nil {
			RK7QueryUpdateOrder.RK7CMD.ExternalProps = new(ExternalProps)
		}
		RK7QueryUpdateOrder.RK7CMD.ExternalProps.Prop = append(RK7QueryUpdateOrder.RK7CMD.ExternalProps.Prop, &Prop{
			Name:  name,
			Value: value,
		})
	}
}
