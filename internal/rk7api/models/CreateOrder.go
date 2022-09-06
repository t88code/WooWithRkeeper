package models

import "encoding/xml"

type RK7QueryCreateOrder struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD   string                      `xml:"CMD,attr"`
		Order *OrderInRK7QueryCreateOrder `xml:"Order"`
	} `xml:"RK7CMD"`
}

// VisitElement is Ссылка на визит. Можно выбрать либо по guid либо по visit
type VisitElement struct {
	Guid  interface{} `xml:"guid,attr,omitempty"`
	Visit int         `xml:"visit,attr,omitempty"`
}

// ExtraTables ...
type ExtraTables struct {
	Item []interface{} `xml:"item"`
}

// Order ... TODO непроверен до конца
type OrderInRK7QueryCreateOrder struct {
	ExtSource            string         `xml:"extSource,attr"`
	ExtID                string         `xml:"extID,attr"`
	OpenTime             string         `xml:"openTime,attr,omitempty"`
	Duration             string         `xml:"duration,attr,omitempty"`
	Holder               string         `xml:"holder,attr,omitempty"`
	PromoCode            string         `xml:"promoCode,attr,omitempty"`
	PersistentComment    string         `xml:"persistentComment,attr,omitempty"`
	NonPersistentComment string         `xml:"nonPersistentComment,attr,omitempty"`
	Guid                 string         `xml:"guid,attr,omitempty"`
	Visit                *VisitElement  `xml:"Visit"`
	Table                *Table         `xml:"Table"`
	Station              *Station       `xml:"Station"`
	Creator              *Creator       `xml:"Creator"`
	Waiter               *Waiter        `xml:"Waiter"`
	OrderCategory        *OrderCategory `xml:"OrderCategory"`
	OrderType            *OrderType     `xml:"OrderType"`
	Defaulter            interface{}    `xml:"Defaulter"`
	GuestType            interface{}    `xml:"GuestType"`
	Guests               *struct {
		Item *[]Guest `xml:"Item"`
	} `xml:"Guests"`
	ExtraTables   *ExtraTables   `xml:"ExtraTables"`
	ExternalProps *ExternalProps `xml:"ExternalProps"`
	DeliveryBlock interface{}    `xml:"DeliveryBlock"`
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
