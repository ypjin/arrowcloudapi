package dao

import (
	"fmt"
	"testing"

	"arrowcloudapi/models"
)

func TestDeleteUser(t *testing.T) {
	username := "user_for_test"
	email := "user_for_test@vmware.com"
	password := "P@ssword"
	realname := "user_for_test"

	u := models.User{
		Username: username,
		Email:    email,
		Password: password,
		Realname: realname,
	}
	id, err := Register(u)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	err = DeleteUser(int(id))
	if err != nil {
		t.Fatalf("Error occurred in DeleteUser: %v", err)
	}

	user := &models.User{}
	sql := "select * from user where user_id = ?"
	if err = GetOrmer().Raw(sql, id).
		QueryRow(user); err != nil {
		t.Fatalf("failed to query user: %v", err)
	}

	if user.Deleted != 1 {
		t.Error("user is not deleted")
	}

	expected := fmt.Sprintf("%s#%d", u.Username, id)
	if user.Username != expected {
		t.Errorf("unexpected username: %s != %s", user.Username,
			expected)
	}

	expected = fmt.Sprintf("%s#%d", u.Email, id)
	if user.Email != expected {
		t.Errorf("unexpected email: %s != %s", user.Email,
			expected)
	}
}
