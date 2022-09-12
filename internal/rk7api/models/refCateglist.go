package models

type Categlist struct {
	Ident               int    `xml:"Ident,attr"`
	ItemIdent           int    `xml:"ItemIdent,attr"`
	GUIDString          string `xml:"GUIDString,attr"`
	Code                int    `xml:"Code,attr"`
	Name                string `xml:"Name,attr"`
	MainParentIdent     int    `xml:"MainParentIdent,attr"`
	Status              int    `xml:"Status,attr"`
	Parent              int    `xml:"Parent,attr"`
	ID_BX24             int    `xml:"genIDBX24,attr"`
	SectionID_BX24      int    `xml:"genSectionIDBX24,attr"`
	WooID               int    `xml:"genWooID,attr"`
	WooParentCategoryID int    `xml:"genWooParentCategoryID,attr"`
}
