package item

import (
	"errors"
	"strings"
	"time"
)

type Item struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"size:1000"`
	Price       float64   `json:"price" gorm:"not null;check:price >= 0"`
	Category    string    `json:"category" gorm:"not null;size:100"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

type UpdateItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

type ItemResponse struct {
	Item
}

type ItemsResponse struct {
	Items []Item `json:"items"`
	Total int    `json:"total"`
}

func (r *CreateItemRequest) Validate() error {
	var errs []string

	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, "name is required")
	}

	if r.Price < 0 {
		errs = append(errs, "price must be non-negative")
	}

	if strings.TrimSpace(r.Category) == "" {
		errs = append(errs, "category is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func (r *UpdateItemRequest) Validate() error {
	var errs []string

	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, "name is required")
	}

	if r.Price < 0 {
		errs = append(errs, "price must be non-negative")
	}

	if strings.TrimSpace(r.Category) == "" {
		errs = append(errs, "category is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
