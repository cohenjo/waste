package types

import (
	"errors"
	"strings"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"-"`
}

var userList = []User{
	User{Username: "user1", Password: "pass1"},
	User{Username: "user2", Password: "pass2"},
	User{Username: "user3", Password: "pass3"},
}

func RegisterNewUser(username, password string) (*User, error) {
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("The password can't be empty")
	} else if !IsUsernameAvailable(username) {
		return nil, errors.New("The username isn't available")
	}

	u := User{Username: username, Password: password}

	userList = append(userList, u)

	return &u, nil
}

func IsUserValid(username, password string) bool {
	for _, u := range userList {
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}

func IsUsernameAvailable(username string) bool {
	for _, u := range userList {
		if u.Username == username {
			return false
		}
	}
	return true
}
