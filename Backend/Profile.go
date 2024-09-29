package forum

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"forum/Error"
	"net/http"
)

type ProfileTemplateData struct {
	Posts         []Post
	Username      string
	ProfileImg    string
	LikedPosts    []string
	LikedComments []string
	Createdposts  []string
	CreatedComments []string
	ClickedButton string
	NumNotifications int
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/Profile"  || guest {
		Error.RenderErrorPage(w,404,"Page Not Found")
	return
	}
	var user Account
	user = GetUserData(w, r)

	filterValue := r.FormValue("filterValue")
	// Check if the Content-Type is application/json
	if r.Header.Get("Content-Type") == "application/json" {

		var request Request
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Failed to decode JSON request", http.StatusBadRequest)
			return
		}
		// Process the request
		switch request.RequestType {
		case "like":
			HandleLikeRequest(w, r, user, request)
		case "delete":
			HandleDeleteRequest(request)
		case "DeleteComment":
			HandleDeleteCommentRequest(request)
		case "UpdateComment": 
			HandleUpdateCommentRequest(request)
		default:
			http.Error(w, "Invalid request type", http.StatusBadRequest)
			return
		}
	}

	switch filterValue {
	case "CreatedPosts":
		ProfilePost(w, r, "CreatedPosts", user)
	case "LikedPosts":
		ProfilePost(w, r, "LikedPosts", user)
	case "DislikedPosts":
		ProfilePost(w, r, "DislikedPosts", user)
	case "CommentedPosts":
		ProfilePost(w, r, "CommentedPosts", user)
	default:
		ProfilePost(w, r, "CreatedPosts", user)
	}

}

func ProfilePost(w http.ResponseWriter, r *http.Request, filter string, user Account) {
	var posts []Post
	clickedButton := "CreatedPosts"

	//Get Created Posts id
	Createdposts := GetCreatedPosts(user.Id)
	LikedPostsId, DislikedPostsId, Likedposts := GetLikedPosts(user.Id)
	Likedcomments := GetLikedComments(user.Id)
	createdcomments := GetCreatedComments(user.Id)

	//To do the filter adjust the fetchPostsFromDB function to accept filters
	if filter == "CreatedPosts" {
		posts, err = fetchPostsFromDB(true, "ProfileFilter", Createdposts)
		if err != nil {
			log.Fatal(err)
		}
		clickedButton = "CreatedPosts"
	} else if filter == "LikedPosts" {
		posts, err = fetchPostsFromDB(true, "ProfileFilter", LikedPostsId)
		if err != nil {
			log.Fatal(err)
		}
		clickedButton = "LikedPosts"
	}else if filter == "DislikedPosts" {
		posts, err = fetchPostsFromDB(true, "ProfileFilter", DislikedPostsId)
		if err != nil {
			log.Fatal(err)
		}
		clickedButton = "DislikedPosts"
	}else if filter == "CommentedPosts" {
		CommentedPostsId := GetCommentedPosts(user.Id)
		posts, err = fetchPostsFromDB(true, "ProfileFilter", CommentedPostsId)
		if err != nil {
			log.Fatal(err)
		}
		clickedButton = "CommentedPosts"
	}

	//Reverse Posts from new to old
	reverseArray(posts)

	numNotifications := GetNumOfNotifications(user.Id)

	data := ProfileTemplateData{
		Posts:         posts,
		Username:      user.Username,
		ProfileImg:    user.ProfileImg,
		LikedPosts:    Likedposts,
		LikedComments: Likedcomments,
		Createdposts:  Createdposts,
		CreatedComments: createdcomments,
		ClickedButton: clickedButton,
		NumNotifications: numNotifications,
	}

	_, err = os.Stat("./Pages/Profile.html")
	
	if os.IsNotExist(err) {
		log.Println("[ERROR] - File 'Profile.html' does not exist or is not accessible.")
	
		w.WriteHeader(http.StatusInternalServerError)
		Error.RenderErrorPage(w,500,"Internal Server Error")
	} else {
	// Reload the templates by parsing the template files
	tmpl, err := template.ParseFiles("./Pages/Profile.html", "./Pages/nav.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
}

