package usecase

import "livoir-blog/internal/domain"

type postUsecase struct {
	postRepo domain.PostRepository
}

func NewPostUsecase(repo domain.PostRepository) domain.PostUsecase {
	return &postUsecase{repo}
}

func (u *postUsecase) GetByID(id int64) (*domain.Post, error) {
	return u.postRepo.GetByID(id)
}
