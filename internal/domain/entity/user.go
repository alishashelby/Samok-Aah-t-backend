package entity

import "time"

type User struct {
	ID         int64
	AuthID     int64
	Name       string
	BirthDate  time.Time
	IsVerified bool
}

func NewUser(authID int64, name string, birthDate time.Time) *User {
	return &User{
		AuthID:    authID,
		Name:      name,
		BirthDate: birthDate,
	}
}

func (u User) IsAnAdult(now time.Time) bool {
	return u.BirthDate.AddDate(18, 0, 0).Before(now)
}

func (u User) IsUserVerified() bool {
	return u.IsVerified
}
