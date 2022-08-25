package sync

import (
	"WooWithRkeeper/internal/bx24api"
	modelsBX24API "WooWithRkeeper/internal/bx24api/models"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	DB_NAME_SQLITE                 = "db.db"
	ERROR_PRODUCT_NOT_FOUND        = "API BX24: error_description: Product is not found; error: "
	ERROR_PRODUCTSECTION_NOT_FOUND = "API BX24: error_description: Раздел не найден.; error: "
)

//func SyncCategList()
func SyncCategList(rk7api rk7api.RK7API, bx24api bx24api.BX24API, db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Println("Start SyncCategList")
	defer logger.Println("End SyncCategList")

	var SyncMenuErrors []string

	//получить все ProductSection из BX24
	logger.Println("Получить список всех ProductSection из BX24")
	ProductSectionList, err := bx24api.ProductSectionList()
	if err != nil {
		return errors.Wrap(err, "failed in bx24api.ProductSectionList()")
	}
	var ProductSectionListMap = make(map[string]*modelsBX24API.ProductSection)
	for i, productSection := range ProductSectionList {
		ProductSectionListMap[productSection.ID] = ProductSectionList[i]
	}
	logger.Printf("Длина списка ProductSectionListMap = %d\n", len(ProductSectionListMap))

	//получить список всех Categlist из RK
	logger.Println("Получить список Categlist из RK7")
	Rk7QueryResultGetRefDataCateglist, err := rk7api.GetRefData("Categlist",
		modelsRK7API.OnlyActive("0"), //неактивные будем грохать в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Ident,ItemIdent,GUIDString,Code,Name,MainParentIdent,Status,Parent,genIDBX24,genSectionIDBX24)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	CateglistInRK7 := (Rk7QueryResultGetRefDataCateglist).(*modelsRK7API.RK7QueryResultGetRefDataCateglist)

	logger.Printf("Длина списка CateglistInRK7 = %d\n", len(CateglistInRK7.RK7Reference.Items.Item))

	var CateglistMap = make(map[int]modelsRK7API.Categlist)
	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		CateglistMap[item.ItemIdent] = CateglistInRK7.RK7Reference.Items.Item[i]
	}
	logger.Printf("Создан CateglistMap, длина: %d", len(CateglistMap))

	// 1 этап синхронизации - синхронизация папок меню без иерархии
	// пройтись по всем блюдам из меню кипера и сравнить с меню из битрикса24
	logger.Println("Запущен 1й этап синхронизации - синхронизация папок меню без иерархии")
	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		logger.Infof("Папка RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status)
		if item.Ident == 0 { //пропустить Categlist с Ident = 0
			logger.Infof("Обновление не требуется")
			continue
		}

		logger.Debugf("ProductSectionListMap(len = %d):", len(ProductSectionListMap))
		for key, value := range ProductSectionListMap {
			logger.Debugf("key: %s; value: %v", key, value)
		}

		if productSection, found := ProductSectionListMap[fmt.Sprint(item.ID_BX24)]; found {

			logger.Infof("Папка найдена в BX24. Name: %s, BX24_XML_ID: %s, BX24_ID: %s, BX24_Sextion_ID: %s", productSection.NAME, productSection.XMLID, productSection.ID, productSection.SECTIONID)

			switch item.Status {
			case 0: // удален - надо удалить в BX24
				err := bx24api.ProductSectionDelete(item.ID_BX24)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось удалить Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
					logger.Infof("Не удалось удалить Папку в BX24")
				} else {
					logger.Infof("Папка успешно удалена из BX24")
				}
				continue
			case 1: // черновик
				err := bx24api.ProductSectionDelete(item.ID_BX24)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось удалить Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
					logger.Infof("Не удалось удалить Папку в BX24")
				} else {
					logger.Infof("Папка успешно удалена из BX24")
				}
				continue
			case 2: //неактивный - надо удалить в BX24
				err := bx24api.ProductSectionDelete(item.ID_BX24)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось удалить Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
					logger.Infof("Не удалось удалить Папку в BX24")
				} else {
					logger.Infof("Папка успешно удалена из BX24")
				}
				continue
			case 3: //активный - продолжаем проверку
				logger.Infof("Требуется обновление")
			default:
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось определить статус Папки в RK: ident %d, name %s, status %d, sync 1", item.Ident, item.Name, item.Status))
				logger.Infof("Ошибка")
				continue
			}

			var SECTIONID int
			if productSection.SECTIONID == "" {
				SECTIONID = 0
			}

			// если папка совпадает, то пропускаем
			if productSection.NAME == item.Name &&
				productSection.XMLID == fmt.Sprint(item.ItemIdent) &&
				SECTIONID == item.SectionID_BX24 &&
				CateglistMap[item.MainParentIdent].ID_BX24 == item.SectionID_BX24 {
				logger.Info("Папка RK7 совпадает с BX24. Обновление в BX24 не требуется")
				continue
			}

			logger.Info("Папка не совпадает с BX24. Требуется обновление в BX24")
			err := bx24api.ProductSectionUpdate(item.ID_BX24,
				modelsBX24API.Name(item.Name),
				modelsBX24API.XMLID(fmt.Sprint(item.ItemIdent)))
			if err != nil {
				if err.Error() != ERROR_PRODUCTSECTION_NOT_FOUND {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v, sync: 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status, err))
					continue
				} else {
					//была ошибка ERROR_PRODUCTSECTION_NOT_FOUND - папка не найдена в BX24
					//следовательно папку необходимо создать, если она Активна(Status==3)
					if item.Status != 3 {
						continue
					} else {
						logger.Info("Папка не обновилась в BX24. Требуется ее создать в BX24")
						ProductSectionIDBX24, err := bx24api.ProductSectionAdd(item.Name,
							modelsBX24API.XMLID(fmt.Sprint(item.ItemIdent)))
						if err != nil {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync: 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
							continue
						} else if ProductSectionIDBX24 == 0 {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24, ProductSectionIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync: 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
							continue
						}

						logger.Info("Папка успешно создана в BX24")
						//ошибок нет, папка создана
						var categlist []*modelsRK7API.Categlist
						CateglistInRK7.RK7Reference.Items.Item[i].ID_BX24 = ProductSectionIDBX24
						categlist = append(categlist, &CateglistInRK7.RK7Reference.Items.Item[i])

						_, err = rk7api.SetRefDataCateglist(categlist)
						if err != nil {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync: 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
							continue
						}
						logger.Info("Папка успешно обновлена в RK7")
					}
				}
			} else {
				logger.Info("Папка успешно обновлена в BX24")
			}
		} else {

			logger.Info("Папка не найдена в BX24")

			//папка не найдена в BX24 и если статус не активный в RK7, то пропустить
			if item.Status != 3 {
				continue
			}

			logger.Info("Требуется создать папку в BX24")
			ProductSectionIDBX24, err := bx24api.ProductSectionAdd(item.Name,
				modelsBX24API.XMLID(fmt.Sprint(item.ItemIdent)))
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v, sync: 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status, err))
				continue
			} else if ProductSectionIDBX24 == 0 {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24, ProductSectionIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync: 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
				continue
			}

			logger.Info("Папка успешно создана в BX24")
			var categlist []*modelsRK7API.Categlist
			CateglistInRK7.RK7Reference.Items.Item[i].ID_BX24 = ProductSectionIDBX24
			categlist = append(categlist, &CateglistInRK7.RK7Reference.Items.Item[i])

			_, err = rk7api.SetRefDataCateglist(categlist)
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync: 1", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
				continue
			}
			logger.Info("Папка успешно обновлена в RK7")
		}
	}

	if len(SyncMenuErrors) > 0 {
		return errors.New(strings.Join(SyncMenuErrors, "\n"))
	}

	// 2 этап синхронизации - синхронизация иерархии папок меню
	// создать CateglistMap
	logger.Println("Запущен 2й этап синхронизации - синхронизация иерархии папок меню")

	// обновить SectionID_BX24 ProductSection и Categlist
	logger.Println("Запущен процесс обновления SectionID_BX24 в ProductSection и Categlist")
	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		logger.Infof("Папка RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync: 2", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status)
		if item.Status != 3 {
			logger.Infof("Обновление не требуется")
			continue
		}

		MainParentIdent := item.MainParentIdent

		var SectionID_BX24 string
		logger.Info("Требуется обновить SectionID в BX24")
		if CateglistMap[MainParentIdent].ID_BX24 != 0 {
			SectionID_BX24 = fmt.Sprint(CateglistMap[MainParentIdent].ID_BX24)
		} else {
			SectionID_BX24 = ""
		}

		err := bx24api.ProductSectionUpdate(item.ID_BX24, modelsBX24API.SectionID(SectionID_BX24))
		if err != nil {
			SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v, sync: 2", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status, err))
			continue
		}

		logger.Info("Папка успешно обновлена в BX24")
		var categlist []*modelsRK7API.Categlist
		CateglistInRK7.RK7Reference.Items.Item[i].SectionID_BX24 = CateglistMap[MainParentIdent].ID_BX24
		categlist = append(categlist, &CateglistInRK7.RK7Reference.Items.Item[i])

		_, err = rk7api.SetRefDataCateglist(categlist)
		if err != nil {
			SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, sync: 2", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status)) //TODO добавить error, sync/1/2
			continue
		}
		logger.Info("Папка успешно обновлена в RK7")
	}

	if len(SyncMenuErrors) > 0 {
		return errors.New(strings.Join(SyncMenuErrors, "\n"))
	}

	VersionRefName, err := GetVersion(rk7api, "Categlist")
	if err != nil {
		return errors.Wrapf(err, "failed GetVersion(rk7api, %s)", "Categlist")
	}

	err = UpdateVersionInDB(db, "Categlist", VersionRefName)
	if err != nil {
		return errors.Wrapf(err, "failed UpdateVersionInDB(db, %s, %d)", "Categlist", VersionRefName)
	}

	return nil
}

