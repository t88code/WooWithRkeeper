package models

import "time"

type Deal struct {
	ID                  string      `json:"ID"`
	TITLE               string      `json:"TITLE"` //Наименование сделки
	TYPEID              string      `json:"TYPE_ID"`
	STAGEID             string      `json:"STAGE_ID"`
	PROBABILITY         interface{} `json:"PROBABILITY"`
	CURRENCYID          string      `json:"CURRENCY_ID"`
	OPPORTUNITY         string      `json:"OPPORTUNITY"`
	ISMANUALOPPORTUNITY string      `json:"IS_MANUAL_OPPORTUNITY"`
	TAXVALUE            string      `json:"TAX_VALUE"`
	LEADID              string      `json:"LEAD_ID"`
	COMPANYID           string      `json:"COMPANY_ID"`
	CONTACTID           string      `json:"CONTACT_ID"`
	QUOTEID             interface{} `json:"QUOTE_ID"`
	BEGINDATE           time.Time   `json:"BEGINDATE"`
	CLOSEDATE           time.Time   `json:"CLOSEDATE"`
	ASSIGNEDBYID        string      `json:"ASSIGNED_BY_ID"`
	CREATEDBYID         string      `json:"CREATED_BY_ID"`
	MODIFYBYID          string      `json:"MODIFY_BY_ID"`
	DATECREATE          time.Time   `json:"DATE_CREATE"`
	DATEMODIFY          time.Time   `json:"DATE_MODIFY"`
	OPENED              string      `json:"OPENED"`
	CLOSED              string      `json:"CLOSED"`
	COMMENTS            string      `json:"COMMENTS"`
	ADDITIONALINFO      interface{} `json:"ADDITIONAL_INFO"`
	LOCATIONID          interface{} `json:"LOCATION_ID"`
	CATEGORYID          string      `json:"CATEGORY_ID"`
	STAGESEMANTICID     string      `json:"STAGE_SEMANTIC_ID"`
	ISNEW               string      `json:"IS_NEW"`
	ISRECURRING         string      `json:"IS_RECURRING"`
	ISRETURNCUSTOMER    string      `json:"IS_RETURN_CUSTOMER"`
	ISREPEATEDAPPROACH  string      `json:"IS_REPEATED_APPROACH"`
	SOURCEID            string      `json:"SOURCE_ID"`
	SOURCEDESCRIPTION   string      `json:"SOURCE_DESCRIPTION"`
	ORIGINATORID        interface{} `json:"ORIGINATOR_ID"`
	ORIGINID            interface{} `json:"ORIGIN_ID"`
	MOVEDBYID           string      `json:"MOVED_BY_ID"`
	MOVEDTIME           time.Time   `json:"MOVED_TIME"`
	UTMSOURCE           interface{} `json:"UTM_SOURCE"`
	UTMMEDIUM           interface{} `json:"UTM_MEDIUM"`
	UTMCAMPAIGN         interface{} `json:"UTM_CAMPAIGN"`
	UTMCONTENT          interface{} `json:"UTM_CONTENT"`
	UTMTERM             interface{} `json:"UTM_TERM"`
}
