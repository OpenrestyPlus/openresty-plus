package main

import (
	"context"
	"log"

	"net.daoke/orp-backend/internal/config"
	"net.daoke/orp-backend/internal/httpapi"
	"net.daoke/orp-backend/internal/store"
)

func main() {
	configuration, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	database, err := store.Open(configuration.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := store.Migrate(database); err != nil {
		log.Fatal(err)
	}
	if err := httpapi.SeedDemoAlertData(context.Background(), database); err != nil {
		log.Fatal(err)
	}
	log.Println("告警演示数据已准备：含本地 Webhook、飞书配置占位和 10 天后到期的演示证书。通道默认停用，不会自动外发。")
}
