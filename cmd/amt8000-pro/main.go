package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andrebires/amt8000-pro/internal/isecnet"
	"github.com/andrebires/amt8000-pro/internal/panel"
	"github.com/andrebires/amt8000-pro/internal/web"
)

func main() {
	cfg := panel.ConfigFromEnv()
	client := isecnet.NewClient(cfg.Host, cfg.Port, cfg.Password, 5*time.Second)
	server := web.NewServer(client)

	log.Printf("amt8000-pro listening on %s for panel %s:%d", cfg.HTTPAddr, cfg.Host, cfg.Port)
	if err := http.ListenAndServe(cfg.HTTPAddr, server.Routes()); err != nil {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}
