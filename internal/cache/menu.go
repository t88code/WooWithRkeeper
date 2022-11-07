package cache

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/wooapi"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type Menu interface {
	//Refresh All menu
	RefreshMenu() error

	//RK7
	RefreshCateglist() error
	RefreshMenuitems() error
	GetMenuitems() ([]*modelsRK7API.MenuitemItem, error)
	GetMenuitemsRK7ByIdent() (map[int]*modelsRK7API.MenuitemItem, error)
	GetCateglistRK7() ([]*modelsRK7API.Categlist, error)
	GetCateglistsRK7ByIdent() (map[int]*modelsRK7API.Categlist, error)

	//WOO
	RefreshProducts() error
	RefreshProductCategories() error
	GetProductsWooByID() (map[int]*modelsWOOAPI.Product, error)

	GetProductCategoriesWooByID() (map[int]*modelsWOOAPI.ProductCategory, error)
	GetProductCategoriesWooBySlug() (map[string]*modelsWOOAPI.ProductCategory, error)

	DeleteProductCategoryFromCache(ID int) error
	AddProductCategoryToCache(category *modelsWOOAPI.ProductCategory) error

	DeleteProductFromCache(ID int) error
	AddProductToCache(product *modelsWOOAPI.Product) error
}

var cacheMenuGlobal menu

type menu struct {
	//RK7 - Menuitems
	MenuitemsRK7        []*modelsRK7API.MenuitemItem
	MenuitemsRK7ByIdent map[int]*modelsRK7API.MenuitemItem
	MenuitemsRK7ByWooID map[int]*modelsRK7API.MenuitemItem

	//RK7 - Categlists
	CateglistsRK7        []*modelsRK7API.Categlist
	CateglistsRK7ByIdent map[int]*modelsRK7API.Categlist

	//WOO - Product
	ProductsWooByID map[int]*modelsWOOAPI.Product

	//WOO - ProductCategories
	ProductCategoriesWooByID   map[int]*modelsWOOAPI.ProductCategory
	ProductCategoriesWooBySlug map[string]*modelsWOOAPI.ProductCategory
}

func (m *menu) AddProductToCache(product *modelsWOOAPI.Product) error {
	logger := logging.GetLogger()
	logger.Info("Start AddProductToCache")
	defer logger.Info("End AddProductToCache")

	logger.Debugf("Блюдо: Name=%s, Id=%d, Parent=%d, Slug=%s, RegularPrice=%s",
		product.Name, product.ID, product.Categories[0].Id, product.Slug, product.RegularPrice)

	if product != nil {
		if m.ProductsWooByID == nil {
			m.ProductsWooByID = make(map[int]*modelsWOOAPI.Product)
		}

		lenProductsWooByID := len(m.ProductsWooByID)

		logger.Debug("До добавления в кеш:")
		logger.Debugf("len(ProductsWooByID)=%d", lenProductsWooByID)

		if product.ID == 0 {
			return errors.New(fmt.Sprintf("ProductsWooByID не был изменен при добавлении элемента Name=%s, ID=%d, Parent=%d, Slug=%s, RegularPrice=%s; ID=0",
				product.Name,
				product.ID,
				product.Categories[0].Id,
				product.Slug,
				product.RegularPrice))
		}

		m.ProductsWooByID[product.ID] = product

		logger.Debug("После добавления в кеш:")
		logger.Debugf("len(ProductsWooByID)=%d", len(m.ProductsWooByID))

		if lenProductsWooByID == len(m.ProductsWooByID) {
			return errors.New(fmt.Sprintf("ProductsWooByID не был изменен при добавлении элемента Name=%s, ID=%d, Categories[0].Id=%d, Slug=%s, RegularPrice=%s",
				product.Name,
				product.ID,
				product.Categories[0].Id,
				product.Slug,
				product.RegularPrice))
		} else {
			return nil
		}
	} else {
		return errors.New("Products is nil")
	}
}

