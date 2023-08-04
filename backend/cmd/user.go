package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"01.kood.tech/git/aaaspoll/social-network/backend/models"
)

func (app *application) GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	currentUserId := r.URL.Query().Get("currentUserId")
	jsonData, err := json.Marshal(getUserInfo(userId, currentUserId, app.db))

	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func getUserInfo(userId, currentUserId string, db *sql.DB) models.User {
	var user models.User
	row := db.QueryRow("SELECT firstName, lastName, email, nickname, dateOfBirth, profilePic, aboutMe, public, online FROM users WHERE userId = ?", userId)
	err := row.Scan(&user.FirstName, &user.LastName, &user.Email, &user.Nickname, &user.DateOfBirth, &user.ProfilePic, &user.AboutMe, &user.Public, &user.Online)
	if err != nil {
		log.Println(err)
	}

	user.CurrentUserFollowStatus = getCurrentUserFollowStatus(userId, currentUserId, db)
	user.FollowsCurrentUser = getViewedUserFollowStatus(userId, currentUserId, db)
	user.Followers = getUserFollowers(userId, currentUserId, db)
	user.Following = getUserFollowing(userId, currentUserId, db)

	return user
}

func getCurrentUserFollowStatus(userId, currentUserId string, db *sql.DB) string {
	var isRequest bool
	// checks if current user is following viewed user
	err := db.QueryRow("SELECT isRequest FROM followers WHERE userId = ? AND followerId = ?", userId, currentUserId).Scan(&isRequest)
	if err != nil {
		if err == sql.ErrNoRows {
			return "Follow"
		} else {
			log.Println(err)
		}
	} else {
		if isRequest {
			return "Requested"
		}
	}

	return "Following"
}

func getViewedUserFollowStatus(userId, currentUserId string, db *sql.DB) bool {
	var isRequest bool
	// checks if viewed user is following current user
	err := db.QueryRow("SELECT isRequest FROM followers WHERE userId = ? AND followerId = ?", currentUserId, userId).Scan(&isRequest)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Println(err)
		}
	} else {
		if isRequest {
			return false
		}
	}

	return true
}

func getUserFollowers(userId, currentUserId string, db *sql.DB) []models.User {
	var followers = []models.User{}

	query := `
		SELECT u.userId, u.firstName, u.lastName, u.profilePic, u.public
		FROM users AS u
		INNER JOIN followers AS f ON u.userId = f.followerId
		WHERE f.userId = ? 
	`

	rows, err := db.Query(query, userId)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var follower models.User
		err := rows.Scan(&follower.UserId, &follower.FirstName, &follower.LastName, &follower.ProfilePic, &follower.Public)
		if err != nil {
			log.Println(err)
		}

		follower.CurrentUserFollowStatus = getCurrentUserFollowStatus(strconv.Itoa(follower.UserId), currentUserId, db)

		followers = append(followers, follower)
	}

	return followers
}

func getUserFollowing(userId, currentUserId string, db *sql.DB) []models.User {
	var following = []models.User{}

	query := `
		SELECT u.userId, u.firstName, u.lastName, u.profilePic, u.public
		FROM users AS u
		INNER JOIN followers AS f ON u.userId = f.userId
		WHERE f.followerId = ?
	`

	rows, err := db.Query(query, userId)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var followingUser models.User
		err := rows.Scan(&followingUser.UserId, &followingUser.FirstName, &followingUser.LastName, &followingUser.ProfilePic, &followingUser.Public)
		if err != nil {
			log.Println(err)
		}

		followingUser.CurrentUserFollowStatus = getCurrentUserFollowStatus(strconv.Itoa(followingUser.UserId), currentUserId, db)

		following = append(following, followingUser)
	}

	return following
}