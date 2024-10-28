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
	transactor      domain.Transactor
	categoryRepo    domain.CategoryRepository
	postVersionRepo domain.PostVersionRepository
}

func NewCategoryUsecase(transactor domain.Transactor, categoryRepo domain.CategoryRepository, postVersionRepo domain.PostVersionRepository) (domain.CategoryUsecase, error) {
	if transactor == nil || categoryRepo == nil || postVersionRepo == nil {
		return nil, errors.New("nil transactor or category repository")
	}
	return &CategoryUsecase{
		transactor:      transactor,
		categoryRepo:    categoryRepo,
		postVersionRepo: postVersionRepo,
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
		err = common.ErrCategoryNameDuplicate
		return nil, err
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
		} else if err != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction", zap.Error(e), zap.String("error_source", "error_propagation"))
			}
		}
	}(tx)
	existingCategory, err := u.categoryRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if existingCategory == nil {
		err = common.ErrCategoryNotFound
		return nil, err
	}
	if existingCategory.Name == request.Name {
		err = common.NewCustomError(http.StatusBadRequest, "name is the same as before")
		return nil, err
	}
	otherCategory, err := u.categoryRepo.GetByName(ctx, request.Name)
	if err != nil && !errors.Is(err, common.ErrCategoryNotFound) {
		return nil, err
	}
	if otherCategory != nil {
		err = common.ErrCategoryNameDuplicate
		return nil, err
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

func (u *CategoryUsecase) AttachToPostVersion(ctx context.Context, request *domain.AttachCategoryToPostVersionRequestDTO) error {
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return err
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

	postVersion, err := u.postVersionRepo.GetByID(ctx, request.PostVersionID)
	if err != nil {
		return err
	}
	if postVersion == nil {
		err = common.ErrPostVersionNotFound
		return err
	}
	category, err := u.categoryRepo.GetByID(ctx, request.CategoryIDs[0])
	if err != nil {
		return err
	}
	if category == nil {
		err = common.ErrCategoryNotFound
		return err
	}
	var postVersionCategories []domain.PostVersionCategory
	for _, categoryID := range request.CategoryIDs {
		postVersionCategories = append(postVersionCategories, domain.PostVersionCategory{
			PostVersionID: request.PostVersionID,
			CategoryID:    categoryID,
		})
	}
	err = u.categoryRepo.AttachToPostVersion(ctx, tx, postVersionCategories)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
