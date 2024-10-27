package usecase

import (
	"context"
	"errors"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"net/http"
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

func (u *CategoryUsecase) Create(ctx context.Context, request *domain.CategoryRequestDTO) (*domain.CategoryResponseDTO, error) {
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
	existingCategory, err := u.categoryRepo.GetByName(ctx, request.Name)
	if err != nil && !errors.Is(err, common.ErrCategoryNotFound) {
		return nil, err
	}
	if existingCategory != nil {
		return nil, common.ErrCategoryNameDuplicate
	}
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
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}, nil
}

func (u *CategoryUsecase) Update(ctx context.Context, id string, request *domain.CategoryRequestDTO) (*domain.CategoryResponseDTO, error) {
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
		}
		e := tx.Rollback()
		if e != nil {
			logger.Log.Error("Failed to rollback transaction", zap.Error(e), zap.String("error_source", "error_propagation"))
		}
	}(tx)
	existingCategory, err := u.categoryRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if existingCategory == nil {
		return nil, common.NewCustomError(http.StatusNotFound, "category not found")
	}
	if existingCategory.Name == request.Name {
		return nil, common.NewCustomError(http.StatusBadRequest, "name is the same as before")
	}
	otherCategory, err := u.categoryRepo.GetByName(ctx, request.Name)
	if err != nil && !errors.Is(err, common.ErrCategoryNotFound) {
		return nil, err
	}
	if otherCategory != nil {
		return nil, common.ErrCategoryNameDuplicate
	}
	now := time.Now()
	category := &domain.Category{
		ID:        id,
		Name:      request.Name,
		UpdatedAt: now,
	}
	err = u.categoryRepo.Update(ctx, tx, category)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &domain.CategoryResponseDTO{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: existingCategory.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}, nil
}
