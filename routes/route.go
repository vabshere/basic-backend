package routes

import (
	"net/http"
	"web-server/controllers"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

//Init initializes routes for the app
func Init() *mux.Router {
	pool := controllers.NewPool()
	go pool.Start()

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./profilePics"))))
	r.HandleFunc("/api/signUp", controllers.SignUp).Methods(http.MethodPost)
	r.HandleFunc("/api/signIn", controllers.SignIn).Methods(http.MethodPost)
	r.HandleFunc("/api/profilePic", controllers.ProfilePicUpload).Methods(http.MethodPost)
	r.HandleFunc("/api/delProfilePic", controllers.DeleteProfilePic).Methods(http.MethodDelete)
	r.HandleFunc("/api/getUser", controllers.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/api/signOut", controllers.SignOut).Methods(http.MethodDelete)
	r.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		controllers.Chat(pool, ws)
	}))
	// r.HandleFunc("/", greet)
	return r
}

// func greet(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello World! %s", time.Now())
// }
