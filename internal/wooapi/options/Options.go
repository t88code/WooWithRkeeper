package options

import (
	"strconv"
)

type OptionStruct struct {
	Key   string
	Value string
}

type Option func(*OptionStruct)

func Page(value int) Option {
	return func(f *OptionStruct) {
		f.Key = "page"
		f.Value = strconv.Itoa(value)
	}
}

func PerPage(value int) Option {
	return func(f *OptionStruct) {
		f.Key = "per_page"
		f.Value = strconv.Itoa(value)
	}
}

func Force(value bool) Option {
	return func(f *OptionStruct) {
		f.Key = "force"
		if value {
			f.Value = "true"
		} else {
			f.Value = "false"
		}
	}
}

func Search(value string) Option {
	return func(f *OptionStruct) {
		f.Key = "search"
		f.Value = value
	}
}
