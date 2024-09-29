package forum

import (
	"forum/Error"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type CreatePostTemplateData struct {
	Username         string
	ProfileImg       string
	NumNotifications int
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/CreatePost" || guest {
		Error.RenderErrorPage(w, 404, "Page Not Found")
		return
	}
	var user Account
	user = GetUserData(w, r)

	if r.Method == "GET" {

		_, err := os.Stat("./Pages/CreatePost.html")

		if os.IsNotExist(err) {
			log.Println("[ERROR] - File 'CreatePost.html' does not exist or is not accessible.")

			w.WriteHeader(http.StatusInternalServerError)
			Error.RenderErrorPage(w, 500, "Internal Server Error")
		} else {

			numNotifications := GetNumOfNotifications(user.Id)
			
			var tmpl = template.Must(template.ParseFiles("./Pages/CreatePost.html", "./Pages/nav.html"))

			data := CreatePostTemplateData{
				Username:         user.Username,
				ProfileImg:       user.ProfileImg,
				NumNotifications: numNotifications,
			}

			err = tmpl.Execute(w, data)
			if err != nil {
				Error.RenderErrorPage(w, http.StatusInternalServerError, "Error executing template")
				return
			}
		}
	} else if r.Method == "POST" {
		// Parse the form data
		err = r.ParseMultipartForm(20 << 20) //20 MB maximum file size
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the uploaded image file
		file, handler, err := r.FormFile("file")
		if err != nil && err != http.ErrMissingFile {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var imgPath string
		if file != nil {
			defer file.Close()
			// Save the image file
			imgPath = "Posts/" + handler.Filename
			distinationFile, err := os.OpenFile(imgPath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer distinationFile.Close()
			io.Copy(distinationFile, file)
		} else {
			imgPath = ""
		}

		// Get the Title and Content value
		title := r.FormValue("Title")
		content := r.FormValue("Content")

		// Get the selected checkboxes
		checkboxes := r.Form["checkbox"]
		// Now you can use the retrieved id and username variables as needed
		InsertPost(user.Id, user.Username, user.ProfileImg, title, content, imgPath, strings.Join(checkboxes, ","), 0, 0)
		http.Redirect(w, r, "/Profile", http.StatusFound)
	}
}

func InsertPost(userID int, username string, userImg string, title string, content string, image string, category string, like int, dislike int) {
	query := `
        INSERT INTO posts (user_id, username,userImg,title,content, image,category, like, dislike)
        VALUES (?, ?, ?, ?, ?,?,?,?,?)
    `

	_, err := Postsdb.Exec(query, userID, username, userImg, title, content, image, category, like, dislike)
	if err != nil {
		log.Fatal(err)
	}
}