//func SyncMenuitems()
func SyncMenuitems(rk7api rk7api.RK7API, bx24api bx24api.BX24API, db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Println("Start SyncMenuitems")
	defer logger.Println("End SyncMenuitems")

	var SyncMenuErrors []string

	//получить все элементы из BX24
	logger.Println("Получить список всех блюд из BX24")
	ProductList, err := bx24api.ProductList()
	if err != nil {
		return errors.Wrap(err, "failed in bx24api.ProductList()")
	}

	var ProductListMap = make(map[string]*modelsBX24API.Product)
	for i, product := range ProductList {
		ProductListMap[product.ID] = ProductList[i]
	}
	logger.Printf("Длина списка ProductListMap = %d\n", len(ProductListMap))

	//получить список всех блюд из RK
	logger.Println("Получить список всех блюд из RK7")
	Rk7QueryResultGetRefDataMenuitems, err := rk7api.GetRefData("Menuitems",
		modelsRK7API.OnlyActive("0"), //неактивные будем менять статус на N ?или может удалять? в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Code,Name,Ident,ItemIdent,GUIDString,MainParentIdent,ExtCode,PRICETYPES^3,CategPath,Status,genIDBX24,genSectionIDBX24)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	// PRICETYPES-3="9223372036854775807" == PRICE="0.00"

	MenuInRK7 := (Rk7QueryResultGetRefDataMenuitems).(*modelsRK7API.RK7QueryResultGetRefDataMenuitems)
	logger.Printf("Длина списка MenuInRK7 = %d\n", len(MenuInRK7.RK7Reference.Items.Item))

	//получить список всех Categlist из RK
	logger.Println("Получить список Categlist из RK7")
	Rk7QueryResultGetRefDataCateglist, err := rk7api.GetRefData("Categlist",
		modelsRK7API.OnlyActive("0"), //неактивные будем грохать в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Ident,ItemIdent,GUIDString,Code,Name,MainParentIdent,Status,Parent,genIDBX24,genSectionIDBX24)"))
	if err != nil {
		return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	CateglistInRK7 := (Rk7QueryResultGetRefDataCateglist).(*modelsRK7API.RK7QueryResultGetRefDataCateglist)

	logger.Printf("Длина списка CateglistInRK7 = %d\n", len(CateglistInRK7.RK7Reference.Items.Item))

	var CateglistMap = make(map[int]modelsRK7API.Categlist)
	for i, item := range CateglistInRK7.RK7Reference.Items.Item {
		CateglistMap[item.ItemIdent] = CateglistInRK7.RK7Reference.Items.Item[i]
	}
	logger.Printf("Создан CateglistMap, длина: %d", len(CateglistMap))

	for i, item := range MenuInRK7.RK7Reference.Items.Item {
		logger.Infof("Блюдо RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, RK_Price: %d", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status, item.PRICETYPES3)
		if item.Ident == 0 { //пропустить блюдо с Ident = 0
			logger.Infof("Обновление не требуется")
			continue
		}

		var status string
		switch item.Status {
		case 0: //удален - если блюдо будет найдено в BX24, то удалить
			logger.Infof("Блюдо удалено в BX24")
		case 1: //черновик
			logger.Infof("Блюдо удалено в BX24")
		case 2: //неактивный
			logger.Infof("Блюдо удалено в BX24")
		case 3: //статус активный
			status = "Y"
		default:
			SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось определить статус Блюда в RK: ident %d, name %s, status %d", item.Ident, item.Name, item.Status))
			logger.Infof("Ошибка")
			continue
		}

		var pricetype3 string
		if item.PRICETYPES3 == 9223372036854775807 {
			pricetype3 = "0.00"
		} else {
			p := fmt.Sprint(item.PRICETYPES3)
			pricetype3 = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
		}

		var Categlist_SectionID int = CateglistMap[item.MainParentIdent].ID_BX24

		if product, found := ProductListMap[fmt.Sprint(item.ID_BX24)]; found {

			logger.Infof("Блюдо найдено в BX24. Name: %s, BX24_XML_ID: %s, BX24_ID: %s, BX24_Sextion_ID: %s, BX24_Active: %s, BX24_Price: %s", product.NAME, product.XMLID, product.ID, product.SECTIONID, product.ACTIVE, product.PRICE)

			//если блюдо удалено, то удалить в BX24
			if item.Status == 0 || item.Status == 1 || item.Status == 2 {
				err := bx24api.ProductDel(item.ID_BX24)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось удалить Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status))
					logger.Infof("Не удалось удалить Блюдо в BX24")
				} else {
					logger.Infof("Блюдо успешно удалено из BX24")
				}
				continue
			}

			var Product_SectionID string = product.SECTIONID
			if product.SECTIONID == "" {
				Product_SectionID = "0"
			}

			if item.SectionID_BX24 != Categlist_SectionID {
				var menuitem []*modelsRK7API.MenuitemItem
				MenuInRK7.RK7Reference.Items.Item[i].SectionID_BX24 = Categlist_SectionID
				item.SectionID_BX24 = Categlist_SectionID
				menuitem = append(menuitem, &MenuInRK7.RK7Reference.Items.Item[i])

				_, err = rk7api.SetRefDataMenuitems(menuitem)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status, err))
					continue
				}
				logger.Infof("Блюдо успешно обновлено в RK7. SectionID_BX24: %d", Categlist_SectionID)
			}

			if product.NAME == item.Name &&
				product.ACTIVE == status &&
				product.XMLID == fmt.Sprint(item.ItemIdent) &&
				Product_SectionID == fmt.Sprint(item.SectionID_BX24) &&
				product.PRICE == pricetype3 {
				logger.Info("Блюдо RK7 совпадает с BX24. Обновление в BX24 не требуется")
				continue
			}

			logger.Info("Блюдо не совпадает с BX24. Требуется обновление в BX24")
			err := bx24api.ProductUpdate(item.ID_BX24,
				modelsBX24API.Name(item.Name),
				modelsBX24API.Active(status),
				modelsBX24API.XMLID(fmt.Sprint(item.ItemIdent)),
				modelsBX24API.SectionID(fmt.Sprint(Categlist_SectionID)),
				modelsBX24API.Price(pricetype3))
			if err != nil {
				if err.Error() != ERROR_PRODUCT_NOT_FOUND {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status, err))
					continue
				} else {
					if item.Status != 3 {
						continue
					} else {
						logger.Info("Блюдо не обновилось в BX24. Требуется его создать в BX24")
						ProductIDBX24, err := bx24api.ProductAdd(item.Name,
							modelsBX24API.Active(status),
							modelsBX24API.XMLID(fmt.Sprint(item.ItemIdent)),
							modelsBX24API.SectionID(fmt.Sprint(Categlist_SectionID)),
							modelsBX24API.Price(pricetype3))
						if err != nil {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status, err))
							continue
						} else if ProductIDBX24 == 0 {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в BX24, ProductIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status))
							continue
						}

						logger.Info("Блюдо успешно создано в BX24")

						var menuitem []*modelsRK7API.MenuitemItem
						MenuInRK7.RK7Reference.Items.Item[i].ID_BX24 = ProductIDBX24
						menuitem = append(menuitem, &MenuInRK7.RK7Reference.Items.Item[i])

						_, err = rk7api.SetRefDataMenuitems(menuitem)
						if err != nil {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status, err))
							continue
						}
						logger.Info("Блюдо успешно обновлено в RK7")
					}
				}
			} else {
				logger.Info("Блюдо успешно обновлено в BX24")
			}
		} else {

			logger.Info("Блюдо не найдено в BX24")

			if item.Status != 3 {
				continue
			}

			logger.Info("Требуется создать блюдо в BX24")
			ProductIDBX24, err := bx24api.ProductAdd(item.Name,
				modelsBX24API.Active(status),
				modelsBX24API.XMLID(fmt.Sprint(item.ItemIdent)),
				modelsBX24API.SectionID(fmt.Sprint(Categlist_SectionID)),
				modelsBX24API.Price(pricetype3))
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status, err))
				continue
			} else if ProductIDBX24 == 0 {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в BX24, ProductIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status))
				continue
			}
			var menuitem []*modelsRK7API.MenuitemItem
			MenuInRK7.RK7Reference.Items.Item[i].ID_BX24 = ProductIDBX24
			menuitem = append(menuitem, &MenuInRK7.RK7Reference.Items.Item[i])

			_, err = rk7api.SetRefDataMenuitems(menuitem)
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, error: %v", item.Name, item.ItemIdent, item.ID_BX24, Categlist_SectionID, item.Status, err))
				continue
			}
			logger.Info("Блюдо успешно обновлено в RK7")
		}
	}

	if len(SyncMenuErrors) > 0 {
		return errors.New(strings.Join(SyncMenuErrors, "\n"))
	}

	VersionRefName, err := GetVersion(rk7api, "Menuitems")
	if err != nil {
		return errors.Wrapf(err, "failed GetVersion(rk7api, %s)", "Menuitems")
	}

	err = UpdateVersionInDB(db, "Menuitems", VersionRefName)
	if err != nil {
		return errors.Wrapf(err, "failed UpdateVersionInDB(db, %s, %d)", "Menuitems", VersionRefName)
	}

	return nil
}

