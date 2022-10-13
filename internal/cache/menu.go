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

type CacheMenu interface {
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
	GetProductsWooByRKeeperID() (map[int]*modelsWOOAPI.Product, error)
	GetProductCategoriesWooByID() (map[int]*modelsWOOAPI.ProductCategory, error)
	GetProductCategoriesWooByName() (map[string]*modelsWOOAPI.ProductCategory, error)
	GetProductCategoriesWooByRKeeperID() (map[int]*modelsWOOAPI.ProductCategory, error)
	DeleteProductCategoryFromCache(ID int) error
}

var cacheMenuGlobal cacheMenu

type cacheMenu struct {
	//RK7 - Menuitems
	MenuitemsRK7           []*modelsRK7API.MenuitemItem
	MenuitemsRK7ByIdent    map[int]*modelsRK7API.MenuitemItem
	versionMenuitemsRK7    int
	timeUpdateMenuitemsRK7 time.Time

	//RK7 - Categlists
	CateglistsRK7          []*modelsRK7API.Categlist
	CateglistsRK7ByIdent   map[int]*modelsRK7API.Categlist
	versionCateglistRK7    int
	timeUpdateCateglistRK7 time.Time

	//WOO - Product
	ProductsWooByID        map[int]*modelsWOOAPI.Product
	ProductsWooByRKeeperID map[int]*modelsWOOAPI.Product
	versionProductsWOO     int
	timeUpdateProductsWOO  time.Time

	//WOO - ProductCategories
	ProductCategoriesWooByID        map[int]*modelsWOOAPI.ProductCategory
	ProductCategoriesWooByName      map[string]*modelsWOOAPI.ProductCategory
	ProductCategoriesWooByRKeeperID map[int]*modelsWOOAPI.ProductCategory
	versionProductCategoriesWoo     int
	timeUpdateProductCategoriesWoo  time.Time
}

func (m *cacheMenu) DeleteProductCategoryFromCache(WOOID int) error {
	logger := logging.GetLogger()
	logger.Info("Start DeleteProductCategoryFromCache")
	defer logger.Info("End DeleteProductCategoryFromCache")

	var name string
	if pc, found := m.ProductCategoriesWooByID[WOOID]; found {
		name = pc.Name
	} else {
		return errors.New(fmt.Sprintf("не удалось удалить папку(id=%d) в woo; name не найден в кеше", WOOID))
	}

	var rkeeperID int
	if pc, found := m.ProductCategoriesWooByID[WOOID]; found {
		rkeeperID = pc.RkeeperID
	} else {
		return errors.New(fmt.Sprintf("не удалось удалить папку(id=%d) в woo; rkeeperID не найден в кеше", WOOID))
	}

	delete(m.ProductCategoriesWooByID, WOOID)
	delete(m.ProductCategoriesWooByName, name)
	delete(m.ProductCategoriesWooByRKeeperID, rkeeperID)

	return nil
}

func (m *cacheMenu) GetMenuitems() ([]*modelsRK7API.MenuitemItem, error) {
	// TODO if version>0 || timeoute>0 {RefreshMenuitems()}
	return m.MenuitemsRK7, nil
}

func (m *cacheMenu) GetMenuitemsRK7ByIdent() (map[int]*modelsRK7API.MenuitemItem, error) {
	return m.MenuitemsRK7ByIdent, nil
}

func (m *cacheMenu) GetCateglistRK7() ([]*modelsRK7API.Categlist, error) {
	return m.CateglistsRK7, nil
}

func (m *cacheMenu) GetCateglistsRK7ByIdent() (map[int]*modelsRK7API.Categlist, error) {
	return m.CateglistsRK7ByIdent, nil
}

func (m *cacheMenu) GetProductsWooByID() (map[int]*modelsWOOAPI.Product, error) {
	return m.ProductsWooByID, nil
}

func (m *cacheMenu) GetProductsWooByRKeeperID() (map[int]*modelsWOOAPI.Product, error) {
	return m.ProductsWooByRKeeperID, nil
}

func (m *cacheMenu) GetProductCategoriesWooByID() (map[int]*modelsWOOAPI.ProductCategory, error) {
	return m.ProductCategoriesWooByID, nil
}

func (m *cacheMenu) GetProductCategoriesWooByName() (map[string]*modelsWOOAPI.ProductCategory, error) {
	return m.ProductCategoriesWooByName, nil
}

func (m *cacheMenu) GetProductCategoriesWooByRKeeperID() (map[int]*modelsWOOAPI.ProductCategory, error) {
	return m.ProductCategoriesWooByRKeeperID, nil
}

