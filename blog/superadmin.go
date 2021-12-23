package blog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserInfo struct {
	Username string `json:"username"`
	Rank     string `json:"rank"`
}

func getRankInt(rank string) int64 {
	switch rank {
	case "bronze":
		return 0
	case "silver":
		return 1
	case "gold":
		return 2
	default:
		return -1
	}
}

func getUserInfo(username string) (*UserInfo, error) {

	info := &UserInfo{Username: username}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `select rank from users where username = ?`
	err := db.QueryRowContext(ctx, q, username).Scan(&info.Rank)

	return info, err
}

func getUsersInfo() ([]UserInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var s []UserInfo
	q := `select username, rank from users`

	rows, err := db.QueryContext(ctx, q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var info UserInfo
		if err := rows.Scan(&info.Username, &info.Rank); err != nil {
			return nil, err
		}

		s = append(s, info)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return s, nil
}

func superadminHandler(w http.ResponseWriter, r *http.Request) {

	data, err := getUsersInfo()
	if err != nil {
		fmt.Fprintf(w, "load Page info failed: %v", err)
		return
	}

	renderTemplate(w, "useradmin.html", data)
}

func saveranksHandler(w http.ResponseWriter, r *http.Request) {

	var data = struct {
		Pairs []UserInfo
	}{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request %v", err), http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	fail := func(err error) {
		http.Error(w, "internal error: failed to save", http.StatusInternalServerError)
		fmt.Printf("internal error %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	// Get a Tx for making transaction requests.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fail(err)
		return
	}
	defer tx.Rollback()

	q := `UPDATE users SET rank = ? WHERE username = ?`
	for _, info := range data.Pairs {
		_, err = tx.ExecContext(ctx, q, info.Rank, info.Username)
		if err != nil {
			fail(err)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		fail(err)
		return
	}

	http.Redirect(w, r, "./superadmin", http.StatusSeeOther)
}

func makeAdminHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, status := ValidateSession(w, r)
		switch status {
		case SessionUnauthorized:
			http.Error(w, "please log in first", http.StatusUnauthorized)
			return
		case SessionInternalError:
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if !IsAdmin(username) {
			http.Error(w, "only the admin can make these operations", http.StatusBadRequest)
			return
		}

		fn(w, r)
	}
}

func IsAdmin(username string) bool {
	if username == "superadmin" || username == "admin" {
		return true
	}

	return false
}
