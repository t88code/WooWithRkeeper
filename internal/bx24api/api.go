package bx24api

import (
	"WooWithRkeeper/internal/bx24api/models"
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

//TODO улучшить логирование Info и Debug

type BX24API interface {
	ProductGet(ID int) (*models.Product, error)
	ProductList(opts ...models.Options) ([]*models.Product, error)
	ProductAdd(Name string, fields ...models.Field) (int, error)
	ProductUpdate(ID int, fields ...models.Field) error
	ProductDel(ID int) error

	ProductSectionGet(ID int) (*models.ProductSection, error)
	ProductSectionList() ([]*models.ProductSection, error)
	ProductSectionAdd(Name string, fields ...models.Field) (int, error)
	ProductSectionUpdate(ID int, fields ...models.Field) error
	ProductSectionDelete(ID int) error

	DealGet(ID int) (*models.Deal, error)
	DealList(opts ...models.Options) ([]*models.Deal, error)
	DealUpdate(ID int, fields ...models.Field) error
	DealAdd(fields ...models.Field) (int, error)

	ProductRowsGet(ID int) ([]*models.ProductRow, error)
	ProductRowsSet(ID int, rows ...models.Row) error

	ContactGet(ID int) (*models.Contact, error)
}

type bx24api struct {
	url string
}

func (b *bx24api) ProductGet(ID int) (*models.Product, error) {

	logger := logging.GetLogger()
	logger.Println("ProductGet:>Start")
	defer logger.Println("ProductGet:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.product.get.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	params.Add("id", fmt.Sprint(ID))
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductGet := new(models.ProductGet)
	err = json.Unmarshal(respBody, ProductGet)
	if err != nil {
		return nil, err
	}

	if ProductGet.ErrorText != "" || ProductGet.ErrorDescription != "" {
		return nil, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductGet.ErrorDescription, ProductGet.ErrorText))
	}

	return ProductGet.Result, nil
}

func (b *bx24api) ProductList(opts ...models.Options) ([]*models.Product, error) {

	logger := logging.GetLogger()
	logger.Println("ProductList:>Start")
	defer logger.Println("ProductList:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.product.list.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	params.Add("order[ID]", "ASC")

	//add fields is BEAUTIFUL!!
	Option := new(models.OptionsStruct)
	for _, opt := range opts {
		opt(Option)
		params.Add(Option.Key, Option.Value)
	}
	params.Add("filter[>ID]", "0")

	var AllProducts []*models.Product

	for {

		req.URL.RawQuery = params.Encode()
		logger.Debugf("RawQuery: %s", req.URL.RawQuery)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Println(err)
			}
		}(resp.Body)

		respBody, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		//logger.Debugf("Response:\n%s", respBody)

		ProductList := new(models.ProductList)
		err = json.Unmarshal(respBody, ProductList)
		if err != nil {
			return nil, err
		}

		if ProductList.ErrorText != "" || ProductList.ErrorDescription != "" {
			return nil, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductList.ErrorDescription, ProductList.ErrorText))
		}

		if len(ProductList.Result) == 0 {
			break
		} else {
			logger.Debugf("Product count: %d", len(ProductList.Result))
		}

		AllProducts = append(AllProducts, ProductList.Result...)

		productIDBX24 := ProductList.Result[len(ProductList.Result)-1].ID
		params.Set("filter[>ID]", productIDBX24)

	}
	logger.Infof("AllProducts len = %d", len(AllProducts))

	return AllProducts, nil
}

func (b *bx24api) ProductAdd(Name string, fields ...models.Field) (int, error) {

	logger := logging.GetLogger()
	logger.Println("ProductAdd:>Start")
	defer logger.Println("ProductAdd:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.product.add.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	params := req.URL.Query()
	params.Add("fields[NAME]", Name)

	//add fields is BEAUTIFUL!!
	Field := new(models.FieldStruct)
	for _, field := range fields {
		field(Field)
		params.Add(Field.Key, Field.Value)
	}
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return 0, err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductAdd := new(models.ProductAdd)
	err = json.Unmarshal(respBody, ProductAdd)
	if err != nil {
		return 0, err
	}

	if ProductAdd.ErrorText != "" || ProductAdd.ErrorDescription != "" {
		return 0, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductAdd.ErrorDescription, ProductAdd.ErrorText))
	}

	return ProductAdd.Result, nil
}

