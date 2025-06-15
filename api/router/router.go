package router

import (
	"net/http"
	"production-go-api-template/api/resource/health"
	"production-go-api-template/pkg/router"

	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *http.ServeMux {
	routerMux := http.NewServeMux()

	routerMux.HandleFunc("GET /livez", health.NewHealthHandler().CheckHandler)
	routerMux.HandleFunc("GET /healthz", health.HealthzHandler)

	itemsRouter := SetupItemRouter(db)
	router.Mount(routerMux, "/api/v1/items", itemsRouter)

	return routerMux
}
