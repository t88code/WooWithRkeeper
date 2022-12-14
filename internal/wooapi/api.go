package wooapi

import (
	"WooWithRkeeper/internal/wc-api-go/client"
	"WooWithRkeeper/internal/wc-api-go/options"
	"WooWithRkeeper/internal/wooapi/models"
	optionsWoo "WooWithRkeeper/internal/wooapi/options"
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type WOOAPI interface {
	ProductGet(ID int) (*models.Product, error)
	ProductList(opts ...optionsWoo.Option) ([]*models.Product, error)
	ProductListAll() ([]*models.Product, error)
	ProductAdd(p *models.Product) (*models.Product, error)
	ProductUpdate(p *models.Product) (*models.Product, error)
	ProductDel(ID int, opts ...optionsWoo.Option) error

	ProductCategoryGet(ID int) (*models.ProductCategory, error)
	ProductCategoryList(opts ...optionsWoo.Option) ([]*models.ProductCategory, error)
	ProductCategoryListAll() ([]*models.ProductCategory, error)
	ProductCategoryAdd(c *models.ProductCategory) (*models.ProductCategory, error)
	ProductCategoryUpdate(pc *models.ProductCategory) (*models.ProductCategory, error)
	ProductCategoryDelete(ID int, opts ...optionsWoo.Option) error
}

type wooapi struct {
	url    string
	key    string
	secret string
	api    client.Client
}

func (w *wooapi) ProductGet(ID int) (*models.Product, error) {
	logger := logging.GetLogger()
	logger.Println("ProductGet:>Start")
	defer logger.Println("ProductGet:>End")

	endpoint := fmt.Sprintf("products/%d", ID)
	logger.Debugf("Endpoint: %s", endpoint)

	if r, err := w.api.Get(endpoint, nil); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			//logger.Info("X-WP-TotalPages: ", r.Header.Get("X-WP-TotalPages"))
			var product models.Product
			err := json.Unmarshal(bodyBytes, &product)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return &product, nil
		}
	}
}

func (w *wooapi) ProductList(opts ...optionsWoo.Option) ([]*models.Product, error) {
	logger := logging.GetLogger()
	logger.Println("ProductList:>Start")
	defer logger.Println("ProductList:>End")

	endpoint := "products"
	logger.Debugf("Endpoint: %s", endpoint)

	params := url.Values{}
	//add fields is BEAUTIFUL!!
	Option := new(optionsWoo.OptionStruct)
	for _, field := range opts {
		field(Option)
		params.Add(Option.Key, Option.Value)
	}

	if r, err := w.api.Get(endpoint, params); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			logger.Debugf("X-WP-TotalPages: %s", r.Header.Get("X-WP-TotalPages"))
			var products []*models.Product
			err := json.Unmarshal(bodyBytes, &products)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return products, nil
		}
	}
}

func (w *wooapi) ProductListAll() ([]*models.Product, error) {

	logger := logging.GetLogger()
	logger.Println("ProductListAll:>Start")
	defer logger.Println("ProductListAll:>End")

	// Get All Products
	var products []*models.Product
	var i = 1
	perPage := 100
	for {
		productsTemp, err := w.ProductList(optionsWoo.PerPage(perPage), optionsWoo.Page(i))
		if err != nil {
			logger.Errorf("ошибка при получении ProductList, PerPage:%d, Page:%d, error:%v", perPage, i, err)
			return nil, errors.Wrapf(err, "ошибка при получении ProductList, PerPage:%d, Page:%d", perPage, i)
		}

		if len(productsTemp) == 0 {
			break
		}

		products = append(products, productsTemp...)
		logger.Debugf("Page load:%d", i)
		i++
	}

	return products, nil
}

func (w *wooapi) ProductAdd(p *models.Product) (*models.Product, error) {
	logger := logging.GetLogger()
	logger.Println("ProductAdd:>Start")
	defer logger.Println("ProductAdd:>End")

	endpoint := fmt.Sprintf("products")
	logger.Debugf("Endpoint: %s", endpoint)

	if p.Name == "" {
		return nil, errors.New("не указано имя продукта")
	}

	if r, err := w.api.Post(endpoint, nil, p); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusCreated { //TODO надо подумать что значит статус 200 и будет ли он возникать
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var product models.Product
			err := json.Unmarshal(bodyBytes, &product)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return &product, nil
		}
	}
}

