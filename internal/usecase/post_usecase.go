package usecase

import (
	"livoir-blog/internal/domain"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

type postUsecase struct {
	postRepo        domain.PostRepository
	postVersionRepo domain.PostVersionRepository
	sanitizer       *bluemonday.Policy
}

func NewPostUsecase(repo domain.PostRepository, postVersionRepo domain.PostVersionRepository) domain.PostUsecase {
	return &postUsecase{
		postRepo:        repo,
		postVersionRepo: postVersionRepo,
		sanitizer:       bluemonday.UGCPolicy(),
	}
}

func (u *postUsecase) GetByID(id int64) (*domain.PostWithVersion, error) {
	return u.postRepo.GetByID(id)
}

func (u *postUsecase) Create(request *domain.CreatePostDTO) error {
	// Sanitize the post content
	request.Content = u.sanitizer.Sanitize(request.Content)
	post := &domain.Post{
		CreatedAt: time.Now(),
	}
	err := u.postRepo.Create(post)
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
	err = u.postVersionRepo.Create(postVersion)
	if err != nil {
		return err
	}
	post.CurrentVersionID = postVersion.ID
	err = u.postRepo.Update(post)
	if err != nil {
		return err
	}
	return nil
}
