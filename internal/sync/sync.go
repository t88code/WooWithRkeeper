package sync

import (
	"WooWithRkeeper/internal/bx24api"
	modelsBX24API "WooWithRkeeper/internal/bx24api/models"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"github.com/pkg/errors"
)

var menu Menu

type Menu struct {
	MenuitemItemInRK7 []*modelsRK7API.MenuitemItem
	MenuRK7MapByIdent map[int]*modelsRK7API.MenuitemItem

	CateglistItemInRK7  []*modelsRK7API.Categlist
	CateglistMapByIdent map[int]*modelsRK7API.Categlist

	ProductListMapByID map[string]*modelsBX24API.Product

	ProductSectionListMapByID map[string]*modelsBX24API.ProductSection
}

func (menu *Menu) RefreshCateglist() error {

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

	if menu.CateglistMapByIdent == nil {
		menu.CateglistMapByIdent = make(map[int]*modelsRK7API.Categlist)
	}

	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		menu.CateglistItemInRK7 = append(menu.CateglistItemInRK7, &CateglistInRK7.RK7Reference.Items.Item[i])
		menu.CateglistMapByIdent[item.ItemIdent] = &CateglistInRK7.RK7Reference.Items.Item[i]
	}
	logger.Infof("Создан CateglistMap, длина: %d", len(menu.CateglistMapByIdent))

	return nil
}

func (menu *Menu) RefreshMenuitems() error {

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

	if menu.MenuRK7MapByIdent == nil {
		menu.MenuRK7MapByIdent = make(map[int]*modelsRK7API.MenuitemItem)
	}

	for i, item := range MenuInRK7.RK7Reference.Items.Item {
		menu.MenuitemItemInRK7 = append(menu.MenuitemItemInRK7, &MenuInRK7.RK7Reference.Items.Item[i])
		menu.MenuRK7MapByIdent[item.ItemIdent] = &MenuInRK7.RK7Reference.Items.Item[i]
	}
	logger.Infof("Длина списка MenuRK7MapByIdent = %d\n", len(menu.MenuRK7MapByIdent))

	return nil
}

func (menu *Menu) RefreshProductList() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshProductList")
	defer logger.Info("End RefreshProductList")
	cfg := config.GetConfig()
	BX24API := bx24api.NewAPI(cfg.BX24.URL)

	//получить все элементы из BX24
	logger.Info("Получить список всех товаров из BX24")
	ProductList, err := BX24API.ProductList()
	if err != nil {
		return errors.Wrap(err, "failed in bx24api.ProductList()")
	}

	if menu.ProductListMapByID == nil {
		menu.ProductListMapByID = make(map[string]*modelsBX24API.Product)
	}

	for i, product := range ProductList {
		menu.ProductListMapByID[product.ID] = ProductList[i]
	}
	logger.Infof("Длина списка ProductListMapByID = %d\n", len(menu.ProductListMapByID))

	return nil
}

func (menu *Menu) RefreshProductSectionList() error {

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

	if menu.ProductSectionListMapByID == nil {
		menu.ProductSectionListMapByID = make(map[string]*modelsBX24API.ProductSection)
	}

	for i, productSection := range ProductSectionList {
		menu.ProductSectionListMapByID[productSection.ID] = ProductSectionList[i]
	}
	logger.Infof("Длина списка ProductSectionListMap = %d\n", len(menu.ProductSectionListMapByID))

	return nil
}

func (menu *Menu) RefreshMenu() error {

	logger := logging.GetLogger()
	logger.Info("Start RefreshMenu")
	defer logger.Info("End RefreshMenu")

	err := menu.RefreshMenuitems()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshMenuitems()")
	}

	err = menu.RefreshCateglist()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshCateglist()")
	}

	err = menu.RefreshProductList()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshProductList()")
	}

	err = menu.RefreshProductSectionList()
	if err != nil {
		return errors.Wrap(err, "failed in menu.RefreshProductSectionList()")
	}

	return nil
}

func NewMenu() (*Menu, error) {

	logger := logging.GetLogger()
	logger.Info("Start NewMenu")
	defer logger.Info("End NewMenu")

	err := menu.RefreshMenu()
	if err != nil {
		return nil, errors.Wrap(err, "failed RefreshMenu()")
	}

	return &menu, nil
}

func GetMenu() (*Menu, error) {

	logger := logging.GetLogger()
	logger.Info("Start GetMenu")
	defer logger.Info("End GetMenu")

	return &menu, nil
}
