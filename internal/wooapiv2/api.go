package wooapiv2

import (
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/hiscaler/woocommerce-go"
	"github.com/hiscaler/woocommerce-go/config"
	"os"
)

var clientGlobal *woocommerce.WooCommerce

func NewClient() *woocommerce.WooCommerce {

	logger := logging.GetLogger()
	logger.Info("Start NewClient")
	defer logger.Info("End NewClient")

	b, err := os.ReadFile("./config/config.woo.json")
	if err != nil {
		logger.Panic(fmt.Sprintf("Read config error: %s", err.Error())) // todo обработка
	}

	var c config.Config
	err = json.Unmarshal(b, &c)
	if err != nil {
		logger.Panic(fmt.Sprintf("Parse config file error: %s", err.Error()))
	}

	clientGlobal = woocommerce.NewClient(c)
	return clientGlobal
}

func GetClient() *woocommerce.WooCommerce {
	return clientGlobal
}
