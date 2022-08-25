package models

type ProductRow struct {
	ID                  string      `json:"ID"`
	OWNERID             string      `json:"OWNER_ID"`
	OWNERTYPE           string      `json:"OWNER_TYPE"`
	PRODUCTID           int         `json:"PRODUCT_ID"`
	PRODUCTNAME         string      `json:"PRODUCT_NAME"`
	ORIGINALPRODUCTNAME string      `json:"ORIGINAL_PRODUCT_NAME"`
	PRODUCTDESCRIPTION  interface{} `json:"PRODUCT_DESCRIPTION"`
	PRICE               int         `json:"PRICE"`
	PRICEEXCLUSIVE      int         `json:"PRICE_EXCLUSIVE"`
	PRICENETTO          int         `json:"PRICE_NETTO"`
	PRICEBRUTTO         int         `json:"PRICE_BRUTTO"`
	PRICEACCOUNT        string      `json:"PRICE_ACCOUNT"`
	QUANTITY            int         `json:"QUANTITY"`
	DISCOUNTTYPEID      int         `json:"DISCOUNT_TYPE_ID"`
	DISCOUNTRATE        int         `json:"DISCOUNT_RATE"`
	DISCOUNTSUM         int         `json:"DISCOUNT_SUM"`
	TAXRATE             int         `json:"TAX_RATE"`
	TAXINCLUDED         string      `json:"TAX_INCLUDED"`
	CUSTOMIZED          string      `json:"CUSTOMIZED"`
	MEASURECODE         int         `json:"MEASURE_CODE"`
	MEASURENAME         string      `json:"MEASURE_NAME"`
	SORT                int         `json:"SORT"`
	RESERVEID           interface{} `json:"RESERVE_ID"`
	RESERVEQUANTITY     int         `json:"RESERVE_QUANTITY"`
}
