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
	RefreshDishRests() error

	GetMenuitems() ([]*modelsRK7API.MenuitemItem, error)
	GetMenuitemsRK7ByIdent() (map[int]*modelsRK7API.MenuitemItem, error)
	GetMenuitemsRK7ByWooID() (map[int]*modelsRK7API.MenuitemItem, error)

	GetCateglistRK7() ([]*modelsRK7API.Categlist, error)
	GetCateglistsRK7ByIdent() (map[int]*modelsRK7API.Categlist, error)

	GetDishRests() ([]*modelsRK7API.DishRest, error)
	GetDishRestsByIdent() (map[int]*modelsRK7API.DishRest, error)

	//WOO
	RefreshProducts() error
	RefreshProductCategories() error

	GetProductsWooByID() (map[int]*modelsWOOAPI.Product, error)

	GetProductCategoriesWooByID() (map[int]*modelsWOOAPI.ProductCategory, error)

	DeleteProductCategoryFromCache(ID int) error
	AddProductCategoryToCache(category *modelsWOOAPI.ProductCategory) error

	DeleteProductFromCache(ID int) error
	AddProductToCache(product *modelsWOOAPI.Product) error
}

var cacheMenuGlobal menu

type menu struct {

	//parametr
	cacheLifetime time.Duration

	//RK7 - Menuitems
	MenuitemsRK7        []*modelsRK7API.MenuitemItem
	MenuitemsRK7ByIdent map[int]*modelsRK7API.MenuitemItem
	MenuitemsRK7ByWooID map[int]*modelsRK7API.MenuitemItem

	//RK7 - Categlists
	CateglistsRK7        []*modelsRK7API.Categlist
	CateglistsRK7ByIdent map[int]*modelsRK7API.Categlist

	//RK7 - StopList
	DishRests        []*modelsRK7API.DishRest
	DishRestsByIdent map[int]*modelsRK7API.DishRest

	//WOO - Product
	ProductsWooByID map[int]*modelsWOOAPI.Product

	//WOO - ProductCategories
	ProductCategoriesWooByID map[int]*modelsWOOAPI.ProductCategory
}

func (m *menu) GetMenuitemsRK7ByWooID() (map[int]*modelsRK7API.MenuitemItem, error) {
	if len(m.MenuitemsRK7ByWooID) == 0 {
		err := m.RefreshMenuitems()
		if err != nil {
			return nil, err
		}
	}
	return m.MenuitemsRK7ByWooID, nil
}

