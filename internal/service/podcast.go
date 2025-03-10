package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"shortcast/internal/config"
	"shortcast/internal/dto"
	"shortcast/internal/model"
	"shortcast/internal/repository"
	"shortcast/internal/utils"
)

type PodcastService struct {
	podcastRepo *repository.PodcastRepository
	userRepo    *repository.UserRepository
	R2Service   *R2Service
	config      *config.Config
}

func NewPodcastService(podcastRepo *repository.PodcastRepository, userRepo *repository.UserRepository, r2Service *R2Service, cfg *config.Config) *PodcastService {
	return &PodcastService{
		podcastRepo: podcastRepo,
		userRepo:    userRepo,
		R2Service:   r2Service,
		config:      cfg,
	}
}

func (s *PodcastService) UploadPodcast(podcastDTO *dto.UploadPodcastRequest, audioFile, coverFile *multipart.FileHeader) (*dto.PodcastResponse, error) {
	// Kullanıcı bilgilerini al
	user, err := s.userRepo.GetUserByID(podcastDTO.UserID)
	if err != nil {
		return nil, fmt.Errorf("kullanıcı bulunamadı: %v", err)
	}

	// R2'ye yükle
	audioKey, err := s.R2Service.UploadFile(audioFile, "audio")
	if err != nil {
		return nil, err
	}

	coverKey, err := s.R2Service.UploadFile(coverFile, "covers")
	if err != nil {
		// Hata durumunda audio dosyasını da sil
		s.R2Service.DeleteFile(audioKey)
		return nil, err
	}

	// Podcast modeli oluştur
	podcast := &model.Podcast{
		Title:    podcastDTO.Title,
		Category: podcastDTO.Category,
		AudioKey: audioKey,
		CoverKey: coverKey,
		UserID:   podcastDTO.UserID,
	}

	// Veritabanına kaydet
	if err := s.podcastRepo.SavePodcast(podcast); err != nil {
		// Hata durumunda yüklenen dosyaları sil
		s.R2Service.DeleteFile(audioKey)
		s.R2Service.DeleteFile(coverKey)
		return nil, err
	}

	// İmzalı URL'leri oluştur
	audioURL, err := utils.GenerateSignedURL(audioKey, s.config)
	if err != nil {
		return nil, err
	}

	coverURL, err := utils.GenerateSignedURL(coverKey, s.config)
	if err != nil {
		return nil, err
	}

	return &dto.PodcastResponse{
		ID:       podcast.ID,
		Title:    podcast.Title,
		Category: podcast.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
		User: dto.UserDTO{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.Username,
		},
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

	// İmzalı URL'leri oluştur
	audioURL, err := utils.GenerateSignedURL(podcast.AudioKey, s.config)
	if err != nil {
		return nil, err
	}

	coverURL, err := utils.GenerateSignedURL(podcast.CoverKey, s.config)
	if err != nil {
		return nil, err
	}

	return &dto.PodcastResponse{
		ID:       podcast.ID,
		Title:    podcast.Title,
		Category: podcast.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
		User: dto.UserDTO{
			ID:        podcast.User.ID,
			FirstName: podcast.User.FirstName,
			LastName:  podcast.User.LastName,
			Username:  podcast.User.Username,
		},
	}, nil
}

func (s *PodcastService) GetUserPodcasts(userID uint) ([]dto.PodcastResponse, error) {
	podcasts, err := s.podcastRepo.GetPodcastsByUserID(userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.PodcastResponse, 0)
	for _, podcast := range *podcasts {
		// İmzalı URL'leri oluştur
		audioURL, err := utils.GenerateSignedURL(podcast.AudioKey, s.config)
		if err != nil {
			return nil, err
		}

		coverURL, err := utils.GenerateSignedURL(podcast.CoverKey, s.config)
		if err != nil {
			return nil, err
		}

		response = append(response, dto.PodcastResponse{
			ID:       podcast.ID,
			Title:    podcast.Title,
			Category: podcast.Category,
			AudioURL: audioURL,
			CoverURL: coverURL,
			User: dto.UserDTO{
				ID:        podcast.User.ID,
				FirstName: podcast.User.FirstName,
				LastName:  podcast.User.LastName,
				Username:  podcast.User.Username,
			},
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
		// İmzalı URL'leri oluştur
		audioURL, err := utils.GenerateSignedURL(podcast.AudioKey, s.config)
		if err != nil {
			return nil, err
		}

		coverURL, err := utils.GenerateSignedURL(podcast.CoverKey, s.config)
		if err != nil {
			return nil, err
		}

		response.Podcasts = append(response.Podcasts, dto.PodcastResponse{
			ID:       podcast.ID,
			Title:    podcast.Title,
			Category: podcast.Category,
			AudioURL: audioURL,
			CoverURL: coverURL,
			User: dto.UserDTO{
				ID:        podcast.User.ID,
				FirstName: podcast.User.FirstName,
				LastName:  podcast.User.LastName,
				Username:  podcast.User.Username,
			},
		})
	}

	response.HasNext = hasMore
	response.HasPrevious = req.Cursor != nil

	return &response, nil
}

func (s *PodcastService) UpdatePodcast(id uint, userID uint, req *dto.UpdatePodcastRequest) (*dto.PodcastResponse, error) {
	// Podcast'i bul
	existingPodcast, err := s.podcastRepo.GetPodcastByID(id)
	if err != nil {
		return nil, err
	}

	// Yetki kontrolü
	if existingPodcast.UserID != userID {
		return nil, errors.New("bu podcast'i düzenleme yetkiniz yok")
	}

	// Güncelleme
	existingPodcast.Title = req.Title
	existingPodcast.Category = req.Category

	// Veritabanını güncelle
	if err := s.podcastRepo.UpdatePodcast(id, existingPodcast); err != nil {
		return nil, err
	}

	// İmzalı URL'leri oluştur
	audioURL, err := utils.GenerateSignedURL(existingPodcast.AudioKey, s.config)
	if err != nil {
		return nil, err
	}

	coverURL, err := utils.GenerateSignedURL(existingPodcast.CoverKey, s.config)
	if err != nil {
		return nil, err
	}

	return &dto.PodcastResponse{
		ID:       existingPodcast.ID,
		Title:    existingPodcast.Title,
		Category: existingPodcast.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
		User: dto.UserDTO{
			ID:        existingPodcast.User.ID,
			FirstName: existingPodcast.User.FirstName,
			LastName:  existingPodcast.User.LastName,
			Username:  existingPodcast.User.Username,
		},
	}, nil
}

func (s *PodcastService) DeletePodcast(id uint, userID uint) error {
	fmt.Printf("Podcast - Silme işlemi başlatıldı. PodcastID: %d, UserID: %d\n", id, userID)

	podcast, err := s.podcastRepo.GetPodcastByID(id)
	if err != nil {
		fmt.Printf("Podcast - HATA: Podcast bulunamadı. PodcastID: %d, Hata: %v\n", id, err)
		return err
	}

	fmt.Printf("Podcast - Podcast bulundu. PodcastID: %d, Title: %s\n", podcast.ID, podcast.Title)

	if podcast.UserID != userID {
		fmt.Printf("Podcast - HATA: Yetkisiz silme denemesi. PodcastID: %d, İsteyen UserID: %d, Sahip UserID: %d\n", id, userID, podcast.UserID)
		return errors.New("bu podcast'i silme yetkiniz yok")
	}

	fmt.Printf("Podcast - Yetki kontrolü başarılı. Dosya silme işlemlerine başlanıyor.\n")

	// Cloudflare R2'den dosyaları sil
	if podcast.AudioKey != "" {
		fmt.Printf("Podcast - Ses dosyası silme işlemi başlatıldı. Key: %s\n", podcast.AudioKey)
		if err := s.R2Service.DeleteFile(podcast.AudioKey); err != nil {
			fmt.Printf("Podcast - HATA: Ses dosyası R2'den silinirken hata oluştu: %v\n", err)
			return fmt.Errorf("ses dosyası silinirken hata oluştu: %v", err)
		}
		fmt.Printf("Podcast - Ses dosyası R2'den başarıyla silindi.\n")
	} else {
		fmt.Printf("Podcast - Ses dosyası Key'i boş, silme işlemi atlanıyor.\n")
	}

	if podcast.CoverKey != "" {
		fmt.Printf("Podcast - Kapak fotoğrafı silme işlemi başlatıldı. Key: %s\n", podcast.CoverKey)
		if err := s.R2Service.DeleteFile(podcast.CoverKey); err != nil {
			fmt.Printf("Podcast - HATA: Kapak fotoğrafı R2'den silinirken hata oluştu: %v\n", err)
			return fmt.Errorf("kapak fotoğrafı silinirken hata oluştu: %v", err)
		}
		fmt.Printf("Podcast - Kapak fotoğrafı R2'den başarıyla silindi.\n")
	} else {
		fmt.Printf("Podcast - Kapak fotoğrafı Key'i boş, silme işlemi atlanıyor.\n")
	}

	fmt.Printf("Podcast - Veritabanından silme işlemi başlatılıyor. PodcastID: %d\n", id)
	err = s.podcastRepo.DeletePodcast(id)
	if err != nil {
		fmt.Printf("Podcast - HATA: Veritabanından silinirken hata oluştu: %v\n", err)
		return err
	}

	fmt.Printf("Podcast - Silme işlemi başarıyla tamamlandı. PodcastID: %d\n", id)
	return nil
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
	podcasts, err := s.podcastRepo.GetLikedPodcasts(userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.PodcastResponse, 0)
	for _, p := range *podcasts {
		// İmzalı URL'leri oluştur
		audioURL, err := utils.GenerateSignedURL(p.AudioKey, s.config)
		if err != nil {
			return nil, err
		}

		coverURL, err := utils.GenerateSignedURL(p.CoverKey, s.config)
		if err != nil {
			return nil, err
		}

		response = append(response, dto.PodcastResponse{
			ID:       p.ID,
			Title:    p.Title,
			Category: p.Category,
			AudioURL: audioURL,
			CoverURL: coverURL,
			User: dto.UserDTO{
				ID:        p.User.ID,
				FirstName: p.User.FirstName,
				LastName:  p.User.LastName,
				Username:  p.User.Username,
			},
		})
	}
	return response, nil
}

func (s *PodcastService) GetPodcastsByCategory(category string) ([]dto.PodcastResponse, error) {
	podcasts, err := s.podcastRepo.GetPodcastsByCategory(category)
	if err != nil {
		return nil, err
	}

	response := make([]dto.PodcastResponse, 0)
	for _, p := range *podcasts {
		// İmzalı URL'leri oluştur
		audioURL, err := utils.GenerateSignedURL(p.AudioKey, s.config)
		if err != nil {
			return nil, err
		}

		coverURL, err := utils.GenerateSignedURL(p.CoverKey, s.config)
		if err != nil {
			return nil, err
		}

		response = append(response, dto.PodcastResponse{
			ID:       p.ID,
			Title:    p.Title,
			Category: p.Category,
			AudioURL: audioURL,
			CoverURL: coverURL,
			User: dto.UserDTO{
				ID:        p.User.ID,
				FirstName: p.User.FirstName,
				LastName:  p.User.LastName,
				Username:  p.User.Username,
			},
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
	// Podcast'i bul
	existingPodcast, err := s.podcastRepo.GetPodcastByID(id)
	if err != nil {
		return nil, err
	}

	// Yetki kontrolü
	if existingPodcast.UserID != userID {
		return nil, errors.New("bu podcast'i düzenleme yetkiniz yok")
	}

	// Eski kapak fotoğrafını sil
	if existingPodcast.CoverKey != "" {
		if err := s.R2Service.DeleteFile(existingPodcast.CoverKey); err != nil {
			fmt.Printf("Podcast - HATA: Eski kapak fotoğrafı silinirken hata oluştu: %v\n", err)
			// Hata olsa bile devam et
		}
	}

	// Yeni kapak fotoğrafını yükle
	newCoverKey, err := s.R2Service.UploadFile(coverFile, "covers")
	if err != nil {
		return nil, err
	}

	// Podcast'i güncelle
	existingPodcast.CoverKey = newCoverKey

	// Veritabanını güncelle
	if err := s.podcastRepo.UpdatePodcast(id, existingPodcast); err != nil {
		// Hata durumunda yüklenen dosyayı sil
		s.R2Service.DeleteFile(newCoverKey)
		return nil, err
	}

	// İmzalı URL'leri oluştur
	audioURL, err := utils.GenerateSignedURL(existingPodcast.AudioKey, s.config)
	if err != nil {
		return nil, err
	}

	coverURL, err := utils.GenerateSignedURL(newCoverKey, s.config)
	if err != nil {
		return nil, err
	}

	return &dto.PodcastResponse{
		ID:       existingPodcast.ID,
		Title:    existingPodcast.Title,
		Category: existingPodcast.Category,
		AudioURL: audioURL,
		CoverURL: coverURL,
		User: dto.UserDTO{
			ID:        existingPodcast.User.ID,
			FirstName: existingPodcast.User.FirstName,
			LastName:  existingPodcast.User.LastName,
			Username:  existingPodcast.User.Username,
		},
	}, nil
}
