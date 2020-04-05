package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"web-server/models"
	"web-server/utils"

	"golang.org/x/crypto/bcrypt"
)

// SignUp creates a new user in the database and creates its session
func SignUp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("called up")
	utils.GlobalSessions.SessionDestroy(w, r)
	decoder := json.NewDecoder(r.Body)
	var u models.User
	err := decoder.Decode(&u)
	if err != nil {
		utils.Respond(1, "Error", http.StatusInternalServerError, w, r)
		return
	}

	hash, err := bcrypt.GenerateFromPassword(u.Password, bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
	}

	u.Password = hash
	if len(u.Name) == 0 || len(u.Username) == 0 || len(u.Password) == 0 {
		utils.Respond(0, "Invalid submission", http.StatusBadRequest, w, r)
		return
	}

	err = models.SaveUser(&u)
	if err != nil {
		s := string(err.Error())
		if s[len("Error "):len("Error 1062")] == "1062" {
			utils.Respond(1, "Username already taken", http.StatusOK, w, r)
			return
		}
		utils.Respond(1, "Error", http.StatusInternalServerError, w, r)
		return
	}

	session := utils.GlobalSessions.SessionStart(w, r)
	utils.SessionSetUser(&u, &session, r)
	utils.RespondJson(0, u, http.StatusCreated, w, r)
}

// SignIn checks if the user exists in the database and creates a session
func SignIn(w http.ResponseWriter, r *http.Request) {
	utils.GlobalSessions.SessionDestroy(w, r)
	decoder := json.NewDecoder(r.Body)
	var u models.User
	err := decoder.Decode(&u)
	if err != nil {
		utils.Respond(1, "Error", http.StatusInternalServerError, w, r)
		return
	}
	fmt.Println(u)
	user, err := models.GetUserByUsername(u.Username)
	if err != nil {
		fmt.Println(err)
		utils.Respond(1, "Error", http.StatusInternalServerError, w, r)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, u.Password); err != nil {
		fmt.Println(err)
		utils.Respond(1, "Authentication failed", http.StatusOK, w, r)
		return
	}

	session := utils.GlobalSessions.SessionStart(w, r)
	utils.SessionSetUser(user, &session, r)
	utils.RespondJson(0, user, http.StatusOK, w, r)
}

// GetUser returns user from the session
func GetUser(w http.ResponseWriter, r *http.Request) {
	session, b := utils.GlobalSessions.SessionCheck(r)
	if b {
		u := utils.SessionGetUser(&session, r)
		utils.RespondJson(0, u, http.StatusOK, w, r)
		return
	}

	utils.RespondJson(1, nil, http.StatusOK, w, r)
	return
}

// SignOut deletes the user session and closes the associated chat connection
func SignOut(w http.ResponseWriter, r *http.Request) {
	deleteFromPool(r)
	utils.GlobalSessions.SessionDestroy(w, r)
	utils.Respond(0, "Success", http.StatusOK, w, r)
	return
}

// ProfilePicUpload uploads/updates an image of the user in session
func ProfilePicUpload(w http.ResponseWriter, r *http.Request) {
	session, b := utils.GlobalSessions.SessionCheck(r)
	if !b {
		utils.Respond(10, "No session for this user", http.StatusOK, w, r)
		return
	}

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("image")
	if err != nil {
		utils.Respond(1, "Error uploading image1", http.StatusInternalServerError, w, r)
		return
	}

	defer file.Close()
	// fmt.Fprintf(w, "%v", handler.Header)
	if filepath.Ext(handler.Filename) != ".png" && filepath.Ext(handler.Filename) != ".jpeg" && filepath.Ext(handler.Filename) != ".jpg" {
		utils.Respond(1, "Invalid file", http.StatusInternalServerError, w, r)
		return
	}
	filename := session.Get("username").(string) + filepath.Ext(handler.Filename)
	if exists, _ := exists("./profilePics"); !exists {
		err := os.Mkdir("./profilePics", 0777)
		if err != nil {
			utils.Respond(1, "Error uploading image2", http.StatusInternalServerError, w, r)
			return
		}
	}

	f, err := os.OpenFile("./profilePics/"+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		utils.Respond(1, "Error uploading image3", http.StatusInternalServerError, w, r)
		return
	}

	err = models.UpdateProfilePic(session.Get("id").(int), 0, filename)
	if err != nil {
		utils.Respond(1, "Error uploading image4", http.StatusInternalServerError, w, r)
		return
	}

	defer f.Close()
	io.Copy(f, file)
	utils.Respond(0, filename, http.StatusOK, w, r)
}

// DeleteProfilePic deletes the image of the user and update its profilePic property
func DeleteProfilePic(w http.ResponseWriter, r *http.Request) {
	session, b := utils.GlobalSessions.SessionCheck(r)
	if !b {
		utils.Respond(10, "No session for this user", http.StatusOK, w, r)
		return
	}

	filename := session.Get("username").(string)
	err := models.UpdateProfilePic(session.Get("id").(int), 1, "")
	if err != nil {
		utils.Respond(1, "Error removing image", http.StatusInternalServerError, w, r)
		return
	}

	files, _ := filepath.Glob("./profilePics/" + filename + "*")
	if files == nil {
		utils.Respond(1, "Error removing image", http.StatusInternalServerError, w, r)
		return
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			utils.Respond(1, "Error removing image", http.StatusInternalServerError, w, r)
			return
		}
	}

	utils.Respond(0, "Success", http.StatusOK, w, r)
}

// exists returns whether the given path (file or directory) exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}
