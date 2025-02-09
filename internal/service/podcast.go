package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"shortcast/internal/dto"
	"shortcast/internal/model"
	"shortcast/internal/repository"
	"strings"
	"time"
)

type PodcastService struct {
	podcastRepo *repository.PodcastRepository
	userRepo    *repository.UserRepository
	r2Service   *R2Service
}

func NewPodcastService(podcastRepo *repository.PodcastRepository, userRepo *repository.UserRepository, r2Service *R2Service) *PodcastService {
	return &PodcastService{
		podcastRepo: podcastRepo,
		userRepo:    userRepo,
		r2Service:   r2Service,
	}
}

func (s *PodcastService) UploadPodcast(podcastDTO *dto.UploadPodcastRequest, audioFile, coverFile *multipart.FileHeader) (*dto.PodcastResponse, error) {
	// R2'ye yükle
	audioURL, err := s.r2Service.UploadFile(audioFile, "audio")
	if err != nil {
		return nil, err
	}

	coverURL, err := s.r2Service.UploadFile(coverFile, "covers")
	if err != nil {
		// Hata durumunda audio dosyasını da sil
		s.r2Service.DeleteFile(audioURL)
		return nil, err
	}

	// Podcast modeli oluştur
	podcast := &model.Podcast{
		Title:    podcastDTO.Title,
		Category: podcastDTO.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
		UserID:   podcastDTO.UserID,
	}

	// Veritabanına kaydet
	if err := s.podcastRepo.SavePodcast(podcast); err != nil {
		// Hata durumunda yüklenen dosyaları sil
		s.r2Service.DeleteFile(audioURL)
		s.r2Service.DeleteFile(coverURL)
		return nil, err
	}

	return &dto.PodcastResponse{
		ID:       podcast.ID,
		Title:    podcast.Title,
		Category: podcast.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
	}, nil
}

// Yardımcı fonksiyon
func saveFile(file *multipart.FileHeader, path string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func (s *PodcastService) GetPodcastByID(id uint) (*dto.PodcastResponse, error) {
	podcast, err := s.podcastRepo.GetPodcastByID(id)
	if err != nil {
		return nil, err
	}

	// AudioURL başına / ekle (eğer yoksa)
	audioURL := podcast.AudioURL
	if !strings.HasPrefix(audioURL, "/") {
		audioURL = "/" + audioURL
	}

	// CoverURL başına / ekle (eğer yoksa)
	coverURL := podcast.CoverURL
	if !strings.HasPrefix(coverURL, "/") {
		coverURL = "/" + coverURL
	}

	return &dto.PodcastResponse{
		ID:       podcast.ID,
		Title:    podcast.Title,
		Category: podcast.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
	}, nil
}

func (s *PodcastService) GetUserPodcasts(userID uint) ([]dto.PodcastResponse, error) {
	podcasts, err := s.podcastRepo.GetPodcastsByUserID(userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.PodcastResponse, 0)
	for _, podcast := range *podcasts {
		// AudioURL başına / ekle (eğer yoksa)
		audioURL := podcast.AudioURL
		if !strings.HasPrefix(audioURL, "/") {
			audioURL = "/" + audioURL
		}

		response = append(response, dto.PodcastResponse{
			ID:       podcast.ID,
			Title:    podcast.Title,
			Category: podcast.Category,
			AudioURL: audioURL,
			CoverURL: podcast.CoverURL,
		})
	}
	return response, nil
}

func (s *PodcastService) DiscoverPodcasts(req *dto.PodcastDiscoverRequest) (*dto.PodcastCursor, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10 // Varsayılan limit
	}

	podcasts, err := s.podcastRepo.DiscoverPodcasts(req.Cursor, req.Direction, limit)
	if err != nil {
		return nil, err
	}

	var response dto.PodcastCursor
	response.Podcasts = make([]dto.PodcastResponse, 0)

	hasMore := len(*podcasts) > limit
	actualPodcasts := *podcasts
	if hasMore {
		actualPodcasts = actualPodcasts[:limit]
		// Eğer daha fazla podcast varsa, next_cursor bir sonraki podcast'in ID'si olmalı
		nextID := actualPodcasts[len(actualPodcasts)-1].ID + 1
		response.NextCursor = &nextID
	}

	for _, podcast := range actualPodcasts {
		response.Podcasts = append(response.Podcasts, dto.PodcastResponse{
			ID:       podcast.ID,
			Title:    podcast.Title,
			Category: podcast.Category,
			AudioURL: podcast.AudioURL,
			CoverURL: podcast.CoverURL,
		})
	}

	response.HasNext = hasMore
	response.HasPrevious = req.Cursor != nil

	return &response, nil
}

