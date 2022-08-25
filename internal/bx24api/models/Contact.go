package models

import "time"

type Contact struct {
	ID                string      `json:"ID"`
	POST              interface{} `json:"POST"`
	COMMENTS          interface{} `json:"COMMENTS"`
	HONORIFIC         interface{} `json:"HONORIFIC"`
	NAME              string      `json:"NAME"`
	SECONDNAME        string      `json:"SECOND_NAME"`
	LASTNAME          string      `json:"LAST_NAME"`
	PHOTO             interface{} `json:"PHOTO"`
	LEADID            interface{} `json:"LEAD_ID"`
	TYPEID            interface{} `json:"TYPE_ID"`
	SOURCEID          interface{} `json:"SOURCE_ID"`
	SOURCEDESCRIPTION interface{} `json:"SOURCE_DESCRIPTION"`
	COMPANYID         interface{} `json:"COMPANY_ID"`
	BIRTHDATE         string      `json:"BIRTHDATE"`
	EXPORT            string      `json:"EXPORT"`
	HASPHONE          string      `json:"HAS_PHONE"`
	HASEMAIL          string      `json:"HAS_EMAIL"`
	HASIMOL           string      `json:"HAS_IMOL"`
	DATECREATE        time.Time   `json:"DATE_CREATE"`
	DATEMODIFY        time.Time   `json:"DATE_MODIFY"`
	ASSIGNEDBYID      string      `json:"ASSIGNED_BY_ID"`
	CREATEDBYID       string      `json:"CREATED_BY_ID"`
	MODIFYBYID        string      `json:"MODIFY_BY_ID"`
	OPENED            string      `json:"OPENED"`
	ORIGINATORID      interface{} `json:"ORIGINATOR_ID"`
	ORIGINID          interface{} `json:"ORIGIN_ID"`
	ORIGINVERSION     interface{} `json:"ORIGIN_VERSION"`
	FACEID            interface{} `json:"FACE_ID"`
	ADDRESS           interface{} `json:"ADDRESS"`
	ADDRESS2          interface{} `json:"ADDRESS_2"`
	ADDRESSCITY       interface{} `json:"ADDRESS_CITY"`
	ADDRESSPOSTALCODE interface{} `json:"ADDRESS_POSTAL_CODE"`
	ADDRESSREGION     interface{} `json:"ADDRESS_REGION"`
	ADDRESSPROVINCE   interface{} `json:"ADDRESS_PROVINCE"`
	ADDRESSCOUNTRY    interface{} `json:"ADDRESS_COUNTRY"`
	ADDRESSLOCADDRID  interface{} `json:"ADDRESS_LOC_ADDR_ID"`
	UTMSOURCE         interface{} `json:"UTM_SOURCE"`
	UTMMEDIUM         interface{} `json:"UTM_MEDIUM"`
	UTMCAMPAIGN       interface{} `json:"UTM_CAMPAIGN"`
	UTMCONTENT        interface{} `json:"UTM_CONTENT"`
	UTMTERM           interface{} `json:"UTM_TERM"`
	EMAIL             []struct {
		ID        string `json:"ID"`
		VALUETYPE string `json:"VALUE_TYPE"`
		VALUE     string `json:"VALUE"`
		TYPEID    string `json:"TYPE_ID"`
	} `json:"EMAIL"`
	PHONE []struct {
		ID        string `json:"ID"`
		VALUETYPE string `json:"VALUE_TYPE"`
		VALUE     string `json:"VALUE"`
		TYPEID    string `json:"TYPE_ID"`
	} `json:"PHONE"`
}
