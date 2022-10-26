package cache

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type CacheOrder interface {
	Get(visitID int) (*modelsRK7API.Order, error)
	Set(order *modelsRK7API.Order) error
	Update(order *modelsRK7API.Order) error
	Delete(visitID int) error
}

var cacheOrderGlobal orders

type orders struct {
	orders map[int]*order
}

type order struct {
	rk7Order   *modelsRK7API.Order
	timeUpdate time.Time
}

// если визит не найден, то выполняется поиск в RK7
func (c *orders) Get(visitID int) (*modelsRK7API.Order, error) {

	logger := logging.GetLogger()
	logger.Info("Start CacheOrder Get")
	defer logger.Info("End CacheOrder Get")

	cfg := config.GetConfig()
	RK7API, err := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7MID.User, cfg.RK7MID.Pass)
	if err != nil {
		return nil, errors.New("failed rk7api.NewAPI()")
	}

	logger.Infof("VisitID=%d", visitID)

	if _, ok := c.orders[visitID]; ok {
		logger.Info("VisitID найден в кеше")
		if time.Now().Sub(c.orders[visitID].timeUpdate) > time.Second*time.Duration(cfg.CACHE.TimeUpdate) {
			getOrderList, err := RK7API.GetOrderList()
			if err != nil {
				return nil, errors.Wrap(err, "ошибка при получении списка заказов; RK7API.GetOrderList")
			}
			//поиск визита
			logger.Info("поиск визита в RK7")
			for _, visit := range getOrderList.Visit {
				if visit.VisitID == visitID {
					//визит найден
					logger.Info("визит найден в RK7")
					logger.Info("проверка заказа в визите")
					if len(visit.Orders.Order) > 0 {
						//заказ найден
						logger.Info("заказы есть в визите")
						orderGuid := visit.Orders.Order[len(visit.Orders.Order)-1].Guid
						getOrder, err := RK7API.GetOrder(orderGuid)
						if err != nil {
							return nil, errors.Wrapf(err, "ошибка при получении заказа; RK7API.GetOrder; visitID=%d, orderGuid=%s", visitID, orderGuid)
						}
						//обновить кеш
						err = c.Set(getOrder.Order)
						if err != nil {
							return nil, errors.Wrapf(err, "не удалось обновить кеш с visitID=%d", visitID)
						}
						logger.Info("кеш обновлен")
						return c.orders[visitID].rk7Order, nil
					} else {
						//заказ не найден
						errorText := fmt.Sprintf("заказ не найден в визите visitID=%d в RK7, возможно он был удален при закрытии смены, хотя визит существует; проверьте настройки RK7: должен быть включен параметр Один заказ на визит и отключен параметр Пустые визиты", visitID)
						logger.Info(errorText)
						return nil, errors.New(errorText)
					}
				}
			}
			//визит не найден
			errorText := fmt.Sprintf("ошибка, визит visitID=%d не найден в RK7, возможно он был удален при закрытии смены, хотя ранее визит в RK7 существовал", visitID)
			logger.Info(errorText)
			return nil, errors.New(errorText)
		} else {
			logger.Info("кеш актуальный, заказ берем из кеша")
			return c.orders[visitID].rk7Order, nil
		}
	} else {
		logger.Info("VisitID не найден в кеше")
		getOrderList, err := RK7API.GetOrderList()
		if err != nil {
			return nil, errors.Wrap(err, "ошибка при получении списка заказов; RK7API.GetOrderList")
		}
		//поиск визита
		logger.Info("поиск визита в RK7")
		for _, visit := range getOrderList.Visit {
			if visit.VisitID == visitID {
				//визит найден
				logger.Info("визит найден в RK7")
				logger.Info("проверка заказа в визите")
				if len(visit.Orders.Order) > 0 {
					//заказ найден //TODO GetORder кажется кеш постоянно обновялется в заказах
					logger.Info("заказы есть в визите")
					orderGuid := visit.Orders.Order[len(visit.Orders.Order)-1].Guid
					getOrder, err := RK7API.GetOrder(orderGuid)
					if err != nil {
						return nil, errors.Wrapf(err, "ошибка при получении заказа; RK7API.GetOrder; visitID=%d, orderGuid=%s", visitID, orderGuid)
					}
					//обновить кеш
					err = c.Set(getOrder.Order)
					if err != nil {
						return nil, errors.Wrapf(err, "не удалось обновить кеш с visitID=%d", visitID)
					}
					logger.Info("кеш обновлен")
					return c.orders[visitID].rk7Order, nil
				} else {
					//заказ не найден
					errorText := fmt.Sprintf("заказ не найден в RK7, возможно он был удален при закрытии смены, хотя визит с visitID=%d существует; проверьте настройки RK7: должен быть включен параметр Один заказ на визит и отключен параметр Пустые визиты", visitID)
					logger.Info(errorText)
					return nil, errors.New(errorText)
				}
			}
		}
		//визит не найден
		logger.Info("визит не найден в RK7")
		return nil, nil
	}
}

func (c *orders) Set(o *modelsRK7API.Order) error {
	visitID := o.Visit
	if c.orders[visitID] == nil {
		c.orders[visitID] = new(order)
	}
	c.orders[visitID].rk7Order = o
	c.orders[visitID].timeUpdate = time.Now()
	return nil
}

func (c *orders) Update(o *modelsRK7API.Order) error {
	visitID := o.Visit
	if c.orders[visitID] == nil {
		c.orders[visitID] = new(order)
	}
	c.orders[visitID].rk7Order = o
	c.orders[visitID].timeUpdate = time.Now()
	return nil
}

func (c *orders) Delete(visitID int) error {
	delete(c.orders, visitID)
	return nil
}

func NewCacheOrder() CacheOrder {
	cacheOrderGlobal.orders = make(map[int]*order)

	return &cacheOrderGlobal
}

func GetCacheOrder() CacheOrder {
	return &cacheOrderGlobal
}

func init() {
	NewCacheOrder()
}
