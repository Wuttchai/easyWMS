package main

import (
	"easywms-demo-v3/internal/config"
	"easywms-demo-v3/internal/database"
	"log"
)

func main() {
	db, err := database.Connect(config.Load())
	if err != nil { log.Fatal(err) }
	if err := database.SyncStockCountStatuses(db); err != nil { log.Fatal(err) }
	log.Print("Stock count statuses synchronized; inventory unchanged")
}
