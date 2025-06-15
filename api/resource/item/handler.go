package item

import (
	"fmt"
	"net/http"
	"production-go-api-template/config"
	"production-go-api-template/pkg/contextkeys"
	"production-go-api-template/pkg/logger"
	"production-go-api-template/pkg/router"
	"production-go-api-template/pkg/validator"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func CreateItemHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	service, err := serviceFromRequest(db, r)
	if err != nil {
		router.RespondWithError(r, w, http.StatusInternalServerError, "service initialization failed", err)
		return
	}

	var req CreateItemRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		router.RespondWithError(r, w, http.StatusBadRequest, "invalid input", err)
		return
	}

	item, err := service.CreateItem(r.Context(), req)
	if err != nil {
		router.RespondWithError(r, w, http.StatusBadRequest, "failed to create item", err)
		return
	}

	router.RespondWithJSON(r, w, http.StatusCreated, ItemResponse{Item: item})
}

func GetItemHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	service, err := serviceFromRequest(db, r)
	if err != nil {
		router.RespondWithError(r, w, http.StatusInternalServerError, "service initialization failed", err)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		router.RespondWithError(r, w, http.StatusBadRequest, "item ID is required", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		router.RespondWithError(r, w, http.StatusBadRequest, "invalid item ID", err)
		return
	}

	item, err := service.GetItem(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			router.RespondWithError(r, w, http.StatusNotFound, "item not found", err)
			return
		}
		router.RespondWithError(r, w, http.StatusInternalServerError, "failed to get item", err)
		return
	}

	router.RespondWithJSON(r, w, http.StatusOK, ItemResponse{Item: item})
}

func GetAllItemsHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	service, err := serviceFromRequest(db, r)
	if err != nil {
		router.RespondWithError(r, w, http.StatusInternalServerError, "service initialization failed", err)
		return
	}

	items, err := service.GetAllItems(r.Context())
	if err != nil {
		router.RespondWithError(r, w, http.StatusInternalServerError, "failed to get items", err)
		return
	}

	response := ItemsResponse{
		Items: items,
		Total: len(items),
	}

	router.RespondWithJSON(r, w, http.StatusOK, response)
}

func UpdateItemHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	service, err := serviceFromRequest(db, r)
	if err != nil {
		router.RespondWithError(r, w, http.StatusInternalServerError, "service initialization failed", err)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		router.RespondWithError(r, w, http.StatusBadRequest, "item ID is required", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		router.RespondWithError(r, w, http.StatusBadRequest, "invalid item ID", err)
		return
	}

	var req UpdateItemRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		router.RespondWithError(r, w, http.StatusBadRequest, "invalid input", err)
		return
	}

	item, err := service.UpdateItem(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			router.RespondWithError(r, w, http.StatusNotFound, "item not found", err)
			return
		}
		router.RespondWithError(r, w, http.StatusBadRequest, "failed to update item", err)
		return
	}

	router.RespondWithJSON(r, w, http.StatusOK, ItemResponse{Item: item})
}

func DeleteItemHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	service, err := serviceFromRequest(db, r)
	if err != nil {
		router.RespondWithError(r, w, http.StatusInternalServerError, "service initialization failed", err)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		router.RespondWithError(r, w, http.StatusBadRequest, "item ID is required", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		router.RespondWithError(r, w, http.StatusBadRequest, "invalid item ID", err)
		return
	}

	err = service.DeleteItem(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			router.RespondWithError(r, w, http.StatusNotFound, "item not found", err)
			return
		}
		router.RespondWithError(r, w, http.StatusInternalServerError, "failed to delete item", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func serviceFromRequest(db *gorm.DB, r *http.Request) (*Service, error) {
	cfg, err := validator.ExtractAndValidateContext[*config.Conf](r.Context(), contextkeys.CtxKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("invalid config context: %w", err)
	}
	log, err := validator.ExtractAndValidateContext[*logger.Logger](r.Context(), contextkeys.CtxKeyLogger)
	if err != nil {
		return nil, fmt.Errorf("invalid logger context: %w", err)
	}

	return NewService(cfg, db, log), nil
}