func (s *PodcastService) UpdatePodcast(id uint, userID uint, req *dto.UpdatePodcastRequest) (*dto.PodcastResponse, error) {
	existingPodcast, err := s.podcastRepo.GetPodcastByID(id)
	if err != nil {
		return nil, err
	}

	if existingPodcast.UserID != userID {
		return nil, errors.New("bu podcast'i düzenleme yetkiniz yok")
	}

	existingPodcast.Title = req.Title
	existingPodcast.Category = req.Category

	err = s.podcastRepo.UpdatePodcast(id, existingPodcast)
	if err != nil {
		return nil, err
	}

	// AudioURL başına / ekle (eğer yoksa)
	audioURL := existingPodcast.AudioURL
	if !strings.HasPrefix(audioURL, "/") {
		audioURL = "/" + audioURL
	}

	// CoverURL başına / ekle (eğer yoksa)
	coverURL := existingPodcast.CoverURL
	if !strings.HasPrefix(coverURL, "/") {
		coverURL = "/" + coverURL
	}

	return &dto.PodcastResponse{
		ID:       existingPodcast.ID,
		Title:    existingPodcast.Title,
		Category: existingPodcast.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
	}, nil
}

func (s *PodcastService) DeletePodcast(id uint, userID uint) error {
	podcast, err := s.podcastRepo.GetPodcastByID(id)
	if err != nil {
		return err
	}

	if podcast.UserID != userID {
		return errors.New("bu podcast'i silme yetkiniz yok")
	}

	// Dosyaları sil
	currentDir, _ := os.Getwd()

	// HLS dizinini sil
	if podcast.AudioURL != "" {
		audioPath := filepath.Join(currentDir, strings.TrimPrefix(podcast.AudioURL, "/"))
		podcastDir := filepath.Dir(audioPath)
		os.RemoveAll(podcastDir)
	}

	// Kapak fotoğrafını sil
	if podcast.CoverURL != "" {
		coverPath := filepath.Join(currentDir, strings.TrimPrefix(podcast.CoverURL, "/"))
		os.Remove(coverPath)
	}

	return s.podcastRepo.DeletePodcast(id)
}

func (s *PodcastService) LikePodcast(podcastID, userID uint) (*dto.LikeResponse, error) {
	// Kullanıcı kontrolü
	if _, err := s.userRepo.GetUserByID(userID); err != nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}

	// Podcast kontrolü
	if _, err := s.podcastRepo.GetPodcastByID(podcastID); err != nil {
		return nil, err
	}

	liked, err := s.podcastRepo.LikePodcast(podcastID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.LikeResponse{
		PodcastID: podcastID,
		UserID:    userID,
		Liked:     liked,
	}, nil
}

func (s *PodcastService) GetLikedPodcasts(userID uint) ([]dto.PodcastResponse, error) {
	if _, err := s.userRepo.GetUserByID(userID); err != nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}

	podcasts, err := s.podcastRepo.GetLikedPodcasts(userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.PodcastResponse, 0) // Boş array ile başla
	for _, podcast := range *podcasts {
		response = append(response, dto.PodcastResponse{
			ID:       podcast.ID,
			Title:    podcast.Title,
			Category: podcast.Category,
			AudioURL: podcast.AudioURL,
			CoverURL: podcast.CoverURL,
		})
	}
	return response, nil
}

