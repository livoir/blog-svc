package usecase

import (
	"livoir-blog/internal/domain"
	"time"
)

type postUsecase struct {
	postRepo domain.PostRepository
}

func NewPostUsecase(repo domain.PostRepository) domain.PostUsecase {
	return &postUsecase{repo}
}

func (u *postUsecase) GetByID(id int64) (*domain.Post, error) {
	return u.postRepo.GetByID(id)
}

func (u *postUsecase) Create(post *domain.Post) error {
	post.CreatedAt = time.Now()
	return u.postRepo.Create(post)
}
