package usecase

import (
	"errors"
	"livoir-blog/internal/domain"
	"time"

	"github.com/microcosm-cc/bluemonday"
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

func (u *postUsecase) GetByID(id string) (*domain.PostWithVersion, error) {
	return u.postRepo.GetByID(id)
}

func (u *postUsecase) Create(request *domain.CreatePostDTO) error {
	// Sanitize the post content
	request.Content = u.sanitizer.Sanitize(request.Content)
	request.Title = u.sanitizer.Sanitize(request.Title)
	post := &domain.Post{
		CreatedAt: time.Now(),
	}
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = u.postRepo.Create(tx, post)
	if err != nil {
		return err
	}
	request.PostID = post.ID
	postVersion := &domain.PostVersion{
		VersionNumber: 1,
		PostID:        post.ID,
		CreatedAt:     time.Now(),
		Title:         request.Title,
		Content:       request.Content,
	}
	err = u.postVersionRepo.Create(tx, postVersion)
	if err != nil {
		return err
	}
	post.CurrentVersionID = postVersion.ID
	err = u.postRepo.Update(tx, post)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (u *postUsecase) Update(id string, request *domain.UpdatePostDTO) error {
	// Sanitize the post content
	request.Content = u.sanitizer.Sanitize(request.Content)
	request.Title = u.sanitizer.Sanitize(request.Title)
	tx, err := u.transactor.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Check if the post exists
	post, err := u.postRepo.GetByIDForUpdate(tx, id)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}
	// Get latest post version
	postVersion, err := u.postVersionRepo.GetLatestByPostIDForUpdate(tx, id)
	if err != nil {
		return err
	}
	if postVersion == nil {
		return errors.New("post version not found")
	}
	if postVersion.PublishedAt == nil {
		postVersion.Title = request.Title
		postVersion.Content = request.Content
		err = u.postVersionRepo.Update(tx, postVersion)
		if err != nil {
			return err
		}
	} else {
		newPostVersion := &domain.PostVersion{
			VersionNumber: postVersion.VersionNumber + 1,
			PostID:        id,
			CreatedAt:     time.Now(),
			Title:         request.Title,
			Content:       request.Content,
		}
		err = u.postVersionRepo.Create(tx, newPostVersion)
		if err != nil {
			return err
		}
		post.CurrentVersionID = newPostVersion.ID
		err = u.postRepo.Update(tx, post)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