func (s *PodcastService) GetPodcastsByCategory(category string) ([]dto.PodcastResponse, error) {
	podcasts, err := s.podcastRepo.GetPodcastsByCategory(category)
	if err != nil {
		return nil, err
	}

	response := make([]dto.PodcastResponse, 0) // Boş array ile başla
	for _, p := range *podcasts {
		response = append(response, dto.PodcastResponse{
			ID:       p.ID,
			Title:    p.Title,
			Category: p.Category,
			AudioURL: p.AudioURL,
			CoverURL: p.CoverURL,
		})
	}
	return response, nil
}

func (s *PodcastService) AddComment(podcastID, userID uint, content string) (*dto.CommentResponse, error) {
	// Kullanıcı kontrolü
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}

	// Podcast kontrolü
	if _, err := s.podcastRepo.GetPodcastByID(podcastID); err != nil {
		return nil, errors.New("podcast bulunamadı")
	}

	comment := &model.Comment{
		PodcastID: podcastID,
		UserID:    userID,
		Content:   content,
	}

	err = s.podcastRepo.AddComment(comment)
	if err != nil {
		return nil, err
	}

	return &dto.CommentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		UserID:    comment.UserID,
		Username:  user.Username,
		CreatedAt: comment.CreatedAt,
	}, nil
}

func (s *PodcastService) GetComments(podcastID uint) ([]dto.CommentResponse, error) {
	comments, err := s.podcastRepo.GetComments(podcastID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.CommentResponse, 0) // Boş array ile başla
	for _, comment := range *comments {
		response = append(response, dto.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			UserID:    comment.UserID,
			Username:  comment.User.Username,
			CreatedAt: comment.CreatedAt,
		})
	}
	return response, nil
}

func (s *PodcastService) UpdatePodcastCover(id uint, userID uint, coverFile *multipart.FileHeader) (*dto.PodcastResponse, error) {
	// Önce podcast'i getir
	existingPodcast, err := s.podcastRepo.GetPodcastByID(id)
	if err != nil {
		return nil, err
	}

	// Podcast'in sahibi olup olmadığını kontrol et
	if existingPodcast.UserID != userID {
		return nil, errors.New("bu podcast'i düzenleme yetkiniz yok")
	}

	// Eski kapak fotoğrafını sil (eğer varsa)
	if existingPodcast.CoverURL != "" {
		oldCoverPath := filepath.Join(".", existingPodcast.CoverURL)
		os.Remove(oldCoverPath)
	}

	// Yeni kapak fotoğrafını kaydet
	currentDir, _ := os.Getwd()
	coverDir := filepath.Join(currentDir, "uploads", "covers")

	if err := os.MkdirAll(coverDir, os.ModePerm); err != nil {
		return nil, err
	}

	coverFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), coverFile.Filename)
	coverPath := filepath.Join(coverDir, coverFileName)
	if err := saveFile(coverFile, coverPath); err != nil {
		return nil, err
	}

	// URL'i güncelle - zaten / ile başlıyor
	coverURL := fmt.Sprintf("/uploads/covers/%s", coverFileName)
	existingPodcast.CoverURL = coverURL

	// Veritabanını güncelle
	err = s.podcastRepo.UpdatePodcast(id, existingPodcast)
	if err != nil {
		// Hata durumunda yüklenen dosyayı sil
		os.Remove(coverPath)
		return nil, err
	}

	// Güncellenmiş podcast'i döndür
	// AudioURL başına / ekle (eğer yoksa)
	audioURL := existingPodcast.AudioURL
	if !strings.HasPrefix(audioURL, "/") {
		audioURL = "/" + audioURL
	}

	return &dto.PodcastResponse{
		ID:       existingPodcast.ID,
		Title:    existingPodcast.Title,
		Category: existingPodcast.Category,
		AudioURL: audioURL,
		CoverURL: existingPodcast.CoverURL,
	}, nil
}
