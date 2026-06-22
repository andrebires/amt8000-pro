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
	options := []web.ServerOption{}
	if cfg.AuditPath != "" {
		options = append(options, web.WithAuditPath(cfg.AuditPath))
	}
	server := web.NewServer(func(conn web.PanelConnection) web.PanelClient {
		return isecnet.NewClient(conn.Host, conn.Port, conn.Password, 5*time.Second)
	}, options...)

	log.Printf("amt8000-pro listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, server.Routes()); err != nil {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}
