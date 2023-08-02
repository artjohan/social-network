package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"01.kood.tech/git/aaaspoll/social-network/backend/models"
)

func (app *application) PostsHandler(w http.ResponseWriter, r *http.Request) {
	var posts []models.Post

	stmt := `SELECT postId, userId, content, COALESCE(img, ""), likes, privacy, created FROM posts ORDER BY created ASC LIMIT 200`
	rows, err := app.db.Query(stmt)
	defer rows.Close()

	for rows.Next() {
		post := models.Post{}
		var img sql.NullString
		err = rows.Scan(&post.PostID, &post.UserID, &post.Content, &img, &post.Likes, &post.Privacy, &post.CreatedAt)
		if err != nil {
			log.Fatalf("Err: %s", err)
		}
		if img.Valid {
			post.Img = img.String
		} else {
			post.Img = ""
		}
		posts = append(posts, post)

	}
	fmt.Println(posts)
	jsonResp, err := json.Marshal(posts)
	if err != nil {
		log.Fatalf("Err: %s", err)
	}

	w.Write(jsonResp)
	return
}

func (app *application) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB max memory

	file, handler, err := r.FormFile("img")
	if err != nil {
		log.Printf("Error Retrieving the File")
		log.Println(err)
		return
	}
	defer file.Close()

	imgDir := "backend/media/posts"

	newFile, err := os.Create(imgDir + "/" + handler.Filename)
	if err != nil {
		log.Println(err)
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		log.Println(err)
		return
	}

	imgPath := imgDir + "/" + handler.Filename

	var post models.Post
	post.Img = imgPath
	userIDStr := r.FormValue("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Error converting userId to int: %s", err)
		return
	}

	post.Content = r.FormValue("content")
	post.UserID = userID

	fmt.Println(post.Img)

	stmt := `INSERT INTO posts (userId, content, img, created) VALUES (?, ?, ?, current_date)`
	_, err = app.db.Exec(stmt, post.UserID, post.Content, post.Img)

	if err != nil {
		log.Printf("Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	return
}
