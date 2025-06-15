package item

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ItemRepository interface {
	Create(ctx context.Context, item Item) (Item, error)
	GetByID(ctx context.Context, id int) (Item, error)
	GetAll(ctx context.Context) ([]Item, error)
	Update(ctx context.Context, id int, item Item) (Item, error)
	Delete(ctx context.Context, id int) error
}

type sqliteItemRepo struct {
	db *gorm.DB
}

func NewSQLiteItemRepo(db *gorm.DB) ItemRepository {
	return &sqliteItemRepo{db: db}
}

func (r *sqliteItemRepo) Create(ctx context.Context, item Item) (Item, error) {
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt

	if err := r.db.WithContext(ctx).Create(&item).Error; err != nil {
		return Item{}, err
	}

	return item, nil
}

func (r *sqliteItemRepo) GetByID(ctx context.Context, id int) (Item, error) {
	var item Item
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Item{}, errors.New("item not found")
		}
		return Item{}, err
	}

	return item, nil
}

func (r *sqliteItemRepo) GetAll(ctx context.Context) ([]Item, error) {
	var items []Item
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *sqliteItemRepo) Update(ctx context.Context, id int, updatedItem Item) (Item, error) {
	var existingItem Item
	if err := r.db.WithContext(ctx).First(&existingItem, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Item{}, errors.New("item not found")
		}
		return Item{}, err
	}

	updatedItem.ID = existingItem.ID
	updatedItem.CreatedAt = existingItem.CreatedAt
	updatedItem.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Save(&updatedItem).Error; err != nil {
		return Item{}, err
	}

	return updatedItem, nil
}

func (r *sqliteItemRepo) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&Item{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}
