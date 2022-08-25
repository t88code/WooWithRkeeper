package models

type FieldMenuitemItem func(*MenuitemItem)

func ID_BX24(value int) FieldMenuitemItem {
	return func(item *MenuitemItem) {
		item.ID_BX24 = value
	}
}

func SectionID_BX24(value int) FieldMenuitemItem {
	return func(item *MenuitemItem) {
		item.SectionID_BX24 = value
	}
}

func Ident(value int) FieldMenuitemItem {
	return func(item *MenuitemItem) {
		item.Ident = value
	}
}
