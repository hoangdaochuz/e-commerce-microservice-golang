package order_cmd

import order_configs "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/configs"

func Start() {
	order_configs.NewOrderDatabase()
}