func (b *bx24api) ProductUpdate(ID int, fields ...models.Field) error {

	logger := logging.GetLogger()
	logger.Println("ProductUpdate:>Start")
	defer logger.Println("ProductUpdate:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.product.update.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	params := req.URL.Query()
	params.Add("ID", fmt.Sprint(ID))

	//add fields is BEAUTIFUL!!
	Field := new(models.FieldStruct)
	for _, field := range fields {
		field(Field)
		params.Add(Field.Key, Field.Value)
	}
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err) //return надо сделать в другую функцию в основную
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductUpdate := new(models.ProductUpdate)
	err = json.Unmarshal(respBody, ProductUpdate)
	if err != nil {
		return err
	}

	if ProductUpdate.ErrorText != "" || ProductUpdate.ErrorDescription != "" {
		return errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductUpdate.ErrorDescription, ProductUpdate.ErrorText))
	}

	return nil
}

func (b *bx24api) ProductDel(ID int) error {

	logger := logging.GetLogger()
	logger.Println("ProductDel:>Start")
	defer logger.Println("ProductDel:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.product.delete.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	params := req.URL.Query()
	params.Add("id", fmt.Sprint(ID))
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductDelete := new(models.ProductDelete)
	err = json.Unmarshal(respBody, ProductDelete)
	if err != nil {
		return err
	}

	if ProductDelete.ErrorText != "" || ProductDelete.ErrorDescription != "" {
		return errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductDelete.ErrorDescription, ProductDelete.ErrorText))
	}

	return nil
}

func (b *bx24api) ProductSectionGet(ID int) (*models.ProductSection, error) {

	logger := logging.GetLogger()
	logger.Println("ProductSectionGet:>Start")
	defer logger.Println("ProductSectionGet:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.productsection.get.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	params.Add("id", fmt.Sprint(ID))
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductSectionGet := new(models.ProductSectionGet)
	err = json.Unmarshal(respBody, ProductSectionGet)
	if err != nil {
		return nil, err
	}

	if ProductSectionGet.ErrorText != "" || ProductSectionGet.ErrorDescription != "" {
		return nil, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductSectionGet.ErrorDescription, ProductSectionGet.ErrorText))
	}

	return ProductSectionGet.Result, nil
}

func (b *bx24api) ProductSectionList() ([]*models.ProductSection, error) {

	logger := logging.GetLogger()
	logger.Println("ProductSectionList:>Start")
	defer logger.Println("ProductSectionList:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.productsection.list.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductSectionList := new(models.ProductSectionList)
	err = json.Unmarshal(respBody, ProductSectionList)
	if err != nil {
		return nil, err
	}

	if ProductSectionList.Error != "" || ProductSectionList.ErrorDescription != "" {
		return nil, errors.New(fmt.Sprintf("error: %v, error_description: %v", ProductSectionList.Error, ProductSectionList.ErrorDescription))
	}

	return ProductSectionList.Result, nil
}

func (b *bx24api) ProductSectionAdd(Name string, fields ...models.Field) (int, error) {

	logger := logging.GetLogger()
	logger.Println("ProductSectionAdd:>Start")
	defer logger.Println("ProductSectionAdd:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.productsection.add.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	params := req.URL.Query()
	params.Add("fields[NAME]", Name)

	//add fields is BEAUTIFUL!!
	Field := new(models.FieldStruct)
	for _, field := range fields {
		field(Field)
		params.Add(Field.Key, Field.Value)
	}
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return 0, err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductSectionAdd := new(models.ProductSectionAdd)
	err = json.Unmarshal(respBody, ProductSectionAdd)
	if err != nil {
		return 0, err
	}

	if ProductSectionAdd.ErrorText != "" || ProductSectionAdd.ErrorDescription != "" {
		return 0, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductSectionAdd.ErrorDescription, ProductSectionAdd.ErrorText))
	}

	return ProductSectionAdd.Result, nil
}