func (m *menu) DeleteProductFromCache(WOOID int) error {
	logger := logging.GetLogger()
	logger.Info("Start DeleteProductFromCache")
	defer logger.Info("End DeleteProductFromCache")

	lenProductsWooByID := len(m.ProductsWooByID)

	logger.Debug("До удаления из кеша:")
	logger.Debugf("len(ProductsWooByID)=%d", lenProductsWooByID)

	if m.ProductsWooByID != nil {
		if _, found := m.ProductsWooByID[WOOID]; found {
			delete(m.ProductsWooByID, WOOID)
		} else {
			logger.Warnf("Не найдено блюдо в кеше WOO по ID=%d", WOOID)
		}
	}

	logger.Debug("После удаления из кеша:")
	logger.Debugf("len(ProductsWooByID)=%d", len(m.ProductsWooByID))

	if lenProductsWooByID == len(m.ProductsWooByID) {
		return errors.New(fmt.Sprintf("ProductsWooByID не был изменен при удалении элемента ID=%d", WOOID))
	}

	return nil
}

func (m *menu) AddProductCategoryToCache(category *modelsWOOAPI.ProductCategory) error {
	logger := logging.GetLogger()
	logger.Info("Start AddProductCategoryToCache")
	defer logger.Info("End AddProductCategoryToCache")

	logger.Debugf("Папка: Name=%s, Id=%d, Parent=%d, Slug=%s",
		category.Name, category.ID, category.Parent, category.Slug)

	if category != nil {
		if m.ProductCategoriesWooByID == nil {
			m.ProductCategoriesWooByID = make(map[int]*modelsWOOAPI.ProductCategory)
		}

		if m.ProductCategoriesWooBySlug == nil {
			m.ProductCategoriesWooBySlug = make(map[string]*modelsWOOAPI.ProductCategory)
		}

		lenProductCategoriesWooByID := len(m.ProductCategoriesWooByID)
		lenProductCategoriesWooBySlug := len(m.ProductCategoriesWooBySlug)

		logger.Debug("До добавления в кеш:")
		logger.Debugf("len(ProductCategoriesWooByID)=%d", lenProductCategoriesWooByID)
		logger.Debugf("len(ProductCategoriesWooBySlug)=%d", lenProductCategoriesWooBySlug)

		if category.ID == 0 {
			return errors.New(fmt.Sprintf("ProductCategoriesWooByID не был изменен при добавлении элемента Name=%s, ID=%d, Parent=%d, Slug=%s; ID=0",
				category.Name,
				category.ID,
				category.Parent,
				category.Slug))
		}
		m.ProductCategoriesWooByID[category.ID] = category

		if category.Slug == "" {
			return errors.New(fmt.Sprintf("ProductCategoriesWooBySlug не был изменен при добавлении элемента Name=%s, ID=%d, Parent=%d, Slug=%s; Slug=''",
				category.Name,
				category.ID,
				category.Parent,
				category.Slug))
		}
		m.ProductCategoriesWooBySlug[category.Slug] = category

		logger.Debug("После добавления в кеш:")
		logger.Debugf("len(ProductCategoriesWooByID)=%d", len(m.ProductCategoriesWooByID))
		logger.Debugf("len(ProductCategoriesWooBySlug)=%d", len(m.ProductCategoriesWooBySlug))

		if lenProductCategoriesWooByID == len(m.ProductCategoriesWooByID) {
			return errors.New(fmt.Sprintf("ProductCategoriesWooByID не был изменен при добавлении элемента Name=%s, ID=%d, Parent=%d, Slug=%s",
				category.Name,
				category.ID,
				category.Parent,
				category.Slug))
		}

		if lenProductCategoriesWooBySlug == len(m.ProductCategoriesWooBySlug) {
			return errors.New(fmt.Sprintf("ProductCategoriesWooBySlug не был изменен при добавлении элемента Name=%s, ID=%d, Parent=%d, Slug=%s",
				category.Name,
				category.ID,
				category.Parent,
				category.Slug))
		}

		return nil
	} else {
		return errors.New("ProductCategory is nil")
	}
}

