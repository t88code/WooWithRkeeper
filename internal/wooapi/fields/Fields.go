package fields

type FieldStruct struct {
	Key   string
	Value string
}

type Field func(*FieldStruct)

func Name(value string) Field {
	return func(f *FieldStruct) {
		f.Key = "Name"
		f.Value = value
	}
}
