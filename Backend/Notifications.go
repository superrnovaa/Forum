package forum

import (
	"database/sql"
	"fmt"
	"forum/Error"
	"html/template"
	"log"
	"net/http"
)

type NotificationsTemplateData struct {
	Username         string
	ProfileImg       string
	Notifications    []Notification
	NumNotifications int
}

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/Notifications" {
		Error.RenderErrorPage(w, 404, "Page Not Found")
		return
	}

	var user Account
	user = GetUserData(w, r)

	if r.Method == "GET" {
		notifications, err := fetchNotifications(user.Id)
		if err != nil {
			fmt.Println(err)
		}

		_, err = Notificationsdb.Exec(`UPDATE newNotifications SET num_notifications = 0 WHERE user_id = ?`, user.Id)
		if err != nil {
			log.Fatal(err)
		}

		data := NotificationsTemplateData{
			Username:         user.Username,
			ProfileImg:       user.ProfileImg,
			Notifications:    notifications,
			NumNotifications: 0,
		}

		var tmpl = template.Must(template.ParseFiles("./Pages/Notifications.html", "./Pages/nav.html"))

		err = tmpl.Execute(w, data)

		if err != nil {
			Error.RenderErrorPage(w, http.StatusInternalServerError, "Error executing template")
			return
		}
	}
}

func fetchNotifications(UserID int) ([]Notification, error) {
	var rows *sql.Rows
	rows, err := Notificationsdb.Query("SELECT follower_user_name, follower_profile, action, post_image, comment FROM notifications WHERE creator_user_id = ?", UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err)
	}
	defer rows.Close()

	notifications := make([]Notification, 0)
	for rows.Next() {
		var notification Notification

		err := rows.Scan(&notification.FollowerUserName, &notification.FollowerProfile, &notification.Action ,&notification.PostImage, &notification.Comment)
		if err != nil { 
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	reverseNotifications(notifications)

	return notifications, nil
}

func InsertNotification(FollowerUserName string, FollowerProfile string, PostId int, Action string, Comment string) {
	//Excuted When Liked, disliked, commented
	var CreatorUserId string
	err := Postsdb.QueryRow("SELECT user_id FROM posts WHERE post_id = ?", PostId).Scan(&CreatorUserId)
	if err != nil {
		log.Fatal(err)
	}


	var PostTitle ,PostImage string
	err = Postsdb.QueryRow("SELECT title, image FROM posts WHERE post_id = ?", PostId).Scan(&PostTitle, &PostImage)
	if err != nil {
		log.Fatal(err)
	}

	if Action == "Like" {
		Action = "Liked your post" + " '" + PostTitle + "'"
	} else if Action == "Dislike" {
		Action = "Disliked your post" + "  '" + PostTitle + "'"
	} else if Action == "Comment" {
		Action = "Commented on your post" + " '" + PostTitle + "' :"
	}

	_, err = Notificationsdb.Exec(`INSERT INTO notifications (creator_user_id, follower_user_name, follower_profile, action, post_image, comment ) VALUES (?, ?,?,?,?,?)`, CreatorUserId, FollowerUserName, FollowerProfile, Action,PostImage, Comment)
	if err != nil {
		log.Fatal(err)
	}

	_, err = Notificationsdb.Exec(`UPDATE newNotifications SET num_notifications = num_notifications + 1 WHERE user_id = ?`, CreatorUserId)
	if err != nil {
		log.Fatal(err)
	}

}

func reverseNotifications(arr []Notification) {
	length := len(arr)
	for i := 0; i < length/2; i++ {
		arr[i], arr[length-i-1] = arr[length-i-1], arr[i]
	}
}

func GetNumOfNotifications(userId int) int {
	var numNotifications int
	err := Notificationsdb.QueryRow("SELECT num_notifications FROM newNotifications WHERE user_id = ?", userId).Scan(&numNotifications)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			_, err = Notificationsdb.Exec(`INSERT INTO newNotifications (user_id, num_notifications) VALUES (?, 0) `, userId)
			if err != nil {
				log.Fatal(err)
			}
			return 0
		} else {
			log.Fatal(err)
		}
	}
	return numNotifications
}
