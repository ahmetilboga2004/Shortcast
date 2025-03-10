package handler

import (
	"fmt"
	"net/url"
	"shortcast/internal/dto"
	"shortcast/internal/service"
	"shortcast/internal/utils"

	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tcolgate/mp3"
)

type PodcastHandler struct {
	podcastService *service.PodcastService
}

func NewPodcastHandler(podcastService *service.PodcastService) *PodcastHandler {
	return &PodcastHandler{podcastService: podcastService}
}

// UploadPodcast godoc
// @Summary      Upload a podcast
// @Description  Upload a podcast with audio file and metadata
// @Tags         podcast
// @Accept       multipart/form-data
// @Produce      json
// @Param        title    formData  string  true  "Podcast title"
// @Param        category formData  string  true  "Podcast category"
// @Param        audio    formData  file    true  "Audio file"
// @Param        cover    formData  file    true  "Cover image"
// @Success      201  {object}  dto.PodcastResponse
// @Failure      400  {object}  map[string]string  "Hatalı istek"
// @Failure      500  {object}  map[string]string  "Sunucu hatası"
// @Router       /podcasts [post]
func (h *PodcastHandler) UploadPodcast(c *fiber.Ctx) error {
	// Kullanıcı bilgilerini al
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// Form verilerini parse et
	var podcastDTO dto.UploadPodcastRequest
	podcastDTO.Title = c.FormValue("title")
	podcastDTO.Category = c.FormValue("category")

	if podcastDTO.Title == "" || podcastDTO.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Başlık ve kategori alanları zorunludur",
		})
	}

	// Ses dosyasını kontrol et
	audioFile, err := c.FormFile("audio")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ses dosyası gerekli",
		})
	}

	// Dosya uzantısını kontrol et
	if !strings.HasSuffix(strings.ToLower(audioFile.Filename), ".mp3") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Sadece MP3 formatı kabul edilmektedir",
		})
	}

	// Dosyayı aç
	file, err := audioFile.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Dosya açılamadı",
		})
	}
	defer file.Close()

	var duration float64
	decoder := mp3.NewDecoder(file)
	var frame mp3.Frame
	skipped := 0
	for {
		if err := decoder.Decode(&frame, &skipped); err != nil {
			break
		}
		duration += frame.Duration().Seconds()
	}

	if duration > 60 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ses dosyası 60 saniyeden uzun olamaz",
		})
	}

	// Kapak fotoğrafını kontrol et
	coverFile, err := c.FormFile("cover")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Kapak fotoğrafı gerekli",
		})
	}

	podcastDTO.UserID = userID

	// Servis katmanına yönlendir
	podcastResponse, err := h.podcastService.UploadPodcast(&podcastDTO, audioFile, coverFile)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(podcastResponse)
}

// GetPodcastByID godoc
// @Summary      Get podcast by ID
// @Description  Retrieve podcast details by ID
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Podcast ID"
// @Success      200  {object}  dto.PodcastResponse
// @Failure      400  {object}  map[string]string  "Geçersiz ID formatı"
// @Failure      404  {object}  map[string]string  "Podcast bulunamadı"
// @Router       /podcasts/{id} [get]
func (h *PodcastHandler) GetPodcastByID(c *fiber.Ctx) error {
	id, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz ID formatı"})
	}

	podcastResponse, err := h.podcastService.GetPodcastByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Podcast bulunamadı"})
	}

	return c.Status(fiber.StatusOK).JSON(podcastResponse)
}

// GetUserPodcasts godoc
// @Summary      Get user's podcasts
// @Description  Retrieve all podcasts of a specific user
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        user_id   path      int  true  "User ID"
// @Success      200  {array}   dto.PodcastResponse
// @Failure      400  {object}  map[string]string  "Geçersiz ID formatı"
// @Failure      404  {object}  map[string]string  "Kullanıcı bulunamadı"
// @Router       /users/{user_id}/podcasts [get]
func (h *PodcastHandler) GetUserPodcasts(c *fiber.Ctx) error {
	userID, err := utils.ParamAsUint(c, "user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz kullanıcı ID formatı",
		})
	}

	podcasts, err := h.podcastService.GetUserPodcasts(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Podcastler getirilirken bir hata oluştu",
		})
	}

	// Eğer podcast yoksa boş array dön
	if podcasts == nil {
		podcasts = []dto.PodcastResponse{}
	}

	return c.JSON(fiber.Map{
		"podcasts": podcasts,
	})
}

