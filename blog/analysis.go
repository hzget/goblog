package blog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	pb "github.com/hzget/analysisdriver"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"time"
)

type AnalyzeHow int64

type AnalyzeReq struct {
	How    AnalyzeHow `json:"how"`
	Author string     `json:"author"`
}

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
		http.Error(w, "please buy the analysis service before using it", http.StatusUnauthorized)
		return
	}

	data, err := getAuthorsInfo()
	if err != nil {
		fmt.Fprintf(w, "get Authors info failed: %v", err)
		return
	}
	renderTemplate(w, "analysis.html", data)
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	user, status := ValidateSession(w, r)
	switch status {
	case SessionUnauthorized:
		http.Error(w, "please log in first", http.StatusUnauthorized)
		return
	case SessionInternalError:
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	a := &AnalyzeReq{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(a); err != nil {
		fmt.Println(err)
		http.Error(w, fmt.Sprintf("failed to decode request %v", err), http.StatusBadRequest)
		return
	}

	userinfo, err := getUserInfo(user)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("loadRank user %s: no such user\n", user)
	case err != nil:
		fmt.Printf("failed to scan rows of loadRank %s: %v\n", user, err)
	default:
	}

	authorinfo, err := getUserInfo(a.Author)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("loadRank user %s: no such author\n", a.Author)
	case err != nil:
		fmt.Printf("failed to scan rows of loadRank %s: %v\n", a.Author, err)
	default:
	}

	if getRankInt(userinfo.Rank) < getRankInt(authorinfo.Rank) {
		fmt.Fprintf(w, "The %s user %s want to analyze %s author %s's articles."+
			"This analysis is not allowed\n",
			userinfo.Rank, user, authorinfo.Rank, a.Author)
		return
	}

	score := AnalyzeByAuthor(a.Author)

	fmt.Fprintf(w, "result %d", score)
}

func AnalyzeByAuthor(name string) int32 {

	conn, err := grpc.Dial(dataAnalysisAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDataAnalysisClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.AnalyzeByAuthor(ctx, &pb.Author{Name: name})
	if err != nil {
		log.Fatalf("could not analyze: %v", err)
	}

	log.Printf("Analysis result: %d\n", r.GetScore())

	return r.GetScore()
}
