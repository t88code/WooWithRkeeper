package models

import (
	"WooWithRkeeper/internal/config"
	"fmt"
)

type FieldStruct struct {
	Key   string
	Value string
}

type Field func(*FieldStruct)

func Name(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "fields[NAME]"
		f.Value = value
	}
}

//field(Field)
//(Name(value string) func(*FieldStruct))(Field)

func Active(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "fields[ACTIVE]"
		f.Value = value
	}
}

func XMLID(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "fields[XML_ID]"
		f.Value = value
	}
}

func SectionID(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "fields[SECTION_ID]"
		f.Value = value
	}
}

func Price(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "fields[PRICE]"
		f.Value = value
	}
}

func VISITID(value string) Field {
	return func(f *FieldStruct) {
		cfg := config.GetConfig()
		f.Key = fmt.Sprintf("fields[%s]", cfg.BX24.FieldVISITID)
		f.Value = value
	}
}

func ORDERNAME(value string) Field {
	return func(f *FieldStruct) {
		cfg := config.GetConfig()
		f.Key = fmt.Sprintf("fields[%s]", cfg.BX24.FieldORDERNAME)
		f.Value = value
	}
}

func TITLE(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "fields[TITLE]"
		f.Value = value
	}
}

func STAGEID(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "fields[STAGE_ID]"
		f.Value = value
	}
}