// DiscoverPodcasts godoc
// @Summary      Discover podcasts
// @Description  Get paginated podcasts for discovery
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        cursor     query    integer  false  "Cursor for pagination"
// @Param        direction  query    string   false  "Direction (next/prev)"
// @Param        limit      query    integer  false  "Number of podcasts per page"
// @Success      200  {object}  dto.PodcastCursor
// @Failure      400  {object}  map[string]string  "Geçersiz istek"
// @Failure      500  {object}  map[string]string  "Sunucu hatası"
// @Router       /podcasts/discover [get]
func (h *PodcastHandler) DiscoverPodcasts(c *fiber.Ctx) error {
	var req dto.PodcastDiscoverRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz sorgu parametreleri",
		})
	}

	result, err := h.podcastService.DiscoverPodcasts(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Podcastler getirilirken bir hata oluştu",
		})
	}

	// Eğer podcast listesi null ise boş array oluştur
	if result.Podcasts == nil {
		result.Podcasts = []dto.PodcastResponse{}
	}
	return c.JSON(result)
}

// UpdatePodcast godoc
// @Summary      Update a podcast
// @Description  Update podcast title and category
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Podcast ID"
// @Param        podcast body dto.UpdatePodcastRequest true "Podcast update info"
// @Success      200  {object}  dto.PodcastResponse
// @Failure      400  {object}  map[string]string  "Geçersiz istek"
// @Failure      401  {object}  map[string]string  "Yetkisiz erişim"
// @Failure      404  {object}  map[string]string  "Podcast bulunamadı"
// @Router       /podcasts/{id} [put]
func (h *PodcastHandler) UpdatePodcast(c *fiber.Ctx) error {
	id, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz podcast ID",
		})
	}

	// Kullanıcı ID'sini al
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// Request body'i parse et
	var req dto.UpdatePodcastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz istek formatı",
		})
	}

	// Validasyon
	if req.Title == "" || req.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Başlık ve kategori gereklidir",
		})
	}

	// Service'i çağır
	updatedPodcast, err := h.podcastService.UpdatePodcast(id, userID, &req)
	if err != nil {
		if err.Error() == "podcast bulunamadı" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Podcast bulunamadı",
			})
		}
		if err.Error() == "bu podcast'i düzenleme yetkiniz yok" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Bu podcast'i düzenleme yetkiniz yok",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Podcast güncellenirken bir hata oluştu",
		})
	}

	return c.JSON(updatedPodcast)
}

// DeletePodcast godoc
// @Summary      Delete a podcast
// @Description  Delete a podcast by ID (only owner can delete)
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Podcast ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]string  "Geçersiz podcast ID"
// @Failure      401  {object}  map[string]string  "Yetkisiz erişim"
// @Failure      404  {object}  map[string]string  "Podcast bulunamadı"
// @Router       /podcasts/{id} [delete]
func (h *PodcastHandler) DeletePodcast(c *fiber.Ctx) error {
	id, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz podcast ID",
		})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	err = h.podcastService.DeletePodcast(id, userID)
	if err != nil {
		if err.Error() == "podcast bulunamadı" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Podcast bulunamadı",
			})
		}
		if err.Error() == "bu podcast'i silme yetkiniz yok" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Bu podcast'i silme yetkiniz yok",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Podcast silinirken bir hata oluştu",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// LikePodcast godoc
// @Summary      Like or unlike a podcast
// @Description  Like a podcast if not liked, unlike if already liked
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Podcast ID"
// @Success      200  {object}  dto.LikeResponse
// @Failure      400  {object}  map[string]string  "Geçersiz podcast ID"
// @Failure      500  {object}  map[string]string  "İşlem başarısız"
// @Router       /podcasts/{id}/like [post]
func (h *PodcastHandler) LikePodcast(c *fiber.Ctx) error {
	podcastID, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz podcast ID",
		})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	response, err := h.podcastService.LikePodcast(podcastID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "İşlem başarısız oldu",
		})
	}

	return c.JSON(response)
}

// GetLikedPodcasts godoc
// @Summary      Get liked podcasts
// @Description  Get all podcasts liked by the authenticated user
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Success      200  {array}   dto.PodcastResponse
// @Failure      500  {object}  map[string]string  "Sunucu hatası"
// @Router       /podcasts/liked [get]
func (h *PodcastHandler) GetLikedPodcasts(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	podcasts, err := h.podcastService.GetLikedPodcasts(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Beğenilen podcastler getirilirken bir hata oluştu",
		})
	}

	// Eğer podcast yoksa boş array dön
	if podcasts == nil {
		podcasts = []dto.PodcastResponse{}
	}

	return c.JSON(fiber.Map{
		"podcasts": podcasts,
	})
}

