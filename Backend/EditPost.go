package forum

import (
	"database/sql"
	"errors"
	"forum/Error"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type EditPostTemplateData struct {
	Username         string
	ProfileImg       string
	PostId           string
	Title            string
	Content          string
	PostImage        string
	Categories       []string
	NumNotifications int
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/EditPost" || guest {
		Error.RenderErrorPage(w, 404, "Page Not Found")
		return
	}

	var user Account
	user = GetUserData(w, r)
	if r.Method == "POST" {

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
		defaultImage := r.FormValue("defaultImage")

		// Get the selected checkboxes
		checkboxes := r.Form["checkbox"]

		postID := r.URL.Query().Get("Id")
		Id, err := strconv.Atoi(postID)
		if err != nil {
			log.Fatal("Invalid post ID:", err)
		}

		UpdatePost(Id, title, content, imgPath, strings.Join(checkboxes, ","), defaultImage)
		http.Redirect(w, r, "/Profile", http.StatusFound)
	} else {

		postID := r.URL.Query().Get("Id")
		err := CheckValidity(postID, user.Id)
		if err != nil {
			Error.RenderErrorPage(w, http.StatusForbidden, err.Error())
			return
		}

		title, content, image, categories := GetPostData(postID)

		numNotifications := GetNumOfNotifications(user.Id)

		var tmpl = template.Must(template.ParseFiles("./Pages/EditPost.html", "./Pages/nav.html"))

		data := EditPostTemplateData{
			Username:         user.Username,
			ProfileImg:       user.ProfileImg,
			PostId:           postID,
			Title:            title,
			Content:          content,
			PostImage:        image,
			Categories:       categories,
			NumNotifications: numNotifications,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			Error.RenderErrorPage(w, http.StatusInternalServerError, "Error executing template")
			return
		}
	}
}

func GetPostData(ID string) (string, string, string, []string) {
	var title, content, image, category string
	Id, err := strconv.Atoi(ID)
	if err != nil {
		log.Fatal("Invalid post ID:", err)
	}

	row := Postsdb.QueryRow("SELECT title, content, image, category FROM posts WHERE post_id = ?", Id)
	err = row.Scan(&title, &content, &image, &category)
	if err != nil {
		log.Fatal("Failed to query database:", err)
	}

	categories := strings.Split(category, ",")

	return title, content, image, categories
}

func UpdatePost(postID int, title string, content string, image string, category string, DefImg string) {
	if DefImg == "Def" {
		query := `
		UPDATE posts
		SET title = ?, content = ?, category = ?
		WHERE post_id = ?`
		_, err := Postsdb.Exec(query, title, content, category, postID)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		query := `
		UPDATE posts
		SET title = ?, content = ?, image = ?, category = ?
		WHERE post_id = ?`
		_, err := Postsdb.Exec(query, title, content, image, category, postID)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func CheckValidity(PostId string, UserId int) error {
	PostIdInt, err := strconv.Atoi(PostId)
	if err != nil {
		return errors.New("Invalid post ID: " + err.Error())
	}

	var creatorUserID int
	err = Postsdb.QueryRow("SELECT user_id FROM posts WHERE post_id = ?", PostIdInt).Scan(&creatorUserID)
	if err == sql.ErrNoRows {
		return errors.New("Post not found!")
	} else if err != nil {
		return errors.New("Error executing database query: " + err.Error())
	}

	if creatorUserID != UserId {
		return errors.New("You are only allowed to edit your posts!")
	}

	return nil
}
