package main

import (
	forum "forum/Backend"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

func main() {

	fs := http.FileServer(http.Dir("Style"))
	http.Handle("/Style/", http.StripPrefix("/Style/", fs))

	fs2 := http.FileServer(http.Dir("Error"))
	http.Handle("/Error/", http.StripPrefix("/Error/", fs2))

	posts := http.FileServer(http.Dir("Posts"))
	http.Handle("/Posts/", http.StripPrefix("/Posts/", posts))
	profileImages := http.FileServer(http.Dir("ProfileImages"))
	http.Handle("/ProfileImages/", http.StripPrefix("/ProfileImages/", profileImages))

	http.HandleFunc("/", forum.Login)
	http.HandleFunc("/SignUp", forum.SignUpHandler)
	http.HandleFunc("/LogOut", forum.LogoutHandler)
	http.HandleFunc("/HomePage", forum.AuthMiddleware(forum.HomeHandler))
	http.HandleFunc("/CreatePost",forum.AuthMiddleware(forum.CreatePostHandler))
	http.HandleFunc("/Profile", forum.AuthMiddleware(forum.ProfileHandler))
	http.HandleFunc("/EditPost", forum.AuthMiddleware(forum.EditPostHandler))
	http.HandleFunc("/Notifications", forum.AuthMiddleware(forum.NotificationsHandler))
	http.HandleFunc("/CommentHandler", forum.CommentHandler)
	http.HandleFunc("/ProfileImageHandler", forum.ProfileImageHandler)
	http.HandleFunc("/CommentLikeHandle", forum.CommentLikeHandle)
	
	forum.CreateTables()

	log.Println("Server started on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