func (b *bx24api) ProductSectionUpdate(ID int, fields ...models.Field) error {

	logger := logging.GetLogger()
	logger.Println("ProductSectionUpdate:>Start")
	defer logger.Println("ProductSectionUpdate:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.productsection.update.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	params := req.URL.Query()
	params.Add("ID", fmt.Sprint(ID))

	//add fields is BEAUTIFUL!!
	Field := new(models.FieldStruct)
	for _, field := range fields {
		field(Field)
		params.Add(Field.Key, Field.Value)
	}
	req.URL.RawQuery = params.Encode()

	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductSectionUpdate := new(models.ProductSectionUpdate)
	err = json.Unmarshal(respBody, ProductSectionUpdate)
	if err != nil {
		return err
	}

	if ProductSectionUpdate.ErrorText != "" || ProductSectionUpdate.ErrorDescription != "" {
		return errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductSectionUpdate.ErrorDescription, ProductSectionUpdate.ErrorText))
	}

	return nil
}

func (b *bx24api) ProductSectionDelete(ID int) error {

	logger := logging.GetLogger()
	logger.Println("ProductSectionDelete:>Start")
	defer logger.Println("ProductSectionDelete:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.productsection.delete.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	params := req.URL.Query()
	params.Add("id", fmt.Sprint(ID))
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductSectionDelete := new(models.ProductSectionDelete)
	err = json.Unmarshal(respBody, ProductSectionDelete)
	if err != nil {
		return err
	}

	if ProductSectionDelete.ErrorText != "" || ProductSectionDelete.ErrorDescription != "" {
		return errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductSectionDelete.ErrorDescription, ProductSectionDelete.ErrorText))
	}

	return nil
}

