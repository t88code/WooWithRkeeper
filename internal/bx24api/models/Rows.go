package models

type RowStruct struct {
	ProductID int
	Price     string
	Quantity  int
}

type Row func(*RowStruct)

func PRODUCT(id int, price string, quantity int) Row {
	return func(row *RowStruct) {
		row.ProductID = id
		row.Price = price
		row.Quantity = quantity
	}
}
