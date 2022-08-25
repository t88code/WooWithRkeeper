package farcards

import "encoding/xml"

type Transaction struct {
	XMLName           xml.Name `xml:"CHECK"`
	Stationcode       int      `xml:"stationcode,attr"`
	Restaurantcode    int      `xml:"restaurantcode,attr"`
	Cashservername    string   `xml:"cashservername,attr"`
	Generateddatetime string   `xml:"generateddatetime,attr"`
	Chmode            int      `xml:"chmode,attr"`
	Locale            int      `xml:"locale,attr"`
	Shiftdate         string   `xml:"shiftdate,attr"`
	Shiftnum          int      `xml:"shiftnum,attr"`
	EXTINFO           struct {
		Reservation string `xml:"reservation,attr"`
		INTERFACES  struct {
			Current   int `xml:"current,attr"`
			INTERFACE struct {
				Type      string `xml:"type,attr"`
				ID        int    `xml:"id,attr"`
				Mode      int    `xml:"mode,attr"`
				Interface int    `xml:"interface,attr"`
				HOLDERS   struct {
					ITEM struct {
						Cardcode int `xml:"cardcode,attr"`
					} `xml:"ITEM"`
				} `xml:"HOLDERS"`
				ALLCARDS string `xml:"ALLCARDS"`
			} `xml:"INTERFACE"`
		} `xml:"INTERFACES"`
	} `xml:"EXTINFO"`
	CHECKDATA struct {
		Checknum          int    `xml:"checknum,attr"`
		Printnum          int    `xml:"printnum,attr"`
		Fiscdocnum        int    `xml:"fiscdocnum,attr"`
		Delprintnum       int    `xml:"delprintnum,attr"`
		Delfiscdocnum     int    `xml:"delfiscdocnum,attr"`
		Extfiscid         string `xml:"extfiscid,attr"`
		Tablename         int    `xml:"tablename,attr"`
		Startservice      string `xml:"startservice,attr"` //TODO добавить обработку времени
		Closedatetime     string `xml:"closedatetime,attr"`
		Ordernum          string `xml:"ordernum,attr"`
		Guests            int    `xml:"guests,attr"`
		Orderguid         string `xml:"orderguid,attr"`
		Checkguid         string `xml:"checkguid,attr"`
		OrderCat          int    `xml:"order_cat,attr"`
		OrderType         int    `xml:"order_type,attr"`
		Persistentcomment string `xml:"persistentcomment,attr"`
		CHECKPERSONS      struct {
			Count  int `xml:"count,attr"`
			PERSON struct {
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Code int    `xml:"code,attr"`
				Role int    `xml:"role,attr"`
			} `xml:"PERSON"`
		} `xml:"CHECKPERSONS"`
		CHECKLINES struct {
			Count int `xml:"count,attr"`
			LINE  []struct {
				ID          string `xml:"id,attr"`
				Code        int    `xml:"code,attr"`
				Name        string `xml:"name,attr"`
				Uni         int    `xml:"uni,attr"`
				Type        string `xml:"type,attr"`
				Price       int    `xml:"price,attr"`
				PrListSum   int    `xml:"pr_list_sum,attr"`
				CategID     string `xml:"categ_id,attr"`
				ServprintID string `xml:"servprint_id,attr"`
				Quantity    int    `xml:"quantity,attr"`
				Sum         int    `xml:"sum,attr"`
				LINETAXES   struct {
					Count int `xml:"count,attr"`
					TAX   struct {
						ID  string `xml:"id,attr"`
						Sum string `xml:"sum,attr"`
					} `xml:"TAX"`
				} `xml:"LINETAXES"`
				LINEPAYMENTS struct {
					LINEPAYMENT struct {
						ID  string `xml:"id,attr"`
						Sum int    `xml:"sum,attr"`
					} `xml:"LINEPAYMENT"`
				} `xml:"LINEPAYMENTS"`
			} `xml:"LINE"`
		} `xml:"CHECKLINES"`
		CHECKCATEGS struct {
			Count int `xml:"count,attr"`
			CATEG struct {
				ID      string `xml:"id,attr"`
				Code    int    `xml:"code,attr"`
				Name    string `xml:"name,attr"`
				Sum     int    `xml:"sum,attr"`
				Discsum int    `xml:"discsum,attr"`
			} `xml:"CATEG"`
		} `xml:"CHECKCATEGS"`
		CHECKPAYMENTS struct {
			Count   int `xml:"count,attr"`
			PAYMENT struct {
				ID      string `xml:"id,attr"`
				Code    int    `xml:"code,attr"`
				Name    string `xml:"name,attr"`
				Uni     int    `xml:"uni,attr"`
				Paytype int    `xml:"paytype,attr"`
				Bsum    int    `xml:"bsum,attr"`
				Sum     int    `xml:"sum,attr"`
			} `xml:"PAYMENT"`
		} `xml:"CHECKPAYMENTS"`
		CHECKTAXES struct {
			Count int `xml:"count,attr"`
			TAX   struct {
				ID   string `xml:"id,attr"`
				Code int    `xml:"code,attr"`
				Rate int    `xml:"rate,attr"`
				Sum  string `xml:"sum,attr"`
				Name string `xml:"name,attr"`
			} `xml:"TAX"`
		} `xml:"CHECKTAXES"`
	} `xml:"CHECKDATA"`
}
