package dao

import (
	"errors"
	"time"

	"models"
	"utils"
)

// Register is used for user to register, the password is encrypted before the record is inserted into database.
func Register(user models.User) (int64, error) {
	o := GetOrmer()
	p, err := o.Raw("insert into user (username, password, realname, email, comment, salt, sysadmin_flag, creation_time, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?)").Prepare()
	if err != nil {
		return 0, err
	}
	defer p.Close()

	salt := utils.GenerateRandomString()

	now := time.Now()
	r, err := p.Exec(user.Username, utils.Encrypt(user.Password, salt), user.Realname, user.Email, user.Comment, salt, user.HasAdminRole, now, now)

	if err != nil {
		return 0, err
	}
	userID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// UserExists returns whether a user exists according username or Email.
func UserExists(user models.User, target string) (bool, error) {

	if user.Username == "" && user.Email == "" {
		return false, errors.New("User name and email are blank.")
	}

	o := GetOrmer()

	sql := `select user_id from user where 1=1 `
	queryParam := make([]interface{}, 1)

	switch target {
	case "username":
		sql += ` and username = ? `
		queryParam = append(queryParam, user.Username)
	case "email":
		sql += ` and email = ? `
		queryParam = append(queryParam, user.Email)
	}

	var u []models.User
	n, err := o.Raw(sql, queryParam).QueryRows(&u)
	if err != nil {
		return false, err
	} else if n == 0 {
		return false, nil
	} else {
		return true, nil
	}
}
