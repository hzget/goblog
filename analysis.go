package blog

import (
	"database/sql"
	"fmt"
	"net/http"
)

func analysisHandler(w http.ResponseWriter, r *http.Request) {
	username, status := ValidateSession(w, r)
	//fmt.Println(username, status)
	switch status {
	case SessionUnauthorized:
		http.Error(w, "please log in first", http.StatusUnauthorized)
		return
	case SessionInternalError:
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userinfo, err := getUserInfo(username)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("loadRank user %s: no such user\n", username)
	case err != nil:
		fmt.Printf("failed to scan rows of loadRank %s: %v\n", username, err)
	default:
	}

	if err != nil || userinfo.Rank == "bronze" {
		http.Error(w, "please buy the analysis service befor using it", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "the analysis service is developping now")
}