func (m *menu) AddProductToCache(product *modelsWOOAPI.Product) error {
	logger := logging.GetLogger()
	logger.Debug("Start AddProductToCache")
	defer logger.Debug("End AddProductToCache")

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
	logger.Debug("Start DeleteProductFromCache")
	defer logger.Debug("End DeleteProductFromCache")

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
	logger.Debug("Start AddProductCategoryToCache")
	defer logger.Debug("End AddProductCategoryToCache")

	logger.Debugf("Папка: Name=%s, Id=%d, Parent=%d, Slug=%s",
		category.Name, category.ID, category.Parent, category.Slug)

	if category != nil {
		if m.ProductCategoriesWooByID == nil {
			m.ProductCategoriesWooByID = make(map[int]*modelsWOOAPI.ProductCategory)
		}

		lenProductCategoriesWooByID := len(m.ProductCategoriesWooByID)

		logger.Debug("До добавления в кеш:")
		logger.Debugf("len(ProductCategoriesWooByID)=%d", lenProductCategoriesWooByID)

		if category.ID == 0 {
			return errors.New(fmt.Sprintf("ProductCategoriesWooByID не был изменен при добавлении элемента Name=%s, ID=%d, Parent=%d, Slug=%s; ID=0",
				category.Name,
				category.ID,
				category.Parent,
				category.Slug))
		}
		m.ProductCategoriesWooByID[category.ID] = category

		logger.Debug("После добавления в кеш:")
		logger.Debugf("len(ProductCategoriesWooByID)=%d", len(m.ProductCategoriesWooByID))

		if lenProductCategoriesWooByID == len(m.ProductCategoriesWooByID) {
			return errors.New(fmt.Sprintf("ProductCategoriesWooByID не был изменен при добавлении элемента Name=%s, ID=%d, Parent=%d, Slug=%s",
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
	logger.Debug("Start DeleteProductCategoryFromCache")
	defer logger.Debug("End DeleteProductCategoryFromCache")

	lenProductCategoriesWooByID := len(m.ProductCategoriesWooByID)

	logger.Debug("До удаления из кеша:")
	logger.Debugf("len(ProductCategoriesWooByID)=%d", lenProductCategoriesWooByID)

	if m.ProductCategoriesWooByID != nil {
		if _, found := m.ProductCategoriesWooByID[WOOID]; found {
			delete(m.ProductCategoriesWooByID, WOOID)
		} else {
			logger.Warnf("Не найдено блюдо в кеше WOO по ID=%d", WOOID)
		}
	}

	logger.Debug("После удаления из кеша:")
	logger.Debugf("len(ProductCategoriesWooByID)=%d", len(m.ProductCategoriesWooByID))

	if lenProductCategoriesWooByID == len(m.ProductCategoriesWooByID) {
		return errors.New(fmt.Sprintf("ProductCategoriesWooByID не был изменен при удалении элемента ID=%d", WOOID))
	}

	return nil
}

func (m *menu) GetMenuitems() ([]*modelsRK7API.MenuitemItem, error) {
	// todo сделать проверку по времени
	if len(m.MenuitemsRK7) == 0 {
		err := m.RefreshMenuitems()
		if err != nil {
			return nil, err
		}
	}
	return m.MenuitemsRK7, nil
}

func (m *menu) GetMenuitemsRK7ByIdent() (map[int]*modelsRK7API.MenuitemItem, error) {
	if len(m.MenuitemsRK7ByIdent) == 0 {
		err := m.RefreshMenuitems()
		if err != nil {
			return nil, err
		}
	}
	return m.MenuitemsRK7ByIdent, nil
}

func (m *menu) GetDishRests() ([]*modelsRK7API.DishRest, error) {
	if len(m.DishRests) == 0 {
		err := m.RefreshDishRests()
		if err != nil {
			return nil, err
		}
	}
	return m.DishRests, nil
}

func (m *menu) GetDishRestsByIdent() (map[int]*modelsRK7API.DishRest, error) {
	if len(m.DishRestsByIdent) == 0 {
		err := m.RefreshDishRests()
		if err != nil {
			return nil, err
		}
	}
	return m.DishRestsByIdent, nil
}

func (m *menu) GetCateglistRK7() ([]*modelsRK7API.Categlist, error) {
	if len(m.CateglistsRK7) == 0 {
		err := m.RefreshCateglist()
		if err != nil {
			return nil, err
		}
	}
	return m.CateglistsRK7, nil
}

func (m *menu) GetCateglistsRK7ByIdent() (map[int]*modelsRK7API.Categlist, error) {
	if len(m.CateglistsRK7ByIdent) == 0 {
		err := m.RefreshCateglist()
		if err != nil {
			return nil, err
		}
	}
	return m.CateglistsRK7ByIdent, nil
}

func (m *menu) GetProductsWooByID() (map[int]*modelsWOOAPI.Product, error) {
	if len(m.ProductsWooByID) == 0 {
		err := m.RefreshProducts()
		if err != nil {
			return nil, err
		}
	}
	return m.ProductsWooByID, nil
}

func (m *menu) GetProductCategoriesWooByID() (map[int]*modelsWOOAPI.ProductCategory, error) {
	if len(m.ProductCategoriesWooByID) == 0 {
		err := m.RefreshProductCategories()
		if err != nil {
			return nil, err
		}
	}
	return m.ProductCategoriesWooByID, nil
}

func (m *menu) RefreshCateglist() error {

	logger := logging.GetLogger()
	logger.Debug("Start RefreshCateglist")
	defer logger.Debug("End RefreshCateglist")
	timeStart := time.Now()
	RK7API := rk7api.GetAPI("REF")
	//получить список всех Categlist из RK
	logger.Debug("Получить список Categlist из RK7")
	Rk7QueryResultGetRefDataCateglist, err := RK7API.GetRefData("Categlist", nil,
		modelsRK7API.OnlyActive("0"), //неактивные будем грохать в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Ident,ItemIdent,GUIDString,Code,Name,MainParentIdent,Status,Parent,genIDBX24,genSectionIDBX24,genWOO_ID,genWOO_PARENT_ID,genWOO_LONGNAME,genWOO_SYNC)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	categlists := (Rk7QueryResultGetRefDataCateglist).(*modelsRK7API.RK7QueryResultGetRefDataCateglist) //todo приведение в структуру работает прекрасно
	m.CateglistsRK7 = categlists.RK7Reference.Items.Item
	logger.Debugf("Длина списка CateglistInRK7 = %d\n", len(m.CateglistsRK7))
	m.CateglistsRK7ByIdent = make(map[int]*modelsRK7API.Categlist)

	for i, item := range m.CateglistsRK7 {
		m.CateglistsRK7ByIdent[item.ItemIdent] = m.CateglistsRK7[i]
	}
	logger.Debugf("Создан CateglistMap, длина: %d", len(m.CateglistsRK7ByIdent))
	logger.Debugf("RefreshCateglist. Время обновления: %s", time.Now().Sub(timeStart))
	return nil
}

func (m *menu) RefreshMenuitems() error {

	logger := logging.GetLogger()
	logger.Debug("Start RefreshMenuitems")
	defer logger.Debug("End RefreshMenuitems")
	timeStart := time.Now()
	cfg := config.GetConfig()
	RK7API := rk7api.GetAPI("REF")
	//получить актуальное меню RK7
	logger.Debug("Получить список всех блюд из RK7")
	var replaceAttribut []rk7api.ReplaceAttribut
	replaceAttribut = append(replaceAttribut, rk7api.ReplaceAttribut{
		Source:      fmt.Sprintf("CLASSIFICATORGROUPS-%d", cfg.RK7.CLASSIFICATORGROUPS),
		Destination: "CLASSIFICATORGROUPS"})
	replaceAttribut = append(replaceAttribut, rk7api.ReplaceAttribut{
		Source:      fmt.Sprintf("PRICETYPES-%d", cfg.RK7.PRICETYPE),
		Destination: "PRICETYPES"})

	propMask := fmt.Sprintf("items.(Code,Name,Ident,ItemIdent,GUIDString,MainParentIdent,ExtCode,PRICETYPES^%d,CategPath,Status,genIDBX24,genSectionIDBX24,genWOO_ID,genWOO_PARENT_ID,genWOO_LONGNAME,genWOO_IMAG*,CLASSIFICATORGROUPS^%d)",
		cfg.RK7.PRICETYPE,
		cfg.RK7.CLASSIFICATORGROUPS)

	Rk7QueryResultGetRefDataMenuitems, err := RK7API.GetRefData("Menuitems", replaceAttribut,
		modelsRK7API.OnlyActive("0"), //неактивные будем менять статус на N ?или может удалять? в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask(propMask))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	menuitems := (Rk7QueryResultGetRefDataMenuitems).(*modelsRK7API.RK7QueryResultGetRefDataMenuitems)
	m.MenuitemsRK7 = menuitems.RK7Reference.Items.Item
	logger.Debugf("Длина списка MenuitemItemInRK7 = %d\n", len(m.MenuitemsRK7))

	m.MenuitemsRK7ByIdent = make(map[int]*modelsRK7API.MenuitemItem)
	m.MenuitemsRK7ByWooID = make(map[int]*modelsRK7API.MenuitemItem)
	for i, item := range m.MenuitemsRK7 {
		m.MenuitemsRK7ByIdent[item.ItemIdent] = m.MenuitemsRK7[i]
		if item.WOO_ID != 0 {
			m.MenuitemsRK7ByWooID[item.WOO_ID] = m.MenuitemsRK7[i]
		}
	}
	logger.Debugf("Длина списка MenuRK7MapByIdent = %d\n", len(m.MenuitemsRK7ByIdent))
	logger.Debugf("Длина списка MenuitemsRK7ByWooID = %d\n", len(m.MenuitemsRK7ByWooID))
	logger.Debugf("RefreshMenuitems. Время обновления: %s", time.Now().Sub(timeStart))
	return nil
}

func (m *menu) RefreshProducts() error {

	logger := logging.GetLogger()
	logger.Debug("Start RefreshProducts")
	defer logger.Debug("End RefreshProducts")

	WOOAPI := wooapi.GetAPI()

	logger.Debug("Получить список всех товаров из WOO")
	products, err := WOOAPI.ProductListAll()
	if err != nil {
		return errors.Wrap(err, "failed in WOOAPI.ProductListAll()")
	}

	m.ProductsWooByID = make(map[int]*modelsWOOAPI.Product)

	for i, product := range products {
		m.ProductsWooByID[product.ID] = products[i]
	}

	logger.Debugf("Длина списка ProductsWooByID = %d\n", len(m.ProductsWooByID))

	return nil
}

func (m *menu) RefreshProductCategories() error {

	logger := logging.GetLogger()
	logger.Debug("Start RefreshProductCategories")
	defer logger.Debug("End RefreshProductCategories")

	WOOAPI := wooapi.GetAPI()

	logger.Debug("Получить список всех ProductCategories из WOO")
	productCategories, err := WOOAPI.ProductCategoryListAll()
	if err != nil {
		return errors.Wrap(err, "failed in WOOAPI.ProductCategoryListAll()")
	}

	m.ProductCategoriesWooByID = make(map[int]*modelsWOOAPI.ProductCategory)

	for i, productCategory := range productCategories {
		logger.Debugf("Product: Name=%s, Slug=%s, ID=%d", productCategory.Name, productCategory.Slug, productCategory.ID)
		m.ProductCategoriesWooByID[productCategory.ID] = productCategories[i]
	}

	logger.Debugf("Длина списка ProductCategoriesWooByID = %d\n", len(m.ProductCategoriesWooByID))

	return nil
}

func (m *menu) RefreshDishRests() error {

	logger := logging.GetLogger()
	logger.Debug("Start RefreshDishRests")
	defer logger.Debug("End RefreshDishRests")
	timeStart := time.Now()

	cfg := config.GetConfig()
	if cfg.MENUSYNC.SyncStopList != 0 {
		RK7API := rk7api.GetAPI("MID")

		logger.Debug("Получить список всех блюд из стоп-листа")

		Rk7QueryResultGetDishRests, err := RK7API.GetDishRests()
		if err != nil {
			return errors.Wrap(err, "Ошибка при выполнении rk7api.GetDishRests")
		}

		m.DishRests = Rk7QueryResultGetDishRests.DishRest
		logger.Debugf("Длина списка DishRests = %d\n", len(m.DishRests))

		m.DishRestsByIdent = make(map[int]*modelsRK7API.DishRest)
		for i, dish := range m.DishRests {
			m.DishRestsByIdent[dish.ID] = m.DishRests[i]
		}
		logger.Debugf("Длина списка DishRests = %d\n", len(m.DishRests))
		logger.Debugf("RefreshDishRests. Время обновления: %s", time.Now().Sub(timeStart))
	} else {
		m.DishRests = make([]*modelsRK7API.DishRest, 0)
		m.DishRestsByIdent = make(map[int]*modelsRK7API.DishRest)
	}
	return nil
}

func (m *menu) RefreshMenu() error {

	logger := logging.GetLogger()
	logger.Debug("Start RefreshMenu")
	defer logger.Debug("End RefreshMenu")

	//TODO оптимизация
	//можно запустить получение меню в отдельных потоках
	//1 поток - RK7REF
	//2 поток - RK7MID
	//3 поток - WOO

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

	err = m.RefreshDishRests()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshDishRests()")
	}

	return nil
}

func NewCacheMenu() (Menu, error) {
	logger := logging.GetLogger()
	logger.Debug("Start NewCacheMenu")
	defer logger.Debug("End NewCacheMenu")
	cacheMenuGlobal = menu{}
	return &cacheMenuGlobal, nil
}

func GetMenu() (Menu, error) {
	logger := logging.GetLogger()
	logger.Debug("Start GetMenu")
	defer logger.Debug("End GetMenu")

	return &cacheMenuGlobal, nil
}