func (b *bx24api) DealGet(ID int) (*models.Deal, error) {
	logger := logging.GetLogger()
	logger.Println("DealGet:>Start")
	defer logger.Println("DealGet:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.deal.get.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	params.Add("id", fmt.Sprint(ID))
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Response:\n%s", respBody)

	DealGet := new(models.DealGet)
	err = json.Unmarshal(respBody, DealGet)
	if err != nil {
		return nil, err
	}

	if DealGet.ErrorText != "" || DealGet.ErrorDescription != "" {
		return nil, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", DealGet.ErrorDescription, DealGet.ErrorText))
	}

	return DealGet.Result, nil
}

func (b *bx24api) DealList(opts ...models.Options) ([]*models.Deal, error) {
	logger := logging.GetLogger()
	logger.Println("DealList:>Start")
	defer logger.Println("DealList:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.deal.list.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()

	//add fields is BEAUTIFUL!!
	Option := new(models.OptionsStruct)
	for _, opt := range opts {
		opt(Option)
		params.Add(Option.Key, Option.Value)
	}
	req.URL.RawQuery = params.Encode()

	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Info(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Response:\n%s", respBody)

	DealList := new(models.DealList)
	err = json.Unmarshal(respBody, DealList)
	if err != nil {
		return nil, err
	}

	if DealList.ErrorText != "" || DealList.ErrorDescription != "" {
		return nil, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", DealList.ErrorDescription, DealList.ErrorText))
	}

	return DealList.Result, nil
}

func (b *bx24api) DealUpdate(ID int, fields ...models.Field) error {
	logger := logging.GetLogger()
	logger.Println("DealUpdate:>Start")
	defer logger.Println("DealUpdate:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.deal.update.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	params := req.URL.Query()
	params.Add("ID", fmt.Sprint(ID))

	//add fields is BEAUTIFUL!!
	Field := new(models.FieldStruct)
	for _, field := range fields {
		field(Field)
		params.Add(Field.Key, Field.Value)
	}
	req.URL.RawQuery = params.Encode()

	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}

	logger.Debugf("Response:\n%s", respBody)

	DealUpdate := new(models.DealUpdate)
	err = json.Unmarshal(respBody, DealUpdate)
	if err != nil {
		return err
	}

	if DealUpdate.ErrorText != "" || DealUpdate.ErrorDescription != "" {
		return errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", DealUpdate.ErrorDescription, DealUpdate.ErrorText))
	}

	return nil
}

func (b *bx24api) DealAdd(fields ...models.Field) (int, error) {

	logger := logging.GetLogger()
	logger.Println("DealAdd:>Start")
	defer logger.Println("DealAdd:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.deal.add.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	params := req.URL.Query()

	//add fields is BEAUTIFUL!!
	Field := new(models.FieldStruct)
	for _, field := range fields {
		field(Field)
		params.Add(Field.Key, Field.Value)
	}
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Info(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return 0, err
	}

	logger.Debugf("Response:\n%s", respBody)

	DealAdd := new(models.DealAdd)
	err = json.Unmarshal(respBody, DealAdd)
	if err != nil {
		return 0, err
	}

	if DealAdd.ErrorText != "" || DealAdd.ErrorDescription != "" {
		return 0, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", DealAdd.ErrorDescription, DealAdd.ErrorText))
	}

	return DealAdd.Result, nil
}

func (b *bx24api) ProductRowsGet(ID int) ([]*models.ProductRow, error) {

	logger := logging.GetLogger()
	logger.Println("ProductRowsGet:>Start")
	defer logger.Println("ProductRowsGet:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.deal.productrows.get.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	params.Add("id", fmt.Sprint(ID))
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductRowsGet := new(models.ProductRowsGet)
	err = json.Unmarshal(respBody, ProductRowsGet)
	if err != nil {
		return nil, err
	}

	if ProductRowsGet.ErrorText != "" || ProductRowsGet.ErrorDescription != "" {
		return nil, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductRowsGet.ErrorDescription, ProductRowsGet.ErrorText))
	}

	return ProductRowsGet.Result, nil
}

func (b *bx24api) ProductRowsSet(ID int, rows ...models.Row) error {

	logger := logging.GetLogger()
	logger.Println("ProductRowsSet:>Start")
	defer logger.Println("ProductRowsSet:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.deal.productrows.set.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	params := req.URL.Query()
	params.Add("ID", fmt.Sprint(ID))

	//add fields is BEAUTIFUL!!
	Row := new(models.RowStruct)
	for i, row := range rows {
		row(Row)
		params.Add(fmt.Sprintf("rows[%d][PRODUCT_ID]", i), fmt.Sprint(Row.ProductID))
		params.Add(fmt.Sprintf("rows[%d][PRICE]", i), fmt.Sprint(Row.Price))
		params.Add(fmt.Sprintf("rows[%d][QUANTITY]", i), fmt.Sprint(Row.Quantity))
	}
	req.URL.RawQuery = params.Encode()
	logger.Debugf("URL:\n%s", req.URL)
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)
	logger.Debugf("RawQueryMap:\n%s", req.URL.Query())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Info(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}

	logger.Debugf("Response:\n%s", respBody)

	ProductRowsSet := new(models.ProductRowsSet)
	err = json.Unmarshal(respBody, ProductRowsSet)
	if err != nil {
		return err
	}

	if ProductRowsSet.ErrorText != "" || ProductRowsSet.ErrorDescription != "" {
		return errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ProductRowsSet.ErrorDescription, ProductRowsSet.ErrorText))
	}

	return nil
}

func (b *bx24api) ContactGet(ID int) (*models.Contact, error) {
	logger := logging.GetLogger()
	logger.Println("ContactGet:>Start")
	defer logger.Println("ContactGet:>End")

	url := fmt.Sprintf("%s/%s", b.url, "crm.contact.get.json")
	logger.Debugf("Request:\n%s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	params.Add("id", fmt.Sprint(ID))
	req.URL.RawQuery = params.Encode()
	logger.Debugf("RawQuery: %s", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Println(err)
		}
	}(resp.Body)

	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Response:\n%s", respBody)

	ContactGet := new(models.ContactGet)
	err = json.Unmarshal(respBody, ContactGet)
	if err != nil {
		return nil, err
	}

	if ContactGet.ErrorText != "" || ContactGet.ErrorDescription != "" {
		return nil, errors.New(fmt.Sprintf("API BX24: error_description: %s; error: %s", ContactGet.ErrorDescription, ContactGet.ErrorText))
	}

	return ContactGet.Result, nil
}

func NewAPI(url string) BX24API {
	return &bx24api{
		url: url,
	}
}
