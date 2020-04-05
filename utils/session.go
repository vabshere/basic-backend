package utils

import (
	"net/http"
	"web-server/models"
	"web-server/utils/session"
)

// GlobalSessions is the global variable for managing sessions
var GlobalSessions *session.Manager

// Run initializes the one time configurations required for using sessions
func Run() {
	GlobalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	go GlobalSessions.GC()
}

// SessionSetUser is used for setting given user's details in given session
func SessionSetUser(user *models.User, session *session.Session, r *http.Request) {
	(*session).Set("id", user.Id)
	(*session).Set("name", user.Name)
	(*session).Set("username", user.Username)
	(*session).Set("profilePic", user.ProfilePic)
}

// SessionGetUser returns user details from given session
func SessionGetUser(session *session.Session, r *http.Request) *models.User {
	id := (*session).Get("id").(int)
	name := (*session).Get("name").(string)
	username := (*session).Get("username").(string)
	profilePic := (*session).Get("profilePic").(string)
	u := models.User{Id: id, Name: name, Username: username, ProfilePic: profilePic}
	return &u
}
