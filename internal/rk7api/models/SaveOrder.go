package models

import "encoding/xml"

type RK7QuerySaveOrder struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD              string `xml:"CMD,attr"`
		Deferred         string `xml:"deferred,attr,omitempty"`
		DontcheckLicense string `xml:"dontcheckLicense,attr,omitempty"`
		Order            struct {
			Visit int    `xml:"visit,attr"`
			Guid  string `xml:"guid,attr"`
		} `xml:"Order"`
		Session struct {
			Station Station `xml:"Station"`
			Dish    *[]Dish `xml:"Dish"`
			Prepay  *Prepay `xml:"Prepay"`
		} `xml:"Session"`

		LicenseInfo LicenseInfo `xml:"LicenseInfo"`
	} `xml:"RK7CMD"`
}

type RK7QueryResultSaveOrder struct {
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
	RK7ErrorN       string   `xml:"RK7ErrorN,attr,omitempty"`
	Order           *Order   `xml:"Order"`
	Session         struct {
		LineGuid  string `xml:"line_guid,attr"`
		SessionID string `xml:"sessionID,attr"`
	} `xml:"Session"`
}
