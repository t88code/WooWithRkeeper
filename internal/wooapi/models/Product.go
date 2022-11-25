package models

type Product struct {
	ID                int           `json:"id,omitempty,omitempty"`
	Name              string        `json:"name,omitempty"`
	Slug              string        `json:"slug,omitempty"`
	Permalink         string        `json:"permalink,omitempty"`
	DateCreated       string        `json:"date_created,omitempty"`
	DateCreatedGmt    string        `json:"date_created_gmt,omitempty"`
	DateModified      string        `json:"date_modified,omitempty"`
	DateModifiedGmt   string        `json:"date_modified_gmt,omitempty"`
	Type              string        `json:"type,omitempty"`
	Status            string        `json:"status,omitempty"`
	Featured          bool          `json:"featured,omitempty"`
	CatalogVisibility string        `json:"catalog_visibility,omitempty"`
	Description       string        `json:"description,omitempty"`
	ShortDescription  string        `json:"short_description,omitempty"`
	Sku               string        `json:"sku,omitempty"`
	Price             string        `json:"price,omitempty"`
	RegularPrice      string        `json:"regular_price,omitempty"`
	SalePrice         string        `json:"sale_price,omitempty"`
	DateOnSaleFrom    interface{}   `json:"date_on_sale_from,omitempty"`
	DateOnSaleFromGmt interface{}   `json:"date_on_sale_from_gmt,omitempty"`
	DateOnSaleTo      interface{}   `json:"date_on_sale_to,omitempty"`
	DateOnSaleToGmt   interface{}   `json:"date_on_sale_to_gmt,omitempty"`
	OnSale            bool          `json:"on_sale,omitempty"`
	Purchasable       bool          `json:"purchasable,omitempty"`
	TotalSales        int           `json:"total_sales,omitempty"`
	Virtual           bool          `json:"virtual,omitempty"`
	Downloadable      bool          `json:"downloadable,omitempty"`
	Downloads         []interface{} `json:"downloads,omitempty"`
	DownloadLimit     int           `json:"download_limit,omitempty"`
	DownloadExpiry    int           `json:"download_expiry,omitempty"`
	ExternalUrl       string        `json:"external_url,omitempty"`
	ButtonText        string        `json:"button_text,omitempty"`
	TaxStatus         string        `json:"tax_status,omitempty"`
	TaxClass          string        `json:"tax_class,omitempty"`
	ManageStock       bool          `json:"manage_stock,omitempty"`
	StockQuantity     interface{}   `json:"stock_quantity,omitempty"`
	Backorders        string        `json:"backorders,omitempty"`
	BackordersAllowed bool          `json:"backorders_allowed,omitempty"`
	Backordered       bool          `json:"backordered,omitempty"`
	LowStockAmount    interface{}   `json:"low_stock_amount,omitempty"`
	SoldIndividually  bool          `json:"sold_individually,omitempty"`
	Weight            string        `json:"weight,omitempty"`
	Dimensions        *Dimensions   `json:"dimensions,omitempty"`
	ShippingRequired  bool          `json:"shipping_required,omitempty"`
	ShippingTaxable   bool          `json:"shipping_taxable,omitempty"`
	ShippingClass     string        `json:"shipping_class,omitempty"`
	ShippingClassId   int           `json:"shipping_class_id,omitempty"`
	ReviewsAllowed    bool          `json:"reviews_allowed,omitempty"`
	AverageRating     string        `json:"average_rating,omitempty"`
	RatingCount       int           `json:"rating_count,omitempty"`
	UpsellIds         []interface{} `json:"upsell_ids,omitempty"`
	CrossSellIds      []interface{} `json:"cross_sell_ids,omitempty"`
	ParentId          int           `json:"parent_id,omitempty"`
	PurchaseNote      string        `json:"purchase_note,omitempty"`
	Categories        []*Categories `json:"categories,omitempty"`
	Tags              []interface{} `json:"tags,omitempty"`
	//Images            []interface{} `json:"images,omitempty"`
	Images            []ProductImage `json:"images,omitempty"`
	Attributes        []interface{}  `json:"attributes,omitempty"`
	DefaultAttributes []interface{}  `json:"default_attributes,omitempty"`
	Variations        []interface{}  `json:"variations,omitempty"`
	GroupedProducts   []interface{}  `json:"grouped_products,omitempty"`
	MenuOrder         int            `json:"menu_order,omitempty"`
	PriceHtml         string         `json:"price_html,omitempty"`
	RelatedIds        []int          `json:"related_ids,omitempty"`
	MetaData          []MetaData     `json:"meta_data,omitempty"`
	StockStatus       string         `json:"stock_status,omitempty"`
	HasOptions        bool           `json:"has_options,omitempty"`
	Links             *Links         `json:"_links,omitempty"`
}
type ProductImage struct {
	Id              int    `json:"id,omitempty"`
	DateCreated     string `json:"date_created,omitempty"`
	DateCreatedGmt  string `json:"date_created_gmt,omitempty"`
	DateModified    string `json:"date_modified,omitempty"`
	DateModifiedGmt string `json:"date_modified_gmt,omitempty"`
	Src             string `json:"src,omitempty"`
	Name            string `json:"name,omitempty"`
	Alt             string `json:"alt,omitempty"`
}
type Dimensions struct {
	Length string `json:"length,omitempty"`
	Width  string `json:"width,omitempty"`
	Height string `json:"height,omitempty"`
}

type Categories struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type MetaData struct {
	Id    int         `json:"id,omitempty"`
	Key   string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type Links struct {
	Self []struct {
		Href string `json:"href,omitempty"`
	} `json:"self,omitempty"`
	Collection []struct {
		Href string `json:"href,omitempty"`
	} `json:"collection,omitempty"`
	Up []struct {
		Href string `json:"href,omitempty"`
	} `json:"up"`
}
