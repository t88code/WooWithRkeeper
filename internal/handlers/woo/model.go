package woo

type WebhookCreatOrder struct {
	Id               int    `json:"id"`
	ParentId         int    `json:"parent_id"`
	Status           string `json:"status"`
	Currency         string `json:"currency"`
	Version          string `json:"version"`
	PricesIncludeTax bool   `json:"prices_include_tax"`
	DateCreated      string `json:"date_created"`
	DateModified     string `json:"date_modified"`
	DiscountTotal    string `json:"discount_total"`
	DiscountTax      string `json:"discount_tax"`
	ShippingTotal    string `json:"shipping_total"`
	ShippingTax      string `json:"shipping_tax"`
	CartTax          string `json:"cart_tax"`
	Total            string `json:"total"`
	TotalTax         string `json:"total_tax"`
	CustomerId       int    `json:"customer_id"`
	OrderKey         string `json:"order_key"`
	Billing          struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Company   string `json:"company"`
		Address1  string `json:"address_1"`
		Address2  string `json:"address_2"`
		City      string `json:"city"`
		State     string `json:"state"`
		Postcode  string `json:"postcode"`
		Country   string `json:"country"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	} `json:"billing"`
	Shipping struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Company   string `json:"company"`
		Address1  string `json:"address_1"`
		Address2  string `json:"address_2"`
		City      string `json:"city"`
		State     string `json:"state"`
		Postcode  string `json:"postcode"`
		Country   string `json:"country"`
		Phone     string `json:"phone"`
	} `json:"shipping"`
	PaymentMethod      string      `json:"payment_method"`
	PaymentMethodTitle string      `json:"payment_method_title"`
	TransactionId      string      `json:"transaction_id"`
	CustomerIpAddress  string      `json:"customer_ip_address"`
	CustomerUserAgent  string      `json:"customer_user_agent"`
	CreatedVia         string      `json:"created_via"`
	CustomerNote       string      `json:"customer_note"`
	DateCompleted      interface{} `json:"date_completed"`
	DatePaid           interface{} `json:"date_paid"`
	CartHash           string      `json:"cart_hash"`
	Number             string      `json:"number"`
	MetaData           []struct {
		Id    int         `json:"id"`
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	} `json:"meta_data"`
	LineItems []struct {
		Id          int           `json:"id"`
		Name        string        `json:"name"`
		ProductId   int           `json:"product_id"`
		VariationId int           `json:"variation_id"`
		Quantity    int           `json:"quantity"`
		TaxClass    string        `json:"tax_class"`
		Subtotal    string        `json:"subtotal"`
		SubtotalTax string        `json:"subtotal_tax"`
		Total       string        `json:"total"`
		TotalTax    string        `json:"total_tax"`
		Taxes       []interface{} `json:"taxes"`
		MetaData    []struct {
			Id    int    `json:"id"`
			Key   string `json:"key"`
			Value struct {
				Duration  string `json:"duration,omitempty"`
				TimeStart string `json:"timeStart,omitempty"`
				Summary   string `json:"summary,omitempty"`
				Persons   string `json:"persons,omitempty"`
				Resources []struct {
					TermId          int         `json:"term_id"`
					Name            string      `json:"name"`
					Hidden          bool        `json:"hidden"`
					LimitedResource bool        `json:"limitedResource"`
					Quantity        bool        `json:"quantity"`
					Price           interface{} `json:"price"`
				} `json:"resources,omitempty"`
				Start struct {
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
					Timezone     string `json:"timezone"`
				} `json:"start,omitempty"`
				End struct {
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
					Timezone     string `json:"timezone"`
				} `json:"end,omitempty"`
				Total int `json:"total,omitempty"`
				Items []struct {
					Label string `json:"label"`
					Price int    `json:"price"`
				} `json:"items,omitempty"`
				Enable          string `json:"enable,omitempty"`
				Deposit         int    `json:"deposit,omitempty"`
				Remaining       int    `json:"remaining,omitempty"`
				TaxTotal        int    `json:"tax_total,omitempty"`
				Tax             int    `json:"tax,omitempty"`
				PaymentSchedule struct {
					Unlimited struct {
						Amount int `json:"amount"`
						Tax    int `json:"tax"`
					} `json:"unlimited"`
				} `json:"payment_schedule,omitempty"`
			} `json:"value"`
			DisplayKey   string `json:"display_key"`
			DisplayValue struct {
				Duration  string `json:"duration,omitempty"`
				TimeStart string `json:"timeStart,omitempty"`
				Summary   string `json:"summary,omitempty"`
				Persons   string `json:"persons,omitempty"`
				Resources []struct {
					TermId          int         `json:"term_id"`
					Name            string      `json:"name"`
					Hidden          bool        `json:"hidden"`
					LimitedResource bool        `json:"limitedResource"`
					Quantity        bool        `json:"quantity"`
					Price           interface{} `json:"price"`
				} `json:"resources,omitempty"`
				Start struct {
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
					Timezone     string `json:"timezone"`
				} `json:"start,omitempty"`
				End struct {
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
					Timezone     string `json:"timezone"`
				} `json:"end,omitempty"`
				Total int `json:"total,omitempty"`
				Items []struct {
					Label string `json:"label"`
					Price int    `json:"price"`
				} `json:"items,omitempty"`
				Enable          string `json:"enable,omitempty"`
				Deposit         int    `json:"deposit,omitempty"`
				Remaining       int    `json:"remaining,omitempty"`
				TaxTotal        int    `json:"tax_total,omitempty"`
				Tax             int    `json:"tax,omitempty"`
				PaymentSchedule struct {
					Unlimited struct {
						Amount int `json:"amount"`
						Tax    int `json:"tax"`
					} `json:"unlimited"`
				} `json:"payment_schedule,omitempty"`
			} `json:"display_value"`
		} `json:"meta_data"`
		Sku   string `json:"sku"`
		Price int    `json:"price"`
		Image struct {
			Id  string `json:"id"`
			Src string `json:"src"`
		} `json:"image"`
		ParentName interface{} `json:"parent_name"`
	} `json:"line_items"`
	TaxLines      []interface{} `json:"tax_lines"`
	ShippingLines []struct {
		Id          int           `json:"id"`
		MethodTitle string        `json:"method_title"`
		MethodId    string        `json:"method_id"`
		InstanceId  string        `json:"instance_id"`
		Total       string        `json:"total"`
		TotalTax    string        `json:"total_tax"`
		Taxes       []interface{} `json:"taxes"`
		MetaData    []struct {
			Id           int    `json:"id"`
			Key          string `json:"key"`
			Value        string `json:"value"`
			DisplayKey   string `json:"display_key"`
			DisplayValue string `json:"display_value"`
		} `json:"meta_data"`
	} `json:"shipping_lines"`
	FeeLines         []interface{} `json:"fee_lines"`
	CouponLines      []interface{} `json:"coupon_lines"`
	Refunds          []interface{} `json:"refunds"`
	PaymentUrl       string        `json:"payment_url"`
	IsEditable       bool          `json:"is_editable"`
	NeedsPayment     bool          `json:"needs_payment"`
	NeedsProcessing  bool          `json:"needs_processing"`
	DateCreatedGmt   string        `json:"date_created_gmt"`
	DateModifiedGmt  string        `json:"date_modified_gmt"`
	DateCompletedGmt interface{}   `json:"date_completed_gmt"` //++
	DatePaidGmt      interface{}   `json:"date_paid_gmt"`      //++
	CurrencySymbol   string        `json:"currency_symbol"`
	Links            struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
		Collection []struct {
			Href string `json:"href"`
		} `json:"collection"`
		Customer []struct { //++
			Href string `json:"href"`
		} `json:"customer"`
	} `json:"_links"`
}