func (m *menu) DeleteProductCategoryFromCache(WOOID int) error {
	logger := logging.GetLogger()
	logger.Info("Start DeleteProductCategoryFromCache")
	defer logger.Info("End DeleteProductCategoryFromCache")

	lenProductCategoriesWooByID := len(m.ProductCategoriesWooByID)
	lenProductCategoriesWooBySlug := len(m.ProductCategoriesWooBySlug)

	logger.Debug("До удаления из кеша:")
	logger.Debugf("len(ProductCategoriesWooByID)=%d", lenProductCategoriesWooByID)
	logger.Debugf("len(ProductCategoriesWooBySlug)=%d", lenProductCategoriesWooBySlug)

	if m.ProductCategoriesWooByID != nil {
		if product, found := m.ProductCategoriesWooByID[WOOID]; found {
			delete(m.ProductCategoriesWooByID, WOOID)
			if _, found := m.ProductCategoriesWooBySlug[product.Slug]; found {
				delete(m.ProductCategoriesWooBySlug, product.Slug)
			} else {
				logger.Warnf("Не найдено блюдо в кеше WOO по Slug=%s", product.Slug)
			}
		} else {
			logger.Warnf("Не найдено блюдо в кеше WOO по ID=%d", WOOID)
		}
	}

	logger.Debug("После удаления из кеша:")
	logger.Debugf("len(ProductCategoriesWooByID)=%d", len(m.ProductCategoriesWooByID))
	logger.Debugf("len(ProductCategoriesWooBySlug)=%d", len(m.ProductCategoriesWooBySlug))

	if lenProductCategoriesWooByID == len(m.ProductCategoriesWooByID) {
		return errors.New(fmt.Sprintf("ProductCategoriesWooByID не был изменен при удалении элемента ID=%d", WOOID))
	}

	if lenProductCategoriesWooBySlug == len(m.ProductCategoriesWooBySlug) {
		return errors.New(fmt.Sprintf("ProductCategoriesWooBySlug не был изменен при удалении элемента ID=%d", WOOID))
	}

	return nil
}

func (m *menu) GetMenuitems() ([]*modelsRK7API.MenuitemItem, error) {
	// TODO if version>0 || timeoute>0 {RefreshMenuitems()}
	return m.MenuitemsRK7, nil
}

func (m *menu) GetMenuitemsRK7ByIdent() (map[int]*modelsRK7API.MenuitemItem, error) {
	return m.MenuitemsRK7ByIdent, nil
}

func (m *menu) GetCateglistRK7() ([]*modelsRK7API.Categlist, error) {
	return m.CateglistsRK7, nil
}

func (m *menu) GetCateglistsRK7ByIdent() (map[int]*modelsRK7API.Categlist, error) {
	return m.CateglistsRK7ByIdent, nil
}

func (m *menu) GetProductsWooByID() (map[int]*modelsWOOAPI.Product, error) {
	return m.ProductsWooByID, nil
}

func (m *menu) GetProductCategoriesWooByID() (map[int]*modelsWOOAPI.ProductCategory, error) {
	return m.ProductCategoriesWooByID, nil
}

func (m *menu) GetProductCategoriesWooBySlug() (map[string]*modelsWOOAPI.ProductCategory, error) {
	return m.ProductCategoriesWooBySlug, nil
}

func (m *menu) RefreshCateglist() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshCateglist")
	defer logger.Info("End RefreshCateglist")
	timeStart := time.Now()
	cfg := config.GetConfig()
	RK7API, err := rk7api.NewAPI(cfg.RK7.URL, cfg.RK7.User, cfg.RK7.Pass)
	if err != nil {
		return errors.Wrap(err, "failed rk7api.NewAPI")
	}

	//получить список всех Categlist из RK
	logger.Info("Получить список Categlist из RK7")
	Rk7QueryResultGetRefDataCateglist, err := RK7API.GetRefData("Categlist",
		modelsRK7API.OnlyActive("0"), //неактивные будем грохать в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Ident,ItemIdent,GUIDString,Code,Name,MainParentIdent,Status,Parent,genIDBX24,genSectionIDBX24,genWOO_ID,genWOO_PARENT_ID,genWOO_LONGNAME)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	categlists := (Rk7QueryResultGetRefDataCateglist).(*modelsRK7API.RK7QueryResultGetRefDataCateglist) //todo приведение в структуру работает прекрасно
	m.CateglistsRK7 = categlists.RK7Reference.Items.Item
	logger.Infof("Длина списка CateglistInRK7 = %d\n", len(m.CateglistsRK7))
	m.CateglistsRK7ByIdent = make(map[int]*modelsRK7API.Categlist)

	for i, item := range m.CateglistsRK7 {
		m.CateglistsRK7ByIdent[item.ItemIdent] = m.CateglistsRK7[i]
	}
	logger.Infof("Создан CateglistMap, длина: %d", len(m.CateglistsRK7ByIdent))
	logger.Infof("RefreshCateglist. Время обновления: %s", time.Now().Sub(timeStart))
	return nil
}

