package models

import "time"

type Product struct {
	ID              string      `json:"ID"`   // ReadOnly
	NAME            string      `json:"NAME"` //+++ Наименование
	CODE            string      `json:"CODE"`
	ACTIVE          string      `json:"ACTIVE"`          //+++ "type": "char"???
	PREVIEWPICTURE  interface{} `json:"PREVIEW_PICTURE"` //--- Картинка для анонса "type": "product_file"???
	DETAILPICTURE   interface{} `json:"DETAIL_PICTURE"`  //--- Детальная картинка "type": "product_file"???
	SORT            string      `json:"SORT"`
	XMLID           string      `json:"XML_ID"` //+++ Внешний ID
	TIMESTAMPX      time.Time   `json:"TIMESTAMP_X"`
	DATECREATE      time.Time   `json:"DATE_CREATE"`
	MODIFIEDBY      string      `json:"MODIFIED_BY"`
	CREATEDBY       string      `json:"CREATED_BY"`
	CATALOGID       string      `json:"CATALOG_ID"`  // ReadOnly, 25 - Товарный каталог CRM
	SECTIONID       string      `json:"SECTION_ID"`  //+++ Раздел - Categlist
	DESCRIPTION     string      `json:"DESCRIPTION"` //--- Детальное описание
	DESCRIPTIONTYPE string      `json:"DESCRIPTION_TYPE"`
	PRICE           string      `json:"PRICE"` //+++ Цена
	CURRENCYID      string      `json:"CURRENCY_ID"`
	VATID           string      `json:"VAT_ID"`       //--- Ставка НДС
	VATINCLUDED     string      `json:"VAT_INCLUDED"` //--- НДС включён в цену "char"???
	MEASURE         string      `json:"MEASURE"`      //--- Единица измерения
	PROPERTY101     interface{} `json:"PROPERTY_101"` //--- Картинки галереи "type": "product_property"???
}
