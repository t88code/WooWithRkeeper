package models

type MenuitemItem struct {
	Code                int     `xml:"Code,attr"`
	Ident               int     `xml:"Ident,attr"`
	ItemIdent           int     `xml:"ItemIdent,attr"`
	GUIDString          string  `xml:"GUIDString,attr"`
	Name                string  `xml:"Name,attr"`
	MainParentIdent     int     `xml:"MainParentIdent,attr"`
	ExtCode             int     `xml:"ExtCode,attr"`
	CategPath           string  `xml:"CategPath,attr"`
	PRICETYPES          int64   `xml:"PRICETYPES,attr"` // Тип цены используется для синхронизации
	Status              int     `xml:"Status,attr"`
	ID_BX24             int     `xml:"genIDBX24,attr"`
	SectionID_BX24      int     `xml:"genSectionIDBX24,attr"`
	WOO_ID              int     `xml:"genWOO_ID,attr"`
	WOO_PARENT_ID       int     `xml:"genWOO_PARENT_ID,attr"`
	WOO_LONGNAME        string  `xml:"genWOO_LONGNAME,attr"`
	WOO_IMAGE_NAME_1    string  `xml:"genWOO_IMAGE,attr"`
	WOO_IMAGE_NAME_2    *string `xml:"genWOO_IMAGE_2,attr"`
	WOO_IMAGE_NAME_3    *string `xml:"genWOO_IMAGE_3,attr"`
	WOO_IMAGE_NAME_4    *string `xml:"genWOO_IMAGE_4,attr"`
	WOO_IMAGE_NAME_5    *string `xml:"genWOO_IMAGE_5,attr"`
	WOO_IMAGE_NAME_6    *string `xml:"genWOO_IMAGE_6,attr"`
	WOO_IMAGE_NAME_7    *string `xml:"genWOO_IMAGE_7,attr"`
	WOO_IMAGE_NAME_8    *string `xml:"genWOO_IMAGE_8,attr"`
	WOO_IMAGE_NAME_9    *string `xml:"genWOO_IMAGE_9,attr"`
	WOO_IMAGE_NAME_10   *string `xml:"genWOO_IMAGE_10,attr"`
	CLASSIFICATORGROUPS int     `xml:"CLASSIFICATORGROUPS,attr"` // Классификация для включения синхронизации с сайтом
}

//func (m *MenuitemItem) GetImageNamesString() (imageNamesString [10]*string) {
//	return [10]*string{
//		m.WOO_IMAGE_NAME_1,
//		m.WOO_IMAGE_NAME_2,
//		m.WOO_IMAGE_NAME_3,
//		m.WOO_IMAGE_NAME_4,
//		m.WOO_IMAGE_NAME_5,
//		m.WOO_IMAGE_NAME_6,
//		m.WOO_IMAGE_NAME_7,
//		m.WOO_IMAGE_NAME_8,
//		m.WOO_IMAGE_NAME_9,
//		m.WOO_IMAGE_NAME_10,
//	}
//}
//
//func (m *MenuitemItem) ImageNamesExistLen() (len int) {
//	imageNamesString := m.GetImageNamesString()
//	for _, imageNameString := range imageNamesString {
//		if imageNameString != nil {
//			len++
//		}
//	}
//	return
//}
