package models

type MenuitemItem struct {
	Code                int    `xml:"Code,attr"`
	Ident               int    `xml:"Ident,attr"`
	ItemIdent           int    `xml:"ItemIdent,attr"`
	GUIDString          string `xml:"GUIDString,attr"`
	Name                string `xml:"Name,attr"`
	MainParentIdent     int    `xml:"MainParentIdent,attr"`
	ExtCode             int    `xml:"ExtCode,attr"`
	CategPath           string `xml:"CategPath,attr"`
	PRICETYPES          int64  `xml:"PRICETYPES,attr"` // Тип цены используется для синхронизации
	Status              int    `xml:"Status,attr"`
	ID_BX24             int    `xml:"genIDBX24,attr"`
	SectionID_BX24      int    `xml:"genSectionIDBX24,attr"`
	WOO_ID              int    `xml:"genWOO_ID,attr"`
	WOO_PARENT_ID       int    `xml:"genWOO_PARENT_ID,attr"`
	WOO_LONGNAME        string `xml:"genWOO_LONGNAME,attr"`
	WOO_IMAGE_NAME      string `xml:"genWOO_IMAGE,attr"`
	WOO_IMAGE_NAME_2    string `xml:"genWOO_IMAGE_2,attr"`
	CLASSIFICATORGROUPS int    `xml:"CLASSIFICATORGROUPS,attr"` // Классификация для включения синхронизации с сайтом
}
