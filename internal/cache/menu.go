package cache

import (
	"WooWithRkeeper/internal/bx24api"
	modelsBX24API "WooWithRkeeper/internal/bx24api/models"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"github.com/pkg/errors"
)

type CacheMenu interface {
	RefreshCateglist() error
	RefreshMenuitems() error
	RefreshProductList() error
	RefreshProductSectionList() error
	RefreshMenu() error

	GetMenuitems() ([]*modelsRK7API.MenuitemItem, error)
	GetMenuRK7MapByIdent() (map[int]*modelsRK7API.MenuitemItem, error)

	GetCateglistItemInRK7() ([]*modelsRK7API.Categlist, error)
	GetCateglistMapByIdent() (map[int]*modelsRK7API.Categlist, error)

	GetProductListMapByID() (map[string]*modelsBX24API.Product, error)

	GetProductSectionListMapByID() (map[string]*modelsBX24API.ProductSection, error)
}

var cacheMenuGlobal menu

type menu struct {
	MenuitemItemInRK7 []*modelsRK7API.MenuitemItem
	MenuRK7MapByIdent map[int]*modelsRK7API.MenuitemItem

	CateglistItemInRK7  []*modelsRK7API.Categlist
	CateglistMapByIdent map[int]*modelsRK7API.Categlist

	ProductListMapByID map[string]*modelsBX24API.Product

	ProductSectionListMapByID map[string]*modelsBX24API.ProductSection
}

func (m *menu) GetMenuitems() ([]*modelsRK7API.MenuitemItem, error) {
	// if version>0 || timeoute>0 {RefreshMenuitems()}
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

func (m *menu) GetProductListMapByID() (map[string]*modelsBX24API.Product, error) {
	return m.ProductListMapByID, nil
}

func (m *menu) GetProductSectionListMapByID() (map[string]*modelsBX24API.ProductSection, error) {
	return m.ProductSectionListMapByID, nil
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
		modelsRK7API.PropMask("items.(Ident,ItemIdent,GUIDString,Code,Name,MainParentIdent,Status,Parent,genIDBX24,genSectionIDBX24)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	CateglistInRK7 := (Rk7QueryResultGetRefDataCateglist).(*modelsRK7API.RK7QueryResultGetRefDataCateglist)

	logger.Infof("Длина списка CateglistInRK7 = %d\n", len(CateglistInRK7.RK7Reference.Items.Item))

	if m.CateglistMapByIdent == nil {
		m.CateglistMapByIdent = make(map[int]*modelsRK7API.Categlist)
	}

	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		m.CateglistItemInRK7 = append(m.CateglistItemInRK7, &CateglistInRK7.RK7Reference.Items.Item[i])
		m.CateglistMapByIdent[item.ItemIdent] = &CateglistInRK7.RK7Reference.Items.Item[i]
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

func (m *menu) RefreshProductList() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshProductList")
	defer logger.Info("End RefreshProductList")
	cfg := config.GetConfig()
	BX24API := bx24api.NewAPI(cfg.BX24.URL)

	//получить все элементы из Woocommerce
	logger.Info("Получить список всех товаров из WOO")
	ProductList, err := BX24API.ProductList()
	if err != nil {
		return errors.Wrap(err, "failed in bx24api.ProductList()")
	}

	if m.ProductListMapByID == nil {
		m.ProductListMapByID = make(map[string]*modelsBX24API.Product)
	}

	for i, product := range ProductList {
		m.ProductListMapByID[product.ID] = ProductList[i]
	}
	logger.Infof("Длина списка ProductListMapByID = %d\n", len(m.ProductListMapByID))

	return nil
}

func (m *menu) RefreshProductSectionList() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshProductSectionList")
	defer logger.Info("End RefreshProductSectionList")
	cfg := config.GetConfig()
	BX24API := bx24api.NewAPI(cfg.BX24.URL)

	//получить все ProductSection из BX24
	logger.Info("Получить список всех ProductSection из BX24")
	ProductSectionList, err := BX24API.ProductSectionList()
	if err != nil {
		return errors.Wrap(err, "failed in bx24api.ProductSectionList()")
	}

	if m.ProductSectionListMapByID == nil {
		m.ProductSectionListMapByID = make(map[string]*modelsBX24API.ProductSection)
	}

	for i, productSection := range ProductSectionList {
		m.ProductSectionListMapByID[productSection.ID] = ProductSectionList[i]
	}
	logger.Infof("Длина списка ProductSectionListMap = %d\n", len(m.ProductSectionListMapByID))

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

	err = m.RefreshProductList()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshProductList()")
	}

	err = m.RefreshProductSectionList()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshProductSectionList()")
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
