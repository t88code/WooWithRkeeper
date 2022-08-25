package models

type MenuitemItem struct {
	Code            int    `xml:"Code,attr"`
	Ident           int    `xml:"Ident,attr"`
	ItemIdent       int    `xml:"ItemIdent,attr"`
	GUIDString      string `xml:"GUIDString,attr"`
	Name            string `xml:"Name,attr"`
	MainParentIdent int    `xml:"MainParentIdent,attr"`
	ExtCode         int    `xml:"ExtCode,attr"`
	CategPath       string `xml:"CategPath,attr"`
	PRICETYPES3     int64  `xml:"PRICETYPES-3,attr"`
	Status          int    `xml:"Status,attr"`
	ID_BX24         int    `xml:"genIDBX24,attr"`
	SectionID_BX24  int    `xml:"genSectionIDBX24,attr"`
}
