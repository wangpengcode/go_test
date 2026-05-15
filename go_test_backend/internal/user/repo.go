package user

import (
	"context"

	"gorm.io/gorm"
)

type UserModel struct {
	UserID string `gorm:"column:user_id;primaryKey"`
	Name   string `gorm:"column:name"`
	Status string `gorm:"column:status"`
}

// TableName 告诉 GORM：Model 对应的表名是什么。
func (UserModel) TableName() string { return "users" }

type Repo struct {
	db *gorm.DB
}

// NewRepo 创建一个 Repo。
func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

// Add 插入一条用户数据。
func (r *Repo) Add(ctx context.Context, u UserModel) (UserModel, error) {
	if err := r.db.WithContext(ctx).Create(&u).Error; err != nil {
		return UserModel{}, err
	}
	return u, nil
}

// Query 按 userID 查询用户。
// 返回值含义：（model, 是否找到 found, error）。
func (r *Repo) Query(ctx context.Context, userID string) (UserModel, bool, error) {
	var out UserModel
	err := r.db.WithContext(ctx).First(&out, "user_id = ?", userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return UserModel{}, false, nil
		}
		return UserModel{}, false, err
	}
	return out, true, nil
}
