package domain

import "time"

type Admin struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	FullName  string     `json:"full_name"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeleteAt  *time.Time `json:"delete_at"`
}
