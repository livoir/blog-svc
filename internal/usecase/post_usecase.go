package usecase

import (
	"fmt"
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

func NewPostUsecase(repo domain.PostRepository, postVersionRepo domain.PostVersionRepository, transactor domain.Transactor) domain.PostUsecase {
	return &postUsecase{
		postRepo:        repo,
		postVersionRepo: postVersionRepo,
		transactor:      transactor,
		sanitizer:       bluemonday.UGCPolicy(),
	}
}

func (u *postUsecase) GetByID(id string) (*domain.PostWithVersion, error) {
	return u.postRepo.GetByID(id)
}

func (u *postUsecase) Create(request *domain.CreatePostDTO) error {
	// Sanitize the post content
	request.Content = u.sanitizer.Sanitize(request.Content)
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
	request.PostId = post.ID
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
	fmt.Println(post.CurrentVersionID)
	err = u.postRepo.Update(tx, post)
	if err != nil {
		return err
	}
	return tx.Commit()
}
