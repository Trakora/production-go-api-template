package item

import (
	"context"
	"production-go-api-template/config"
	"production-go-api-template/pkg/logger"

	"gorm.io/gorm"
)

type Service struct {
	DB   *gorm.DB
	Log  *logger.Logger
	repo ItemRepository
}

func NewService(cfg *config.Conf, db *gorm.DB, log *logger.Logger) *Service {
	return &Service{
		DB:   db,
		Log:  log,
		repo: NewSQLiteItemRepo(db),
	}
}

func (s *Service) CreateItem(ctx context.Context, req CreateItemRequest) (Item, error) {
	log := s.Log.WithRequestID(ctx)

	if err := req.Validate(); err != nil {
		log.Errorf("validation failed for create item: %v", err)
		return Item{}, err
	}

	log.Infof("Creating new item: %s", req.Name)

	item := Item{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
	}

	createdItem, err := s.repo.Create(ctx, item)
	if err != nil {
		log.Errorf("failed to create item: %v", err)
		return Item{}, err
	}

	log.Infof("Successfully created item with ID: %d", createdItem.ID)
	return createdItem, nil
}

func (s *Service) GetItem(ctx context.Context, id int) (Item, error) {
	log := s.Log.WithRequestID(ctx)
	log.Infof("Fetching item with ID: %d", id)

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Errorf("failed to get item with ID %d: %v", id, err)
		return Item{}, err
	}

	log.Infof("Successfully retrieved item: %s", item.Name)
	return item, nil
}

func (s *Service) GetAllItems(ctx context.Context) ([]Item, error) {
	log := s.Log.WithRequestID(ctx)
	log.Infof("Fetching all items")

	items, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Errorf("failed to get all items: %v", err)
		return nil, err
	}

	log.Infof("Successfully retrieved %d items", len(items))
	return items, nil
}

func (s *Service) UpdateItem(ctx context.Context, id int, req UpdateItemRequest) (Item, error) {
	log := s.Log.WithRequestID(ctx)

	if err := req.Validate(); err != nil {
		log.Errorf("validation failed for update item: %v", err)
		return Item{}, err
	}

	log.Infof("Updating item with ID: %d", id)

	item := Item{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
	}

	updatedItem, err := s.repo.Update(ctx, id, item)
	if err != nil {
		log.Errorf("failed to update item with ID %d: %v", id, err)
		return Item{}, err
	}

	log.Infof("Successfully updated item with ID: %d", updatedItem.ID)
	return updatedItem, nil
}

func (s *Service) DeleteItem(ctx context.Context, id int) error {
	log := s.Log.WithRequestID(ctx)
	log.Infof("Deleting item with ID: %d", id)

	err := s.repo.Delete(ctx, id)
	if err != nil {
		log.Errorf("failed to delete item with ID %d: %v", id, err)
		return err
	}
	log.Infof("Successfully deleted item with ID: %d", id)
	return nil
}