func (m *cacheMenu) RefreshCateglist() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshCateglist")
	defer logger.Info("End RefreshCateglist")
	cfg := config.GetConfig()
	RK7API, err := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7MID.User, cfg.RK7MID.Pass)
	if err != nil {
		return errors.New("failed rk7api.NewAPI()")
	}

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

	if m.CateglistsRK7ByIdent == nil {
		m.CateglistsRK7ByIdent = make(map[int]*modelsRK7API.Categlist)
	}

	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		m.CateglistsRK7 = append(m.CateglistsRK7, &CateglistInRK7.RK7Reference.Items.Item[i])
		m.CateglistsRK7ByIdent[item.ItemIdent] = &CateglistInRK7.RK7Reference.Items.Item[i]
	}
	logger.Infof("Создан CateglistMap, длина: %d", len(m.CateglistsRK7ByIdent))

	return nil
}

func (m *cacheMenu) RefreshMenuitems() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshMenuitems")
	defer logger.Info("End RefreshMenuitems")
	cfg := config.GetConfig()
	RK7API, err := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7MID.User, cfg.RK7MID.Pass)
	if err != nil {
		return errors.New("failed rk7api.NewAPI()")
	}
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
	logger.Infof("Длина списка MenuitemsRK7ByIdent = %d\n", len(MenuInRK7.RK7Reference.Items.Item))

	if m.MenuitemsRK7ByIdent == nil {
		m.MenuitemsRK7ByIdent = make(map[int]*modelsRK7API.MenuitemItem)
	}

	for i, item := range MenuInRK7.RK7Reference.Items.Item {
		m.MenuitemsRK7 = append(m.MenuitemsRK7, &MenuInRK7.RK7Reference.Items.Item[i])
		m.MenuitemsRK7ByIdent[item.ItemIdent] = &MenuInRK7.RK7Reference.Items.Item[i]
	}
	logger.Infof("Длина списка MenuitemsRK7ByIdent = %d\n", len(m.MenuitemsRK7ByIdent))

	return nil
}

func (m *cacheMenu) RefreshProducts() error {

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

	if m.ProductsWooByID == nil {
		m.ProductsWooByID = make(map[int]*modelsWOOAPI.Product)
	}

	for i, product := range products {
		m.ProductsWooByID[product.ID] = products[i]
	}

	logger.Infof("Длина списка ProductsWooByID = %d\n", len(m.ProductsWooByID))

	return nil
}

func (m *cacheMenu) RefreshProductCategories() error {

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

	if m.ProductCategoriesWooByID == nil {
		m.ProductCategoriesWooByID = make(map[int]*modelsWOOAPI.ProductCategory)
	}

	if m.ProductCategoriesWooByName == nil {
		m.ProductCategoriesWooByName = make(map[string]*modelsWOOAPI.ProductCategory)
	}

	for i, productCategory := range productCategories {
		m.ProductCategoriesWooByID[productCategory.ID] = productCategories[i]
		m.ProductCategoriesWooByName[productCategory.Name] = productCategories[i]
	}

	logger.Infof("Длина списка ProductCategoriesWooByID = %d\n", len(m.ProductCategoriesWooByID))
	logger.Infof("Длина списка ProductCategoriesWooByName = %d\n", len(m.ProductCategoriesWooByName))

	return nil
}

func (m *cacheMenu) RefreshMenu() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshMenu")
	defer logger.Info("End RefreshMenu")

	err := m.RefreshMenuitems()
	if err != nil {
		return errors.Wrap(err, "failed in cacheMenu.RefreshMenuitems()")
	}

	err = m.RefreshCateglist()
	if err != nil {
		return errors.Wrap(err, "failed in cacheMenu.RefreshCateglist()")
	}

	err = m.RefreshProducts()
	if err != nil {
		return errors.Wrap(err, "failed in cacheMenu.RefreshProducts()")
	}

	err = m.RefreshProductCategories()
	if err != nil {
		return errors.Wrap(err, "failed in cacheMenu.RefreshProductCategories()")
	}

	return nil
}

func NewCacheMenu() (CacheMenu, error) {
	logger := logging.GetLogger()
	logger.Info("Start NewCacheMenu")
	defer logger.Info("End NewCacheMenu")

	return &cacheMenuGlobal, nil
}

func GetMenu() (CacheMenu, error) {
	logger := logging.GetLogger()
	logger.Info("Start GetMenu")
	defer logger.Info("End GetMenu")

	return &cacheMenuGlobal, nil
}
