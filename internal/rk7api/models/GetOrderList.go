package models

import "encoding/xml"

type RK7QueryGetOrderList struct {
	XMLName xml.Name `xml:"RK7Query"`
	RK7CMD  struct {
		CMD string `xml:"CMD,attr"`
	} `xml:"RK7CMD"`
}

type RK7QueryResultGetOrderList struct {
	XMLName         xml.Name `xml:"RK7QueryResult"`
	ServerVersion   string   `xml:"ServerVersion,attr"`
	XmlVersion      int      `xml:"XmlVersion,attr"`
	NetName         string   `xml:"NetName,attr"`
	Status          string   `xml:"Status,attr"`
	CMD             string   `xml:"CMD,attr"`
	Lastversion     int      `xml:"lastversion,attr"`
	ErrorText       string   `xml:"ErrorText,attr"`
	DateTime        string   `xml:"DateTime,attr"`
	WorkTime        int      `xml:"WorkTime,attr"`
	Processed       string   `xml:"Processed,attr"`
	ArrivalDateTime string   `xml:"ArrivalDateTime,attr"`
	Visit           []struct {
		VisitID           int     `xml:"VisitID,attr"`
		Guid              string  `xml:"guid,attr"`
		Finished          int     `xml:"Finished,attr"`
		GuestsCount       string  `xml:"GuestsCount,attr"`
		PersistentComment string  `xml:"PersistentComment,attr"`
		Guests            *Guests `xml:"Guests"`
		Orders            struct {
			Order []struct {
				OrderID        string `xml:"OrderID,attr"`
				OrderName      string `xml:"OrderName,attr"`
				URL            string `xml:"url,attr"`
				Version        int    `xml:"Version,attr"`
				Crc32          string `xml:"crc32,attr"`
				Guid           string `xml:"guid,attr"`
				TableID        int    `xml:"TableID,attr"`
				TableCode      int    `xml:"TableCode,attr"`
				OrderCategID   int    `xml:"OrderCategID,attr"`
				OrderCategCode int    `xml:"OrderCategCode,attr"`
				WaiterID       int    `xml:"WaiterID,attr"`
				WaiterCode     int    `xml:"WaiterCode,attr"`
				OrderSum       int    `xml:"OrderSum,attr"`
				ToPaySum       int    `xml:"ToPaySum,attr"`
				PriceListSum   int    `xml:"PriceListSum,attr"`
				TotalPieces    int    `xml:"TotalPieces,attr"`
				Finished       int    `xml:"Finished,attr"`
				Bill           int    `xml:"Bill,attr"`
				CreateTime     string `xml:"CreateTime,attr"`
				FinishTime     string `xml:"FinishTime,attr"`
				Dessert        string `xml:"Dessert,attr"`
				Reserve        int    `xml:"reserve,attr"`
				Duration       string `xml:"duration,attr"`
				OrderTypeID    int    `xml:"OrderTypeID,attr"`
				OrderTypeCode  int    `xml:"OrderTypeCode,attr"`
				ExternalID     []struct {
					ExtSource string `xml:"ExtSource,attr"`
					ExtID     string `xml:"ExtID,attr"`
				} `xml:"ExternalID"`
			} `xml:"Order"`
		} `xml:"Orders"`
	} `xml:"Visit"`
}
