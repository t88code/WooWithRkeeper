package models

type Author struct {
	ID   string `xml:"id,attr"`
	Code string `xml:"code,attr"`
	Name string `xml:"name,attr"`
	Role *Role  `xml:"Role"`
}

type Creator struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
	Role *Role  `xml:"Role"`
}

type Waiter struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
	Role *Role  `xml:"Role"`
}

type Role struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
}

type OrderCategory struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
}

type OrderType struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
}

type Table struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
}

type Guests struct {
	Count int      `xml:"count,attr"`
	Guest *[]Guest `xml:"Guest"`
}

type Guest struct {
	GuestLabel string     `xml:"guestLabel,attr"`
	CardCode   string     `xml:"cardCode,attr"`
	ClientID   int64      `xml:"clientID,attr"`
	AddressID  int64      `xml:"addressID,attr"`
	Interface  *Interface `xml:"Interface"` //TODO не точно и требует проверки
}

type Interface struct {
	Code int `xml:"code,attr"`
}

type ExternalProps struct {
	Prop []*Prop `xml:"Prop"`
}

type Prop struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Dish struct {
	ID          int    `xml:"id,attr,omitempty"`
	Code        int    `xml:"code,attr,omitempty"`
	Name        string `xml:"name,attr"`
	Uni         string `xml:"uni,attr,omitempty"`
	LineGuid    string `xml:"line_guid,attr,omitempty"`
	State       string `xml:"state,attr,omitempty"`
	Guid        string `xml:"guid,attr,omitempty"`
	Price       int    `xml:"price,attr,omitempty"`
	Amount      int    `xml:"amount,attr,omitempty"`
	Quantity    int    `xml:"quantity,attr,omitempty"`
	SrcQuantity string `xml:"srcQuantity,attr,omitempty"`
}

type Station struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type PriceScale struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
}

type TradeGroup struct {
	ID   int    `xml:"id,attr"`
	Code int    `xml:"code,attr"`
	Name string `xml:"name,attr"`
}

type Session struct {
	Uni          int         `xml:"uni,attr"`
	LineGuid     string      `xml:"line_guid,attr"`
	State        int         `xml:"state,attr"`
	SessionID    int         `xml:"sessionID,attr"`
	IsDraft      int         `xml:"isDraft,attr"`
	RemindTime   string      `xml:"remindTime,attr"`
	StartService string      `xml:"startService,attr"`
	Printed      int         `xml:"printed,attr"`
	CookMins     int         `xml:"cookMins,attr"`
	Station      *Station    `xml:"Station"`
	Author       *Author     `xml:"Author"`
	Creator      *Creator    `xml:"Creator"`
	Dish         []*Dish     `xml:"Dish"`
	PriceScale   *PriceScale `xml:"PriceScale"`
	TradeGroup   *TradeGroup `xml:"TradeGroup"`
}

type Order struct {
	Visit                int            `xml:"visit,attr"`
	OrderIdent           int            `xml:"orderIdent,attr"`
	Guid                 string         `xml:"guid,attr"`
	URL                  string         `xml:"url,attr"`
	OrderName            string         `xml:"orderName,attr"`
	Version              int            `xml:"version,attr"`
	Crc32                string         `xml:"crc32,attr"`
	OrderSum             int            `xml:"orderSum,attr"`
	UnpaidSum            int            `xml:"unpaidSum,attr"`
	DiscountSum          int            `xml:"discountSum,attr"`
	TotalPieces          string         `xml:"totalPieces,attr"`
	SeqNumber            int            `xml:"seqNumber,attr"`
	Paid                 int            `xml:"paid,attr"`
	Finished             int            `xml:"finished,attr"`
	PersistentComment    string         `xml:"persistentComment,attr"`
	NonPersistentComment string         `xml:"nonPersistentComment,attr"`
	OpenTime             string         `xml:"openTime,attr"`
	CookMins             int            `xml:"cookMins,attr"`
	Creator              *Creator       `xml:"Creator"`
	Waiter               *Waiter        `xml:"Waiter"`
	OrderCategory        *OrderCategory `xml:"OrderCategory"`
	OrderType            *OrderType     `xml:"OrderType"`
	Table                *Table         `xml:"Table"`
	Station              *Station       `xml:"Station"`
	ExternalProps        *ExternalProps `xml:"ExternalProps"`
	Session              []Session      `xml:"Session,omitempty"`
}

type LicenseInfo struct {
	Anchor          string `xml:"anchor,attr,omitempty"`
	LicenseToken    string `xml:"licenseToken,attr,omitempty"`
	LicenseInstance struct {
		Guid      string `xml:"guid,attr,omitempty"`
		SeqNumber int    `xml:"seqNumber,attr"`
	} `xml:"LicenseInstance"`
}

type Prepay struct {
	Code               int    `xml:"code,attr,omitempty"`
	ID                 int    `xml:"id,attr,omitempty"`
	Amount             int    `xml:"amount,attr,omitempty"`
	ExtTransactionInfo string `xml:"ExtTransactionInfo,attr,omitempty"`
	CardCode           string `xml:"cardCode,attr,omitempty"`
	Promised           int    `xml:"promised,attr,omitempty"`
}