// GetPodcastsByCategory godoc
// @Summary      Get podcasts by category
// @Description  Get all podcasts in a specific category
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        category   path      string  true  "Category name"
// @Success      200  {array}   dto.PodcastResponse
// @Failure      400  {object}  map[string]string  "Kategori belirtilmedi"
// @Failure      500  {object}  map[string]string  "Sunucu hatası"
// @Router       /podcasts/category/{category} [get]
func (h *PodcastHandler) GetPodcastsByCategory(c *fiber.Ctx) error {
	category := c.Params("category")
	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Kategori belirtilmedi",
		})
	}

	// URL decode işlemi
	decodedCategory, err := url.QueryUnescape(category)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz kategori formatı",
		})
	}

	podcasts, err := h.podcastService.GetPodcastsByCategory(decodedCategory)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Podcastler getirilirken bir hata oluştu",
		})
	}

	// Eğer podcast yoksa boş array dön
	if podcasts == nil {
		podcasts = []dto.PodcastResponse{}
	}

	return c.JSON(fiber.Map{
		"podcasts": podcasts,
	})
}

// AddComment godoc
// @Summary      Add comment to podcast
// @Description  Add a new comment to a podcast
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        id      path      int  true  "Podcast ID"
// @Param        comment body      dto.CommentRequest  true  "Comment content"
// @Success      201  {object}  dto.CommentResponse
// @Failure      400  {object}  map[string]string  "Geçersiz istek"
// @Failure      500  {object}  map[string]string  "Sunucu hatası"
// @Router       /podcasts/{id}/comments [post]
func (h *PodcastHandler) AddComment(c *fiber.Ctx) error {
	podcastID, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz podcast ID",
		})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var req dto.CommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz yorum formatı",
		})
	}

	comment, err := h.podcastService.AddComment(podcastID, userID, req.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Yorum eklenirken bir hata oluştu",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// GetComments godoc
// @Summary      Get podcast comments
// @Description  Get all comments for a specific podcast
// @Tags         podcast
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Podcast ID"
// @Success      200  {array}   dto.CommentResponse
// @Failure      400  {object}  map[string]string  "Geçersiz podcast ID"
// @Failure      500  {object}  map[string]string  "Sunucu hatası"
// @Router       /podcasts/{id}/comments [get]
func (h *PodcastHandler) GetComments(c *fiber.Ctx) error {
	podcastID, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz podcast ID",
		})
	}

	comments, err := h.podcastService.GetComments(podcastID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Yorumlar getirilirken bir hata oluştu",
		})
	}

	// Eğer yorum yoksa boş array dön
	if comments == nil {
		comments = []dto.CommentResponse{}
	}

	return c.JSON(fiber.Map{
		"comments": comments,
	})
}

// UpdatePodcastCover godoc
// @Summary      Update podcast cover
// @Description  Update cover image of a podcast
// @Tags         podcast
// @Accept       multipart/form-data
// @Produce      json
// @Param        id    path      int   true  "Podcast ID"
// @Param        cover formData  file  true  "Cover image"
// @Success      200  {object}  dto.PodcastResponse
// @Failure      400  {object}  map[string]string  "Geçersiz istek"
// @Failure      401  {object}  map[string]string  "Yetkisiz erişim"
// @Failure      404  {object}  map[string]string  "Podcast bulunamadı"
// @Router       /podcasts/{id}/cover [put]
func (h *PodcastHandler) UpdatePodcastCover(c *fiber.Ctx) error {
	id, err := utils.ParamAsUint(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz podcast ID",
		})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	coverFile, err := c.FormFile("cover")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Kapak fotoğrafı gerekli",
		})
	}

	podcastResponse, err := h.podcastService.UpdatePodcastCover(id, userID, coverFile)
	if err != nil {
		if err.Error() == "podcast bulunamadı" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Podcast bulunamadı",
			})
		}
		if err.Error() == "bu podcast'i düzenleme yetkiniz yok" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Kapak fotoğrafı güncellenirken bir hata oluştu",
		})
	}

	return c.JSON(podcastResponse)
}

// GetFileContent godoc
// @Summary      Get file content from R2
// @Description  Get file content directly from R2 storage
// @Tags         podcast
// @Accept       json
// @Produce      octet-stream
// @Param        key   path      string  true  "File key"
// @Success      200  {file}    binary
// @Failure      400  {object}  map[string]string  "Geçersiz key"
// @Failure      404  {object}  map[string]string  "Dosya bulunamadı"
// @Router       /podcasts/file/{key} [get]
func (h *PodcastHandler) GetFileContent(c *fiber.Ctx) error {
	key := c.Params("*") // wildcard parametresini al
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Dosya anahtarı gerekli",
		})
	}

	// Dosya içeriğini al
	content, err := h.podcastService.R2Service.GetFile(key)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Dosya bulunamadı",
		})
	}

	// Content-Type'ı belirle
	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(key), ".mp3") {
		contentType = "audio/mpeg"
	} else if strings.HasSuffix(strings.ToLower(key), ".jpg") || strings.HasSuffix(strings.ToLower(key), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(key), ".png") {
		contentType = "image/png"
	}

	// Dosya adını al
	fileName := key[strings.LastIndex(key, "/")+1:]

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	return c.Send(content)
}
