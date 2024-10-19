package domain

type Post struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PostRepository interface {
	GetByID(id int64) (*Post, error)
}

type PostUsecase interface {
	GetByID(id int64) (*Post, error)
}
