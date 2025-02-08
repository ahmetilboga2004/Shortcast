package service

import (
	"errors"
	"shortcast/internal/config"
	"shortcast/internal/dto"
	"shortcast/internal/model"
	"shortcast/internal/repository"
	"shortcast/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	authRepo *repository.AuthRepository
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(authRepo *repository.AuthRepository, userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) error {

	_, err := s.userRepo.GetUserByEmail(req.Email)
	if err == nil {
		return errors.New("bu email adresi zaten kullanılıyor")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	_, err = s.userRepo.GetUserByUsername(req.Username)
	if err == nil {
		return errors.New("bu kullanıcı adı zaten kullanılıyor")
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := model.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
	}

	return s.authRepo.CreateUser(&user)
}

func (s *AuthService) Login(emailOrUsername, password string) (string, error) {
	var user *model.User
	var err error

	user, err = s.userRepo.GetUserByEmail(emailOrUsername)
	if err != nil {
		user, err = s.userRepo.GetUserByUsername(emailOrUsername)
		if err != nil {
			return "", errors.New("kullanıcı bulunamadı")
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("şifre yanlış")
	}

	token, err := utils.GenerateJWT(user.ID, s.cfg.SecretKey)
	if err != nil {
		return "", errors.New("token oluşturulurken bir hata oluştu")
	}

	return token, nil
}

func (s *AuthService) Logout(token *jwt.Token) error {
	// Token'ın raw halini al
	tokenStr := token.Raw

	// Token'ın süresini al
	claims := token.Claims.(jwt.MapClaims)
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("token süresi alınamadı")
	}

	// Token'ı blacklist'e ekle
	return s.authRepo.BlacklistToken(tokenStr, time.Unix(int64(exp), 0))
}
