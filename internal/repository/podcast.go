package repository

import (
	"errors"
	"fmt"
	"shortcast/internal/model"

	"gorm.io/gorm"
)

type PodcastRepository struct {
	db *gorm.DB
}

func NewPodcastRepository(db *gorm.DB) *PodcastRepository {
	return &PodcastRepository{db: db}
}

func (r *PodcastRepository) SavePodcast(podcast *model.Podcast) error {
	return r.db.Create(podcast).Error
}

func (r *PodcastRepository) GetPodcastByID(id uint) (*model.Podcast, error) {
	var podcast model.Podcast
	if err := r.db.Preload("User").First(&podcast, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("podcast bulunamadı")
		}
		return nil, err
	}

	return &podcast, nil
}

func (r *PodcastRepository) GetPodcastsByUserID(userId uint) (*[]model.Podcast, error) {
	var podcasts []model.Podcast
	err := r.db.Preload("User").Where("user_id = ?", userId).Find(&podcasts).Error
	if err != nil {
		return nil, fmt.Errorf("veritabanından podcastler alınırken hata: %v", err)
	}

	if len(podcasts) == 0 {
		return &[]model.Podcast{}, nil
	}

	return &podcasts, nil
}

func (r *PodcastRepository) DiscoverPodcasts(cursor *uint, direction string, limit int) (*[]model.Podcast, error) {
	var podcasts []model.Podcast
	query := r.db.Model(&model.Podcast{}).Preload("User")

	if cursor != nil {
		if direction == "next" {
			query = query.Where("id > ?", *cursor)
		} else {
			query = query.Where("id < ?", *cursor)
		}
	}

	if direction == "prev" {
		query = query.Order("id DESC")
	} else {
		query = query.Order("id ASC")
	}

	err := query.Limit(limit + 1).Find(&podcasts).Error // Bir fazla alıyoruz ki sonraki sayfa var mı bilelim
	if err != nil {
		return nil, err
	}

	return &podcasts, nil
}

func (r *PodcastRepository) UpdatePodcast(id uint, podcast *model.Podcast) error {
	result := r.db.Model(&model.Podcast{}).Where("id = ?", id).Updates(podcast)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("podcast bulunamadı")
	}
	return nil
}

func (r *PodcastRepository) DeletePodcast(id uint) error {
	result := r.db.Delete(&model.Podcast{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("podcast bulunamadı")
	}
	return nil
}

func (r *PodcastRepository) LikePodcast(podcastID, userID uint) (bool, error) {
	var like model.Like

	// Önce mevcut like'ı kontrol et
	result := r.db.Where("podcast_id = ? AND user_id = ?", podcastID, userID).First(&like)

	if result.Error == nil {
		// Like zaten var, unlike yap (soft delete)
		if err := r.db.Delete(&like).Error; err != nil {
			return false, err
		}
		return false, nil // false = unliked
	}

	// Like yok, yeni like ekle
	newLike := model.Like{
		PodcastID: podcastID,
		UserID:    userID,
	}

	if err := r.db.Create(&newLike).Error; err != nil {
		return false, err
	}

	return true, nil // true = liked
}

func (r *PodcastRepository) GetLikedPodcasts(userID uint) (*[]model.Podcast, error) {
	var podcasts []model.Podcast
	err := r.db.Preload("User").
		Joins("JOIN likes ON likes.podcast_id = podcasts.id").
		Where("likes.user_id = ? AND likes.deleted_at IS NULL", userID).
		Find(&podcasts).Error
	return &podcasts, err
}

func (r *PodcastRepository) GetPodcastsByCategory(category string) (*[]model.Podcast, error) {
	var podcasts []model.Podcast
	err := r.db.Preload("User").Where("category = ?", category).Find(&podcasts).Error
	return &podcasts, err
}

func (r *PodcastRepository) AddComment(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *PodcastRepository) GetComments(podcastID uint) (*[]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Joins("User").Where("podcast_id = ?", podcastID).
		Order("created_at desc").Find(&comments).Error
	return &comments, err
}
