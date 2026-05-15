package user

import (
	"context"

	"gorm.io/gorm"
)

type Model struct {
	UserID string `gorm:"column:user_id;primaryKey"`
	Name   string `gorm:"column:name"`
	Status string `gorm:"column:status"`
}

func (Model) TableName() string { return "users" }

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Add(ctx context.Context, u Model) (Model, error) {
	if err := r.db.WithContext(ctx).Create(&u).Error; err != nil {
		return Model{}, err
	}
	return u, nil
}

func (r *Repo) Query(ctx context.Context, userID string) (Model, bool, error) {
	var out Model
	err := r.db.WithContext(ctx).First(&out, "user_id = ?", userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return Model{}, false, nil
		}
		return Model{}, false, err
	}
	return out, true, nil
}
