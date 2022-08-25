package models

type ProductSection struct {
	ID        string `json:"ID"` //--- (поиск делается в момент перебора всех элементов Categlist) RK7.ID_BX24
	CATALOGID string `json:"CATALOG_ID"`
	SECTIONID string `json:"SECTION_ID"` //+++ RK7.SectionID_BX24
	NAME      string `json:"NAME"`       //+++ RK7.Name
	CODE      string `json:"CODE"`
	XMLID     string `json:"XML_ID"` //+++ RK7.ItemIdent
}
