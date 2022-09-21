package models

type Author struct {
	ID   string `xml:"id,attr,omitempty"`
	Code string `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
	Role *Role  `xml:"Role"`
}

type Creator struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
	Role *Role  `xml:"Role"`
}

type Waiter struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
	Role *Role  `xml:"Role"`
}

type Role struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type OrderCategory struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type OrderType struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type Table struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type Guests struct {
	Count int      `xml:"count,attr,omitempty"`
	Guest *[]Guest `xml:"Guest"`
}

type Guest struct {
	GuestLabel string     `xml:"guestLabel,attr,omitempty"`
	CardCode   string     `xml:"cardCode,attr,omitempty"`
	ClientID   int64      `xml:"clientID,attr,omitempty"`
	AddressID  int64      `xml:"addressID,attr,omitempty"`
	Interface  *Interface `xml:"Interface"`
}

type Interface struct {
	Code int    `xml:"code,attr,omitempty"`
	ID   string `xml:"id,attr,omitempty"`
}

type ExternalProps struct {
	Prop []*Prop `xml:"Prop"`
}

type Prop struct {
	Name  string `xml:"name,attr,omitempty"`
	Value string `xml:"value,attr,omitempty"`
}

type Dish struct {
	ID          int    `xml:"id,attr,omitempty"`
	Code        int    `xml:"code,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
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
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type TradeGroup struct {
	ID   int    `xml:"id,attr,omitempty"`
	Code int    `xml:"code,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

type Session struct {
	Uni          int         `xml:"uni,attr,omitempty"`
	LineGuid     string      `xml:"line_guid,attr,omitempty"`
	State        int         `xml:"state,attr,omitempty"`
	SessionID    int         `xml:"sessionID,attr,omitempty"`
	IsDraft      int         `xml:"isDraft,attr,omitempty"`
	RemindTime   string      `xml:"remindTime,attr,omitempty"`
	StartService string      `xml:"startService,attr,omitempty"`
	Printed      int         `xml:"printed,attr,omitempty"`
	CookMins     int         `xml:"cookMins,attr,omitempty"`
	Station      *Station    `xml:"Station"`
	Author       *Author     `xml:"Author"`
	Creator      *Creator    `xml:"Creator"`
	Dish         []*Dish     `xml:"Dish"`
	PriceScale   *PriceScale `xml:"PriceScale"`
	TradeGroup   *TradeGroup `xml:"TradeGroup"`
}

type Order struct {
	Visit                int            `xml:"visit,attr,omitempty"`
	OrderIdent           int            `xml:"orderIdent,attr,omitempty"`
	Guid                 string         `xml:"guid,attr,omitempty"`
	URL                  string         `xml:"url,attr,omitempty"`
	OrderName            string         `xml:"orderName,attr,omitempty"`
	Version              int            `xml:"version,attr,omitempty"`
	Crc32                string         `xml:"crc32,attr,omitempty"`
	OrderSum             int            `xml:"orderSum,attr,omitempty"`
	UnpaidSum            int            `xml:"unpaidSum,attr,omitempty"`
	DiscountSum          int            `xml:"discountSum,attr,omitempty"`
	TotalPieces          string         `xml:"totalPieces,attr,omitempty"`
	SeqNumber            int            `xml:"seqNumber,attr,omitempty"`
	Paid                 int            `xml:"paid,attr,omitempty"`
	Finished             int            `xml:"finished,attr,omitempty"`
	PersistentComment    string         `xml:"persistentComment,attr,omitempty"`
	NonPersistentComment string         `xml:"nonPersistentComment,attr,omitempty"`
	OpenTime             string         `xml:"openTime,attr,omitempty"`
	CookMins             int            `xml:"cookMins,attr,omitempty"`
	Creator              *Creator       `xml:"Creator"`
	Waiter               *Waiter        `xml:"Waiter"`
	OrderCategory        *OrderCategory `xml:"OrderCategory"`
	OrderType            *OrderType     `xml:"OrderType"`
	Table                *Table         `xml:"Table"`
	Station              *Station       `xml:"Station"`
	ExternalProps        *ExternalProps `xml:"ExternalProps"`
	Session              []Session      `xml:"Session,omitempty"`
}

type Prepay struct {
	Code               int        `xml:"code,attr,omitempty"`
	ID                 int        `xml:"id,attr,omitempty"`
	Guid               string     `xml:"guid,attr,omitempty"`
	Amount             int        `xml:"amount,attr,omitempty"`
	Deleted            int        `xml:"deleted,attr,omitempty"`
	Promised           string     `xml:"promised,attr,omitempty"`
	LineGuid           string     `xml:"line_guid,attr,omitempty"`
	CardCode           string     `xml:"cardCode,attr,omitempty"`
	ExtTransactionInfo string     `xml:"extTransactionInfo,attr,omitempty"`
	Interface          *Interface `xml:"Interface"`
}
