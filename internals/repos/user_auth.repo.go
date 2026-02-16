package repos

import (
	"homeserver/internals/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(username, name, password string) (*models.User, error) {
	user := &models.User{
		Username: username,
		Name:     name,
		Password: password,
	}

	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User

	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
