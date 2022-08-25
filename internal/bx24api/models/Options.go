package models

import "fmt"

type OptionsStruct struct {
	Key   string
	Value string
}

type Options func(*OptionsStruct)

func Filter(key string, value string) Options {
	return func(f *OptionsStruct) {
		f.Key = fmt.Sprintf("filter[%s]", key)
		f.Value = value
	}
}

func Order(key string, value string) Options {
	return func(f *OptionsStruct) {
		f.Key = fmt.Sprintf("order[%s]", key)
		f.Value = value
	}
}
