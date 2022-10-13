package models

type ProductCategory struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Parent      int    `json:"parent,omitempty"`
	Description string `json:"description,omitempty"`
	Display     string `json:"display,omitempty"`
	Image       *Image `json:"image,omitempty"`
	MenuOrder   int    `json:"menu_order,omitempty"`
	Count       int    `json:"count,omitempty"`
	Links       *Links
	RkeeperID   int `json:"rkeeperID,omitempty"`
}

type Image struct {
	Src string `json:"src,omitempty"`
}
