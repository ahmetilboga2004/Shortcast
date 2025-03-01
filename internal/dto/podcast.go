package dto

import (
	"time"
)

type UploadPodcastRequest struct {
	UserID   uint
	Title    string `form:"title" validate:"required"`
	Category string `form:"category" validate:"required"`
}

type PodcastResponse struct {
	ID       uint    `json:"id"`
	Title    string  `json:"title"`
	Category string  `json:"category"`
	AudioURL string  `json:"audio_url"`
	CoverURL string  `json:"cover_url"`
	User     UserDTO `json:"user"`
}

type PodcastCursor struct {
	NextCursor  *uint             `json:"next_cursor,omitempty"`
	PrevCursor  *uint             `json:"prev_cursor,omitempty"`
	Podcasts    []PodcastResponse `json:"podcasts"`
	HasNext     bool              `json:"has_next"`
	HasPrevious bool              `json:"has_previous"`
}

type PodcastDiscoverRequest struct {
	Cursor    *uint  `query:"cursor"`    // İsteğe bağlı cursor
	Direction string `query:"direction"` // "next" veya "prev"
	Limit     int    `query:"limit"`     // Sayfa başına podcast sayısı
}

type UpdatePodcastRequest struct {
	Title    string `json:"title" validate:"required"`
	Category string `json:"category" validate:"required"`
}

type CommentRequest struct {
	Content string `json:"content" validate:"required"`
}

type CommentResponse struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type LikeResponse struct {
	PodcastID uint `json:"podcast_id"`
	UserID    uint `json:"user_id"`
	Liked     bool `json:"liked"`
}

// UpdatePodcastCoverRequest, kapak fotoğrafı güncellemek için
type UpdatePodcastCoverRequest struct {
	CoverURL string `json:"cover_url"`
}
