package models

import (
	"encoding/xml"
)

type RK7QueryGetXMLLicenseInstanceSeqNumber struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD         string                                     `xml:"CMD,attr"`
		LicenseInfo *LicenseInfoGetXMLLicenseInstanceSeqNumber `xml:"LicenseInfo"`
	} `xml:"RK7CMD"`
}

type LicenseInfoGetXMLLicenseInstanceSeqNumber struct {
	Anchor          string `xml:"anchor,attr"`
	LicenseToken    string `xml:"licenseToken,attr"`
	LicenseInstance struct {
		Guid string `xml:"guid,attr"`
	} `xml:"LicenseInstance"`
}

type RK7QueryResultGetXMLLicenseInstanceSeqNumber struct {
	XMLName       xml.Name    `xml:"RK7QueryResult"`
	ServerVersion string      `xml:"ServerVersion,attr"`
	XmlVersion    int         `xml:"XmlVersion,attr"`
	NetName       string      `xml:"NetName,attr"`
	Status        string      `xml:"Status,attr"`
	CMD           string      `xml:"CMD,attr"`
	ErrorText     string      `xml:"ErrorText,attr"`
	DateTime      string      `xml:"DateTime,attr"`
	WorkTime      string      `xml:"WorkTime,attr"`
	Processed     string      `xml:"Processed,attr"`
	LicenseInfo   LicenseInfo `xml:"LicenseInfo"`
}