//func SyncMenuService()
func SyncMenuService() {
	// TODO сделать метод который принудительно все обновляет - меню в битриксе, базу локальную обнуляет

	logger := logging.GetLogger()
	logger.Println("Start Service SyncMenu")
	defer logger.Println("End Service SyncMenu")

	cfg := config.GetConfig()
	RK7API := rk7api.NewAPI(cfg.RK7.URL, cfg.RK7.User, cfg.RK7.Pass)
	BX24API := bx24api.NewAPI(cfg.BX24.URL)

	if Exists(DB_NAME_SQLITE) != true {
		logger.Print(DB_NAME_SQLITE, " not exist")
		err := CreateDB()
		if err != nil {
			logger.Fatalf("%s, %v", DB_NAME_SQLITE, err)
		}
	} else {
		logger.Print(DB_NAME_SQLITE, " exist")
	}

	db, err := sqlx.Connect("sqlite3", DB_NAME_SQLITE)
	if err != nil {
		logger.Fatalf("failed sqlx.Connect, err: %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	_, err = NewMenu()
	if err != nil {
		errorText := fmt.Sprintf("Ошибка при попытке получить справочники меню RK и товаров BX24, err: %v", err)
		logger.Fatalf(errorText)
		err := telegram.SendMessage(errorText)
		if err != nil {
			logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
		}
	}

	for {
		// сверить справочники Categlist
		verifyVersionResult, err := VerifyVersion(RK7API, db, "Categlist")
		if err != nil {
			errorText := fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err)
			logger.Println(errorText)
			err := telegram.SendMessage(errorText)
			if err != nil {
				logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
			}
		} else if verifyVersionResult {
			logger.Info("Версия справочников Categlist совпадает между RK и DB")
			logger.Info("Запускаем проверку Categlist и ProductSection")
			resultVerifyCateglistWithProductSection, err := verifyCateglistWithProductSection()
			if err != nil {
				errorText := fmt.Sprintf("Ошибка при сверке Categlist с ProductSection: \n%v\n", err)
				logger.Info(errorText)
				err := telegram.SendMessage(errorText)
				if err != nil {
					logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
				}
			} else if resultVerifyCateglistWithProductSection {
				logger.Println("Обновление ProductSection не требуется")
			} else {
				logger.Println("Требуется обновление ProductSection")
				err := SyncCategList(RK7API, BX24API, db)
				if err != nil {
					errorText := fmt.Sprintf("Ошибка при синхронизации Categlist SyncMenu: \n%v\n", err)
					logger.Println(errorText)
					err := telegram.SendMessage(errorText)
					if err != nil {
						logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
					}
				} else {
					logger.Printf("Синхронизация Categlist выполнена успешно. Версия справочника Categlist в DB обновлена")
				}
			}
		} else {
			logger.Println("Требуется обновление Categlist")
			err := SyncCategList(RK7API, BX24API, db)
			if err != nil {
				errorText := fmt.Sprintf("Ошибка при синхронизации Categlist SyncMenu: \n%v\n", err)
				logger.Println(errorText)
				err := telegram.SendMessage(errorText)
				if err != nil {
					logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
				}
			} else {
				logger.Printf("Синхронизация Categlist выполнена успешно. Версия справочника Categlist в DB обновлена")
			}
		}

		// сверить справочники Menuitems
		verifyVersionResult, err = VerifyVersion(RK7API, db, "Menuitems")
		if err != nil {
			errorText := fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err)
			logger.Println(errorText)
			err := telegram.SendMessage(errorText)
			if err != nil {
				logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
			}
		} else if verifyVersionResult {
			logger.Info("Версия справочников Menuitems совпадает между RK и DB")
			logger.Info("Запускаем проверку Menuitems и Products")
			resultVerifyMenuitemsWithProducts, err := verifyMenuitemsWithProducts()
			if err != nil {
				errorText := fmt.Sprintf("Ошибка при сверке Menuitems и Products: \n%v\n", err)
				logger.Info(errorText)
				err := telegram.SendMessage(errorText)
				if err != nil {
					logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
				}
			} else if resultVerifyMenuitemsWithProducts {
				logger.Println("Обновление Products не требуется")
			} else {
				logger.Println("Требуется обновление меню")
				err := SyncMenuitems(RK7API, BX24API, db)
				if err != nil {
					errorText := fmt.Sprintf("Ошибка при синхронизации меню SyncMenu: \n%v\n", err)
					logger.Println(errorText)
					err := telegram.SendMessage(errorText)
					if err != nil {
						logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
					}
				} else {
					logger.Printf("Синхронизация меню выполнена успешно. Версия справочника Menuitems в DB обновлена")
				}
			}
		} else {
			logger.Println("Требуется обновление меню")
			err := SyncMenuitems(RK7API, BX24API, db)
			if err != nil {
				errorText := fmt.Sprintf("Ошибка при синхронизации меню SyncMenu: \n%v\n", err)
				logger.Println(errorText)
				err := telegram.SendMessage(errorText)
				if err != nil {
					logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
				}
			} else {
				logger.Printf("Синхронизация меню выполнена успешно. Версия справочника Menuitems в DB обновлена")
			}
		}

		logger.Printf("time sleep %d minuts\n", cfg.MENUSYNC.Timeout)

		time.Sleep(time.Minute * time.Duration(cfg.MENUSYNC.Timeout))
	}
}

