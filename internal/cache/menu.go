package cache

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/wooapi"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"github.com/pkg/errors"
)

type CacheMenu interface {
	RefreshMenu() error

	//RK7
	RefreshCateglist() error
	RefreshMenuitems() error
	GetMenuitems() ([]*modelsRK7API.MenuitemItem, error)
	GetMenuRK7MapByIdent() (map[int]*modelsRK7API.MenuitemItem, error)
	GetCateglistItemInRK7() ([]*modelsRK7API.Categlist, error)
	GetCateglistMapByIdent() (map[int]*modelsRK7API.Categlist, error)

	//WOO
	RefreshProducts() error
	RefreshProductCategories() error
	GetProductsMapByID() (map[int]*modelsWOOAPI.Product, error)
	GetProductCategoriesMapByID() (map[int]*modelsWOOAPI.ProductCategory, error)
}

var cacheMenuGlobal menu

type menu struct {
	//RK7
	MenuitemItemInRK7   []*modelsRK7API.MenuitemItem
	MenuRK7MapByIdent   map[int]*modelsRK7API.MenuitemItem
	CateglistItemInRK7  []*modelsRK7API.Categlist
	CateglistMapByIdent map[int]*modelsRK7API.Categlist

	//WOO
	ProductsMapByID            map[int]*modelsWOOAPI.Product
	ProductCategoriesMapByID   map[int]*modelsWOOAPI.ProductCategory
	ProductCategoriesMapByName map[string]*modelsWOOAPI.ProductCategory
}

func (m *menu) GetMenuitems() ([]*modelsRK7API.MenuitemItem, error) {
	// TODO if version>0 || timeoute>0 {RefreshMenuitems()}
	return m.MenuitemItemInRK7, nil
}

func (m *menu) GetMenuRK7MapByIdent() (map[int]*modelsRK7API.MenuitemItem, error) {
	return m.MenuRK7MapByIdent, nil
}

func (m *menu) GetCateglistItemInRK7() ([]*modelsRK7API.Categlist, error) {
	return m.CateglistItemInRK7, nil
}

func (m *menu) GetCateglistMapByIdent() (map[int]*modelsRK7API.Categlist, error) {
	return m.CateglistMapByIdent, nil
}

func (m *menu) GetProductsMapByID() (map[int]*modelsWOOAPI.Product, error) {
	return m.ProductsMapByID, nil
}

func (m *menu) GetProductCategoriesMapByID() (map[int]*modelsWOOAPI.ProductCategory, error) {
	return m.ProductCategoriesMapByID, nil
}

func (m *menu) GetProductCategoriesMapByName() (map[string]*modelsWOOAPI.ProductCategory, error) {
	return m.ProductCategoriesMapByName, nil
}

func (m *menu) RefreshCateglist() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshCateglist")
	defer logger.Info("End RefreshCateglist")
	cfg := config.GetConfig()
	RK7API := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7MID.User, cfg.RK7MID.Pass)

	//получить список всех Categlist из RK
	logger.Info("Получить список Categlist из RK7")
	Rk7QueryResultGetRefDataCateglist, err := RK7API.GetRefData("Categlist",
		modelsRK7API.OnlyActive("0"), //неактивные будем грохать в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Ident,ItemIdent,GUIDString,Code,Name,MainParentIdent,Status,Parent,genIDBX24,genSectionIDBX24,genWooID,genWooParentCategoryID)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	CateglistInRK7 := (Rk7QueryResultGetRefDataCateglist).(*modelsRK7API.RK7QueryResultGetRefDataCateglist)

	logger.Infof("Длина списка CateglistInRK7 = %d\n", len(CateglistInRK7.RK7Reference.Items.Item))

	if m.CateglistMapByIdent == nil {
		m.CateglistMapByIdent = make(map[int]*modelsRK7API.Categlist)
	}

	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		m.CateglistItemInRK7 = append(m.CateglistItemInRK7, CateglistInRK7.RK7Reference.Items.Item[i])
		m.CateglistMapByIdent[item.ItemIdent] = CateglistInRK7.RK7Reference.Items.Item[i]
	}
	logger.Infof("Создан CateglistMap, длина: %d", len(m.CateglistMapByIdent))

	return nil
}

