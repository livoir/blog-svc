package usecase

import (
	"context"
	"errors"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"net/http"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

type postUsecase struct {
	postRepo        domain.PostRepository
	postVersionRepo domain.PostVersionRepository
	transactor      domain.Transactor
	sanitizer       *bluemonday.Policy
}

func NewPostUsecase(repo domain.PostRepository, postVersionRepo domain.PostVersionRepository, transactor domain.Transactor) (domain.PostUsecase, error) {
	if repo == nil || postVersionRepo == nil || transactor == nil {
		return nil, errors.New("nil repository or transactor")
	}
	return &postUsecase{
		postRepo:        repo,
		postVersionRepo: postVersionRepo,
		transactor:      transactor,
		sanitizer:       bluemonday.UGCPolicy(),
	}, nil
}

func (u *postUsecase) GetByID(ctx context.Context, id string) (*domain.PostWithVersion, error) {
	post, err := u.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, common.ErrPostNotFound
	}
	return post, nil
}

func (u *postUsecase) Create(ctx context.Context, request *domain.CreatePostDTO) (*domain.PostResponseDTO, error) {
	// Sanitize the post content
	request.Content = u.sanitizer.Sanitize(request.Content)
	request.Title = u.sanitizer.Sanitize(request.Title)
	now := time.Now()
	post := &domain.Post{
		CreatedAt: now,
		UpdatedAt: now,
	}
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func(tx domain.Transaction) {
		if p := recover(); p != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "panic_recovery"))
			}
			panic(p)
		} else if err != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "error_handling"))
			}
		}
	}(tx)
	err = u.postRepo.Create(ctx, tx, post)
	if err != nil {
		return nil, err
	}
	postVersion := &domain.PostVersion{
		VersionNumber: 1,
		PostID:        post.ID,
		CreatedAt:     time.Now(),
		Title:         request.Title,
		Content:       request.Content,
	}
	err = u.postVersionRepo.Create(ctx, tx, postVersion)
	if err != nil {
		return nil, err
	}
	err = u.postRepo.Update(ctx, tx, post)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &domain.PostResponseDTO{
		PostID:  post.ID,
		Title:   request.Title,
		Content: request.Content,
	}, nil
}

func (u *postUsecase) Update(ctx context.Context, id string, request *domain.UpdatePostDTO) (*domain.PostResponseDTO, error) {
	// Sanitize the post content
	request.Content = u.sanitizer.Sanitize(request.Content)
	request.Title = u.sanitizer.Sanitize(request.Title)
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func(tx domain.Transaction) {
		if p := recover(); p != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "panic_recovery"))
			}
			panic(p)
		} else if err != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "error_handling"))
			}
		}
	}(tx)
	// Check if the post exists
	post, err := u.postRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, common.ErrPostNotFound
	}
	// Get latest post version
	postVersion, err := u.postVersionRepo.GetLatestByPostIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if postVersion == nil {
		return nil, common.ErrPostVersionNotFound
	}
	if postVersion.PublishedAt == nil {
		postVersion.Title = request.Title
		postVersion.Content = request.Content
		err = u.postVersionRepo.Update(ctx, tx, postVersion)
		if err != nil {
			return nil, err
		}
	} else {
		newPostVersion := &domain.PostVersion{
			VersionNumber: postVersion.VersionNumber + 1,
			PostID:        id,
			CreatedAt:     time.Now(),
			Title:         request.Title,
			Content:       request.Content,
		}
		err = u.postVersionRepo.Create(ctx, tx, newPostVersion)
		if err != nil {
			return nil, err
		}
		post.CurrentVersionID = newPostVersion.ID
		err = u.postRepo.Update(ctx, tx, post)
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &domain.PostResponseDTO{
		PostID:  post.ID,
		Title:   postVersion.Title,
		Content: postVersion.Content,
	}, nil
}

func (u *postUsecase) Publish(ctx context.Context, id string) (*domain.PublishResponseDTO, error) {
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func(tx domain.Transaction) {
		if p := recover(); p != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "panic_recovery"))
			}
			panic(p)
		} else if err != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "error_handling"))
			}
		}
	}(tx)
	postVersion, err := u.postVersionRepo.GetLatestByPostIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if postVersion == nil {
		return nil, common.ErrPostVersionNotFound
	}
	post, err := u.postRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, common.ErrPostNotFound
	}
	if postVersion.PublishedAt != nil {
		return nil, common.NewCustomError(http.StatusForbidden, "post already published")
	}
	now := time.Now()
	postVersion.PublishedAt = &now
	err = u.postVersionRepo.Update(ctx, tx, postVersion)
	if err != nil {
		return nil, err
	}
	post.UpdatedAt = now
	post.CurrentVersionID = postVersion.ID
	err = u.postRepo.Update(ctx, tx, post)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &domain.PublishResponseDTO{
		PostID:      postVersion.PostID,
		PublishedAt: postVersion.PublishedAt,
		Title:       postVersion.Title,
		Content:     postVersion.Content,
	}, nil
}

func (u *postUsecase) DeletePostVersionByPostID(ctx context.Context, id string) error {
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return err
	}
	defer func(tx domain.Transaction) {
		if p := recover(); p != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "panic_recovery"))
			}
			panic(p)
		} else if err != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction",
					zap.Error(e),
					zap.String("error_source", "error_handling"))
			}
		}
	}(tx)
	postVersion, err := u.postVersionRepo.GetLatestByPostIDForUpdate(ctx, tx, id)
	if err != nil {
		return err
	}
	if postVersion == nil {
		logger.Log.Error("Post version not found", zap.String("postID", id))
		return common.ErrPostVersionNotFound
	}
	if postVersion.PublishedAt != nil {
		logger.Log.Error("Post version is published", zap.String("postID", id))
		return common.NewCustomError(http.StatusConflict, "post version is published")
	}
	err = u.postVersionRepo.Delete(ctx, tx, postVersion.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
