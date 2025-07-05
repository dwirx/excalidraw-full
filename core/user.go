package core

import "time"

type (
	User struct {
		ID        uint      `json:"id" gorm:"primarykey"`
		GitHubID  int64     `json:"githubId" gorm:"unique"`
		Login     string    `json:"login"`
		AvatarURL string    `json:"avatarUrl"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
)
