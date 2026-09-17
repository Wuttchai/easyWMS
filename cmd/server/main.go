package main

import (
	"easywms-demo-v3/internal/config"
	"easywms-demo-v3/internal/database"
	"easywms-demo-v3/internal/handlers"
	"easywms-demo-v3/internal/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.MigrateAndSeed(db); err != nil {
		log.Fatal(err)
	}
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}
	h := handlers.Handler{DB: db, JWTSecret: cfg.JWTSecret}
	routes.Register(r, h)
	log.Printf("EasyWMS Demo V3 running at http://localhost:%s", cfg.AppPort)
	log.Fatal(r.Run(":" + cfg.AppPort))
}
