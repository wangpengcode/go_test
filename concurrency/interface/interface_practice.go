package _interface

import (
	"context"
)

type UserModelTest struct {
	Name   string
	UserId string
}

type UserRepoTest interface {
	Add(ctx context.Context, u UserModelTest) (UserModelTest, error)
	Query(ctx context.Context, userID string) (UserModelTest, bool, error)
}

type UserRepoImpl struct {
	userRepoTest *UserRepoTest
}

func InterfacePractice() {
	a := new(UserRepoImpl)
	user, _, _ := a.Query(context.Background(), "123")
	println(user.UserId, " ", user.Name)
}

func (a UserRepoImpl) Query(ctx context.Context, userID string) (UserModelTest, bool, error) {
	userModelTest := UserModelTest{UserId: "123", Name: "joshua"}
	return userModelTest, false, nil
}

func (a UserRepoImpl) Add(ctx context.Context, u UserModelTest) (UserModelTest, error) {
	userModelTest := UserModelTest{UserId: "123", Name: "joshua"}
	return userModelTest, nil
}
