package models

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// connectDB connects the golang server to the database server
func connectDb() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:mysql@/basic?charset=utf8")
	if err != nil {
		return nil, err
	}
	return db, nil
}

type password []byte

// MarshalJSON is the custom method used for marshalling by JSON.Marshal
func (password) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

// User is the tye of all users
type User struct {
	Name       string   `json:"name"`
	Username   string   `json:"username"`
	Password   password `json:"password"`
	ProfilePic string   `json:"profilePic"`
	Id         int      `json:"id"`
}

// SaveUser saves a user into the database
func SaveUser(u *User) error {
	db, err := connectDb()
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO users (name, username, password) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.Name, u.Username, u.Password)
	return err
}

// UpdateProfilePic updates a user's profilePic property in the database
func UpdateProfilePic(id, code int, filename string) error {
	db, err := connectDb()
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare("UPDATE users SET profilePic=? WHERE id=?")
	if err != nil {
		return err
	}
	var p string
	if code == 0 {
		p = filename
	} else if code == 1 {
		p = ""
	}
	_, err = stmt.Exec(p, id)
	return err
}

// GetUserById returns the user associated with given Id
func GetUserById(id int) (*User, error) {
	db, err := connectDb()
	fmt.Println(id)
	defer db.Close()
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("SELECT * FROM users WHERE id=?")
	if err != nil {
		return nil, err
	}
	var user User
	err = stmt.QueryRow(id).Scan(&user.Id, &user.Name, &user.Username, &user.Password, &user.ProfilePic)
	return &user, err
}

// GetUserByUsername returns the user associated with given username
func GetUserByUsername(u string) (*User, error) {
	db, err := connectDb()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("SELECT * FROM users WHERE username=?")
	if err != nil {
		return nil, err
	}
	var user User
	err = stmt.QueryRow(u).Scan(&user.Id, &user.Name, &user.Username, &user.Password, &user.ProfilePic)
	return &user, err
}