func (m *menu) RefreshMenuitems() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshMenuitems")
	defer logger.Info("End RefreshMenuitems")
	cfg := config.GetConfig()
	RK7API := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7MID.User, cfg.RK7MID.Pass)

	//получить актуальное меню RK7
	logger.Info("Получить список всех блюд из RK7")
	Rk7QueryResultGetRefDataMenuitems, err := RK7API.GetRefData("Menuitems",
		modelsRK7API.OnlyActive("0"), //неактивные будем менять статус на N ?или может удалять? в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Code,Name,Ident,ItemIdent,GUIDString,MainParentIdent,ExtCode,PRICETYPES^3,CategPath,Status,genIDBX24,genSectionIDBX24)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}
	MenuInRK7 := (Rk7QueryResultGetRefDataMenuitems).(*modelsRK7API.RK7QueryResultGetRefDataMenuitems)
	logger.Infof("Длина списка MenuRK7MapByIdent = %d\n", len(MenuInRK7.RK7Reference.Items.Item))

	if m.MenuRK7MapByIdent == nil {
		m.MenuRK7MapByIdent = make(map[int]*modelsRK7API.MenuitemItem)
	}

	for i, item := range MenuInRK7.RK7Reference.Items.Item {
		m.MenuitemItemInRK7 = append(m.MenuitemItemInRK7, &MenuInRK7.RK7Reference.Items.Item[i])
		m.MenuRK7MapByIdent[item.ItemIdent] = &MenuInRK7.RK7Reference.Items.Item[i]
	}
	logger.Infof("Длина списка MenuRK7MapByIdent = %d\n", len(m.MenuRK7MapByIdent))

	return nil
}

func (m *menu) RefreshProducts() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshProducts")
	defer logger.Info("End RefreshProducts")
	cfg := config.GetConfig()
	WOOAPI := wooapi.NewAPI(cfg.WOOCOMMERCE.URL, cfg.WOOCOMMERCE.Key, cfg.WOOCOMMERCE.Secret)

	logger.Info("Получить список всех товаров из WOO")
	products, err := WOOAPI.ProductListAll()
	if err != nil {
		return errors.Wrap(err, "failed in WOOAPI.ProductListAll()")
	}

	if m.ProductsMapByID == nil {
		m.ProductsMapByID = make(map[int]*modelsWOOAPI.Product)
	}

	for i, product := range products {
		m.ProductsMapByID[product.ID] = products[i]
	}

	logger.Infof("Длина списка ProductsMapByID = %d\n", len(m.ProductsMapByID))

	return nil
}

func (m *menu) RefreshProductCategories() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshProductCategories")
	defer logger.Info("End RefreshProductCategories")
	cfg := config.GetConfig()
	WOOAPI := wooapi.NewAPI(cfg.WOOCOMMERCE.URL, cfg.WOOCOMMERCE.Key, cfg.WOOCOMMERCE.Secret)

	logger.Info("Получить список всех ProductCategories из WOO")
	productCategories, err := WOOAPI.ProductCategoryListAll()
	if err != nil {
		return errors.Wrap(err, "failed in WOOAPI.ProductCategoryListAll()")
	}

	if m.ProductCategoriesMapByID == nil {
		m.ProductCategoriesMapByID = make(map[int]*modelsWOOAPI.ProductCategory)
	}

	for i, productCategory := range productCategories {
		m.ProductCategoriesMapByID[productCategory.ID] = productCategories[i]
	}
	logger.Infof("Длина списка ProductCategoriesMapByID = %d\n", len(m.ProductCategoriesMapByID))

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

func NewCacheMenu() (CacheMenu, error) {

	logger := logging.GetLogger()
	logger.Info("Start NewCacheMenu")
	defer logger.Info("End NewCacheMenu")

	err := cacheMenuGlobal.RefreshMenu()
	if err != nil {
		return nil, errors.Wrap(err, "failed RefreshMenu()")
	}

	return &cacheMenuGlobal, nil
}

func GetCacheMenu() (CacheMenu, error) {

	logger := logging.GetLogger()
	logger.Info("Start GetCacheMenu")
	defer logger.Info("End GetCacheMenu")

	return &cacheMenuGlobal, nil
}
