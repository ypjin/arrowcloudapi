package db

import (
	"dao"

	"models"

	"service/auth"
)

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct{}

// Authenticate calls dao to authenticate user.
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	u, err := dao.LoginByDb(m)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func init() {
	auth.Register("db_auth", &Auth{})
}