func (w *wooapi) ProductUpdate(p *models.Product) (*models.Product, error) {
	logger := logging.GetLogger()
	logger.Println("ProductUpdate:>Start")
	defer logger.Println("ProductUpdate:>End")

	if p.ID == 0 {
		return nil, errors.New("не указана ID продукта")
	}

	endpoint := fmt.Sprintf("products/%d", p.ID)
	logger.Debugf("Endpoint: %s", endpoint)

	if r, err := w.api.Put(endpoint, p); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK { //TODO есть ли еще статусы кроме 200
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var product models.Product
			err := json.Unmarshal(bodyBytes, &product)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return &product, nil
		}
	}
}

func (w *wooapi) ProductDel(ID int, opts ...optionsWoo.Option) error {
	logger := logging.GetLogger()
	logger.Println("ProductDel:>Start")
	defer logger.Println("ProductDel:>End")

	endpoint := fmt.Sprintf("products/%d", ID)
	logger.Debugf("Endpoint: %s", endpoint)

	params := url.Values{}
	//add fields is BEAUTIFUL!!
	Option := new(optionsWoo.OptionStruct)
	for _, field := range opts {
		field(Option)
		params.Add(Option.Key, Option.Value)
	}

	if r, err := w.api.Delete(endpoint, params); err != nil {
		return errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			if ErrorWoo.Code == "woocommerce_rest_already_trashed" {
				return nil
			}
			return &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			//logger.Info("X-WP-TotalPages: ", r.Header.Get("X-WP-TotalPages"))
			var product models.Product
			err := json.Unmarshal(bodyBytes, &product)
			if err != nil {
				return errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil
		}
	}
}

func (w *wooapi) ProductCategoryGet(ID int) (*models.ProductCategory, error) {
	logger := logging.GetLogger()
	logger.Println("ProductCategoryGet:>Start")
	defer logger.Println("ProductCategoryGet:>End")

	endpoint := fmt.Sprintf("products/categories/%d", ID)
	logger.Debugf("Endpoint: %s", endpoint)

	if r, err := w.api.Get(endpoint, nil); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			//logger.Info("X-WP-TotalPages: ", r.Header.Get("X-WP-TotalPages"))
			var productCategory models.ProductCategory
			err := json.Unmarshal(bodyBytes, &productCategory)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return &productCategory, nil
		}
	}
}

func (w *wooapi) ProductCategoryList(opts ...optionsWoo.Option) ([]*models.ProductCategory, error) {
	logger := logging.GetLogger()
	logger.Println("ProductCategoryList:>Start")
	defer logger.Println("ProductCategoryList:>End")

	endpoint := "products/categories"
	logger.Debugf("Endpoint: %s", endpoint)

	params := url.Values{}
	//add fields is BEAUTIFUL!!
	Option := new(optionsWoo.OptionStruct)
	for _, field := range opts {
		field(Option)
		params.Add(Option.Key, Option.Value)
	}

	if r, err := w.api.Get(endpoint, params); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			logger.Debugf("X-WP-TotalPages: %s", r.Header.Get("X-WP-TotalPages"))
			var productsCategory []*models.ProductCategory
			err := json.Unmarshal(bodyBytes, &productsCategory)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return productsCategory, nil
		}
	}
}

func (w *wooapi) ProductCategoryListAll() ([]*models.ProductCategory, error) {

	logger := logging.GetLogger()
	logger.Println("ProductCategoryListAll:>Start")
	defer logger.Println("ProductCategoryListAll:>End")

	// Get All Products
	var productsCategory []*models.ProductCategory
	var i = 1
	perPage := 100
	for {
		productsCategoryTemp, err := w.ProductCategoryList(optionsWoo.PerPage(perPage), optionsWoo.Page(i))
		if err != nil {
			logger.Errorf("ошибка при получении ProductCategoryList, PerPage:%d, Page:%d, error:%v", perPage, i, err)
			return nil, errors.Wrapf(err, "ошибка при получении ProductCategoryList, PerPage:%d, Page:%d", perPage, i)
		}

		if len(productsCategoryTemp) == 0 {
			break
		}

		productsCategory = append(productsCategory, productsCategoryTemp...)
		logger.Debugf("Page load:%d", i)
		i++
	}

	return productsCategory, nil
}

