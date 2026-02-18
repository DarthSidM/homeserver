package services

import (
	"errors"
	"os"
	"time"

	"homeserver/internals/models"
	"homeserver/internals/repos"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserAuthService struct {
	repo *repos.UserRepository
}

func NewUserAuthService(repo *repos.UserRepository) *UserAuthService {
	return &UserAuthService{repo: repo}
}

func (s *UserAuthService) CreateUser(username, name, password string) (*models.User, error) {
	hashedPassword, err := s.CreatePasswordHash(password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(username, name, hashedPassword)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserAuthService) FindUserByUsername(username string) (*models.User, error) {
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserAuthService) CreatePasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (s *UserAuthService) CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *UserAuthService) IssueJWT(username string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
