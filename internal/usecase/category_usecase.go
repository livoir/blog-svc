package usecase

import (
	"context"
	"errors"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/logger"
	"time"

	"go.uber.org/zap"
)

type CategoryUsecase struct {
	transactor   domain.Transactor
	categoryRepo domain.CategoryRepository
}

func NewCategoryUsecase(transactor domain.Transactor, categoryRepo domain.CategoryRepository) (domain.CategoryUsecase, error) {
	if transactor == nil || categoryRepo == nil {
		return nil, errors.New("nil transactor or category repository")
	}
	return &CategoryUsecase{
		transactor:   transactor,
		categoryRepo: categoryRepo,
	}, nil
}

func (u *CategoryUsecase) Create(ctx context.Context, request *domain.CreateCategoryDTO) (*domain.CategoryResponseDTO, error) {
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func(tx domain.Transaction) {
		if p := recover(); p != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction", zap.Error(e), zap.String("error_source", "panic_recovery"))
			}
			panic(p)
		} else if err != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction", zap.Error(e), zap.String("error_source", "error_propagation"))
			}
		}
	}(tx)
	now := time.Now()
	category := &domain.Category{
		Name:      request.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = u.categoryRepo.Create(ctx, tx, category)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &domain.CategoryResponseDTO{
		ID:   category.ID,
		Name: category.Name,
	}, nil
}