func verifyCateglistWithProductSection() (bool, error) {
	logger := logging.GetLogger()
	logger.Println("Start verifyCateglistWithProductSection")
	defer logger.Println("End verifyCateglistWithProductSection")

	menu, err := GetMenu()
	if err != nil {
		return false, errors.New("Ошибка при попытке получить меню из кеша")
	}

	//сравнить количество не получится, потому что в BX24 могут создать папку, а в RK есть удаленные папки
	//сравнить содержание позиций
	for _, item := range menu.CateglistItemInRK7 {
		logger.Infof("Папка RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status)
		if item.Ident == 0 {
			logger.Infof("Categlist с Ident = 0 пропускаем")
			continue
		}

		logger.Debugf("ProductSectionListMap(len = %d):", len(menu.ProductSectionListMapByID))
		for key, value := range menu.ProductSectionListMapByID {
			logger.Debugf("key: %s; value: %v", key, value)
		}

		switch item.Status {
		case 0:
			logger.Infof("Папка удалена в RK - пропускаем")
			continue
		case 1:
			logger.Infof("Папка в черновике в RK - пропускаем")
			continue
		case 2:
			logger.Infof("Папка неактивна в RK - пропускаем")
			continue
		case 3: //активный - продолжаем проверку
			logger.Infof("Папка неактивна в RK - продолжаем проверку")
		default:
			return false, errors.New("Неизвестный статус у папки")
		}

		if productSection, found := menu.ProductSectionListMapByID[fmt.Sprint(item.ID_BX24)]; found {

			logger.Infof("Папка найдено в BX24. Name: %s, BX24_XML_ID: %s, BX24_ID: %s, BX24_Sextion_ID: %s", productSection.NAME, productSection.XMLID, productSection.ID, productSection.SECTIONID)

			var SECTIONID int
			if productSection.SECTIONID == "" {
				SECTIONID = 0
			}

			// если папка совпадает, то пропускаем
			if productSection.NAME == item.Name &&
				productSection.XMLID == fmt.Sprint(item.ItemIdent) &&
				SECTIONID == item.SectionID_BX24 &&
				menu.CateglistMapByIdent[item.MainParentIdent].ID_BX24 == item.SectionID_BX24 {
				logger.Info("Папка RK7 совпадает с BX24. Пропускаем")
				continue
			}

			logger.Info("Папка не совпадает с BX24. Запускаем процесс обновления.")
			return false, nil
		} else {
			logger.Info("Папка не найдена в системе BX24. Запускаем процесс обновления.")
			return false, nil
		}
	}
	logger.Info("Сверка папок завершена. Обновление не требуется.")
	return true, nil
}

func verifyMenuitemsWithProducts() (bool, error) {
	logger := logging.GetLogger()
	logger.Println("Start verifyMenuitemsWithProducts")
	defer logger.Println("End verifyMenuitemsWithProducts")

	menu, err := GetMenu()
	if err != nil {
		return false, errors.New("Ошибка при попытке получить меню из кеша")
	}

	//сравнить количество не получится, потому что в BX24 могут создать папку, а в RK есть удаленные папки
	//сравнить содержание позиций
	for _, item := range menu.MenuitemItemInRK7 {
		logger.Infof("Блюдо RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d, RK_Price: %d", item.Name, item.ItemIdent, item.ID_BX24, item.SectionID_BX24, item.Status, item.PRICETYPES3)
		if item.Ident == 0 {
			logger.Infof("MenuitemItem с Ident = 0 пропускаем")
			continue
		}

		var status string
		switch item.Status {
		case 0: //удален - если блюдо будет найдено в BX24, то удалить
			logger.Infof("Блюдо удалено в RK. Пропускаем")
			continue
		case 1: //черновик
			logger.Infof("Блюдо удалено в RK. Пропускаем")
			continue
		case 2: //неактивный
			logger.Infof("Блюдо удалено в RK. Пропускаем")
			continue
		case 3: //статус активный
			logger.Infof("Блюдо активно в RK. Продолжаем проверку")
			status = "Y"
		default:
			return false, errors.New("Неизвестный статус у блюда")
		}

		var pricetype3 string
		if item.PRICETYPES3 == 9223372036854775807 {
			pricetype3 = "0.00"
		} else {
			p := fmt.Sprint(item.PRICETYPES3)
			pricetype3 = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
		}

		var Categlist_SectionID int = menu.CateglistMapByIdent[item.MainParentIdent].ID_BX24

		if product, found := menu.ProductListMapByID[fmt.Sprint(item.ID_BX24)]; found {

			logger.Infof("Блюдо найдено в BX24. Name: %s, BX24_XML_ID: %s, BX24_ID: %s, BX24_Sextion_ID: %s, BX24_Active: %s, BX24_Price: %s", product.NAME, product.XMLID, product.ID, product.SECTIONID, product.ACTIVE, product.PRICE)

			var Product_SectionID string = product.SECTIONID
			if product.SECTIONID == "" {
				Product_SectionID = "0"
			}

			if item.SectionID_BX24 != Categlist_SectionID {
				logger.Infof("У блюдо RK7 не совпадает Categlist_SectionID=%d с BX24. Требуется обновление.", Categlist_SectionID)
				return false, nil
			}

			if product.NAME == item.Name &&
				product.ACTIVE == status &&
				product.XMLID == fmt.Sprint(item.ItemIdent) &&
				Product_SectionID == fmt.Sprint(item.SectionID_BX24) &&
				product.PRICE == pricetype3 {
				logger.Info("Блюдо RK7 совпадает с BX24. Обновление в BX24 не требуется")
				continue
			}

			logger.Info("Блюдо не совпадает с BX24. Требуется обновление в BX24.")
			return false, nil
		} else {
			logger.Info("Блюдо не найдено в BX24. Требуется обновление.")
			return false, nil
		}
	}
	logger.Info("Сверка блюд завершена. Обновление не требуется.")
	return true, nil
}

// TODO сделать ручник! на обновление меню

//func SyncMenuService1()
func SyncMenuService1() {
	// TODO сделать метод который принудительно все обновляет - меню в битриксе, базу локальную обнуляет

	logger := logging.GetLogger()
	logger.Println("Start Service SyncMenu")
	defer logger.Println("End Service SyncMenu")

	cfg := config.GetConfig()
	RK7API := rk7api.NewAPI(cfg.RK7.URL, cfg.RK7.User, cfg.RK7.Pass)
	BX24API := bx24api.NewAPI(cfg.BX24.URL)

	if Exists(DB_NAME_SQLITE) != true {
		logger.Print(DB_NAME_SQLITE, " not exist")
		err := CreateDB()
		if err != nil {
			logger.Fatalf("%s, %v", DB_NAME_SQLITE, err)
		}
	} else {
		logger.Print(DB_NAME_SQLITE, " exist")
	}

	db, err := sqlx.Connect("sqlite3", DB_NAME_SQLITE)
	if err != nil {
		logger.Fatalf("failed sqlx.Connect, err: %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	for {
		// сверить справочники Categlist
		verifyVersionResult, err := VerifyVersion(RK7API, db, "Categlist")
		if err != nil {
			errorText := fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err)
			logger.Println(errorText)
			err := telegram.SendMessage(errorText)
			if err != nil {
				logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
			}
		} else if verifyVersionResult {
			logger.Println("Обновление Categlist не требуется")

		} else {

			logger.Println("Требуется обновление Categlist")
			err := SyncCategList(RK7API, BX24API, db)
			if err != nil {
				errorText := fmt.Sprintf("Ошибка при синхронизации Categlist SyncMenu: \n%v\n", err)
				logger.Println(errorText)
				err := telegram.SendMessage(errorText)
				if err != nil {
					logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
				}
			} else {
				logger.Printf("Синхронизация Categlist выполнена успешно. Версия справочника Categlist в DB обновлена")
			}
		}

		// сверить справочники Menuitems
		verifyVersionResult, err = VerifyVersion(RK7API, db, "Menuitems")
		if err != nil {
			errorText := fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err)
			logger.Println(errorText)
			err := telegram.SendMessage(errorText)
			if err != nil {
				logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
			}
		} else if verifyVersionResult {
			// версия не сменилась, тогда проверить может количество сменилось

			logger.Println("Обновление меню не требуется")
		} else {
			logger.Println("Требуется обновление меню")
			err := SyncMenuitems(RK7API, BX24API, db)
			if err != nil {
				errorText := fmt.Sprintf("Ошибка при синхронизации меню SyncMenu: \n%v\n", err)
				logger.Println(errorText)
				err := telegram.SendMessage(errorText)
				if err != nil {
					logger.Printf("Не удалось отправить сообщение в телеграм:\n%s\n", errorText)
				}
			} else {
				logger.Printf("Синхронизация меню выполнена успешно. Версия справочника Menuitems в DB обновлена")
			}
		}

		logger.Printf("time sleep %d minuts\n", cfg.MENUSYNC.Timeout)

		time.Sleep(time.Minute * time.Duration(cfg.MENUSYNC.Timeout))
	}
}