func (w *wooapi) ProductCategoryAdd(c *models.ProductCategory) (*models.ProductCategory, error) {
	logger := logging.GetLogger()
	logger.Println("ProductCategoryAdd:>Start")
	defer logger.Println("ProductCategoryAdd:>End")

	endpoint := fmt.Sprintf("products/categories")
	logger.Debugf("Endpoint: %s", endpoint)

	if c.Name == "" {
		return nil, errors.New("не указано имя категории")
	}

	if r, err := w.api.Post(endpoint, nil, c); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusCreated { //TODO надо подумать что значит статус 200 и будет ли он возникать
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var productCategory models.ProductCategory
			err := json.Unmarshal(bodyBytes, &productCategory)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return &productCategory, nil
		}
	}
}

func (w *wooapi) ProductCategoryUpdate(pc *models.ProductCategory) (*models.ProductCategory, error) {
	logger := logging.GetLogger()
	logger.Println("ProductCategoryUpdate:>Start")
	defer logger.Println("ProductCategoryUpdate:>End")

	if pc.ID == 0 {
		return nil, errors.New("не указана ID папки меню")
	}

	endpoint := fmt.Sprintf("products/categories/%d", pc.ID)
	logger.Debugf("Endpoint: %s", endpoint)

	if r, err := w.api.Put(endpoint, pc); err != nil {
		return nil, errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK { //TODO есть ли еще статусы кроме 200
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {
			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil, &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var productCategory models.ProductCategory
			err := json.Unmarshal(bodyBytes, &productCategory)
			if err != nil {
				return nil, errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return &productCategory, nil
		}
	}
}

func (w *wooapi) ProductCategoryDelete(ID int, opts ...optionsWoo.Option) error {
	logger := logging.GetLogger()
	logger.Println("ProductCategoryDelete:>Start")
	defer logger.Println("ProductCategoryDelete:>End")

	endpoint := fmt.Sprintf("products/categories/%d", ID)
	logger.Debugf("Endpoint: %s", endpoint)

	params := url.Values{}
	//add fields is BEAUTIFUL!!
	Option := new(optionsWoo.OptionStruct)
	for _, field := range opts {
		field(Option)
		params.Add(Option.Key, Option.Value)
	}

	if r, err := w.api.Delete(endpoint, params); err != nil {
		return errors.Wrapf(err, "ошибка при отправке запроса в Woo Api, endpoint:%s", endpoint)
	} else if r.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			var ErrorWoo models.ErrorWoo
			err := json.Unmarshal(bodyBytes, &ErrorWoo)
			if err != nil {
				return errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			if ErrorWoo.Code == "woocommerce_rest_already_trashed" {
				return nil
			}
			return &ErrorWoo
		}
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Errorf("failed Body.Close()")
			}
		}(r.Body)
		if bodyBytes, err := ioutil.ReadAll(r.Body); err != nil {
			return errors.Wrapf(err, "ошибка при ioutil.ReadAll(r.Body): error: %v", err)
		} else {

			logger.Debugf(string(bodyBytes))
			//logger.Info("X-WP-TotalPages: ", r.Header.Get("X-WP-TotalPages"))
			var productCategory models.ProductCategory
			err := json.Unmarshal(bodyBytes, &productCategory)
			if err != nil {
				return errors.Wrapf(err, "ошибка при json.Unmarshal(): error: %v", err)
			}
			return nil
		}
	}
}

func NewAPI(url, key, secret string) WOOAPI {

	factory := client.Factory{}

	api := factory.NewClient(options.Basic{
		URL:    url,
		Key:    key,
		Secret: secret,
		Options: options.Advanced{
			WPAPI:       true,
			WPAPIPrefix: "/wp-json/",
			Version:     "wc/v3",
		},
	})

	return &wooapi{
		url:    url,
		key:    key,
		secret: secret,
		api:    api,
	}
}
