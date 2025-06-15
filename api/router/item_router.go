package router

import (
	"net/http"
	"production-go-api-template/api/resource/item"

	"gorm.io/gorm"
)

type ItemHandler struct {
	DB  *gorm.DB
}

func NewItemHandler(db *gorm.DB) *ItemHandler {
	return &ItemHandler{DB: db}
}

func (h *ItemHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /", h.CreateItemHandler)
	mux.HandleFunc("GET /", h.GetAllItemsHandler)
	mux.HandleFunc("GET /{id}", h.GetItemHandler)
	mux.HandleFunc("PUT /{id}", h.UpdateItemHandler)
	mux.HandleFunc("DELETE /{id}", h.DeleteItemHandler)
}

func (h *ItemHandler) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	item.CreateItemHandler(h.DB, w, r)
}

func (h *ItemHandler) GetAllItemsHandler(w http.ResponseWriter, r *http.Request) {
	item.GetAllItemsHandler(h.DB, w, r)
}

func (h *ItemHandler) GetItemHandler(w http.ResponseWriter, r *http.Request) {
	item.GetItemHandler(h.DB, w, r)
}

func (h *ItemHandler) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	item.UpdateItemHandler(h.DB, w, r)
}

func (h *ItemHandler) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	item.DeleteItemHandler(h.DB, w, r)
}


func SetupItemRouter(db *gorm.DB) *http.ServeMux {
	itemRouter := http.NewServeMux()

	h := NewItemHandler(db)
	h.RegisterRoutes(itemRouter)

	return itemRouter
}