func (m *menu) RefreshMenuitems() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshMenuitems")
	defer logger.Info("End RefreshMenuitems")
	timeStart := time.Now()
	cfg := config.GetConfig()
	RK7API, err := rk7api.NewAPI(cfg.RK7.URL, cfg.RK7.User, cfg.RK7.Pass)
	if err != nil {
		return errors.Wrap(err, "failed rk7api.NewAPI")
	}

	//получить актуальное меню RK7
	logger.Info("Получить список всех блюд из RK7")
	Rk7QueryResultGetRefDataMenuitems, err := RK7API.GetRefData("Menuitems",
		modelsRK7API.OnlyActive("0"), //неактивные будем менять статус на N ?или может удалять? в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Code,Name,Ident,ItemIdent,GUIDString,MainParentIdent,ExtCode,PRICETYPES^3,CategPath,Status,genIDBX24,genSectionIDBX24,genWOO_ID,genWOO_PARENT_ID,genWOO_LONGNAME)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}
	menuitems := (Rk7QueryResultGetRefDataMenuitems).(*modelsRK7API.RK7QueryResultGetRefDataMenuitems)
	m.MenuitemsRK7 = menuitems.RK7Reference.Items.Item
	logger.Infof("Длина списка MenuitemItemInRK7 = %d\n", len(m.MenuitemsRK7))

	m.MenuitemsRK7ByIdent = make(map[int]*modelsRK7API.MenuitemItem)
	for i, item := range m.MenuitemsRK7 {
		m.MenuitemsRK7ByIdent[item.ItemIdent] = m.MenuitemsRK7[i]
	}
	logger.Infof("Длина списка MenuRK7MapByIdent = %d\n", len(m.MenuitemsRK7ByIdent))
	logger.Infof("RefreshMenuitems. Время обновления: %s", time.Now().Sub(timeStart))
	return nil
}

func (m *menu) RefreshProducts() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshProducts")
	defer logger.Info("End RefreshProducts")

	WOOAPI := wooapi.GetAPI()

	logger.Info("Получить список всех товаров из WOO")
	products, err := WOOAPI.ProductListAll()
	if err != nil {
		return errors.Wrap(err, "failed in WOOAPI.ProductListAll()")
	}

	m.ProductsWooByID = make(map[int]*modelsWOOAPI.Product)

	for i, product := range products {
		m.ProductsWooByID[product.ID] = products[i]
	}

	logger.Infof("Длина списка ProductsWooByID = %d\n", len(m.ProductsWooByID))

	return nil
}

func (m *menu) RefreshProductCategories() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshProductCategories")
	defer logger.Info("End RefreshProductCategories")

	WOOAPI := wooapi.GetAPI()

	logger.Info("Получить список всех ProductCategories из WOO")
	productCategories, err := WOOAPI.ProductCategoryListAll()
	if err != nil {
		return errors.Wrap(err, "failed in WOOAPI.ProductCategoryListAll()")
	}

	m.ProductCategoriesWooByID = make(map[int]*modelsWOOAPI.ProductCategory)
	m.ProductCategoriesWooBySlug = make(map[string]*modelsWOOAPI.ProductCategory)

	for i, productCategory := range productCategories {
		logger.Debugf("Product: Name=%s, Slug=%s, ID=%d", productCategory.Name, productCategory.Slug, productCategory.ID)
		m.ProductCategoriesWooByID[productCategory.ID] = productCategories[i]
		m.ProductCategoriesWooBySlug[productCategory.Slug] = productCategories[i]
	}

	logger.Infof("Длина списка ProductCategoriesWooByID = %d\n", len(m.ProductCategoriesWooByID))
	logger.Infof("Длина списка ProductCategoriesWooBySlug = %d\n", len(m.ProductCategoriesWooBySlug))

	return nil
}

func (m *menu) RefreshMenu() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshMenu")
	defer logger.Info("End RefreshMenu")

	err := m.RefreshMenuitems()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshMenuitems()")
	}

	err = m.RefreshCateglist()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshCateglist()")
	}

	err = m.RefreshProducts()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshProducts()")
	}

	err = m.RefreshProductCategories()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshProductCategories()")
	}

	return nil
}

func NewCacheMenu() (Menu, error) {
	logger := logging.GetLogger()
	logger.Info("Start NewCacheMenu")
	defer logger.Info("End NewCacheMenu")

	return &cacheMenuGlobal, nil
}

func GetMenu() (Menu, error) {
	logger := logging.GetLogger()
	logger.Info("Start GetMenu")
	defer logger.Info("End GetMenu")

	return &cacheMenuGlobal, nil
}
