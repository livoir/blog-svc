package usecase

import (
	"livoir-blog/internal/domain"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

type postUsecase struct {
	postRepo  domain.PostRepository
	sanitizer *bluemonday.Policy
}

func NewPostUsecase(repo domain.PostRepository) domain.PostUsecase {
	return &postUsecase{
		postRepo:  repo,
		sanitizer: bluemonday.UGCPolicy(),
	}
}

func (u *postUsecase) GetByID(id int64) (*domain.Post, error) {
	return u.postRepo.GetByID(id)
}

func (u *postUsecase) Create(post *domain.Post) error {
	// Sanitize the post content
	post.Content = u.sanitizer.Sanitize(post.Content)

	post.CreatedAt = time.Now()
	return u.postRepo.Create(post)
}
