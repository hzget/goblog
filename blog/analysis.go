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

const (
	ByAuthor = 1
	ByPostId = 2
)

type AnalyzeReq struct {
	How    AnalyzeHow `json:"how"`
	Author string     `json:"author"`
	PostId int64      `json:"id"`
}

func analysisHandler(w http.ResponseWriter, r *http.Request) {
	username, err := ValidateSession(w, r)
	if err != nil {
		RespondAlert(w, err)
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
		printAlert(w, "please buy the analysis service before using it", http.StatusBadRequest)
		return
	}

	data, err := getAuthorsInfo()
	if err != nil {
		printAlert(w, fmt.Sprintf("get Authors info failed: %v", err),
			http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "analysis.html", data)
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := ValidateSession(w, r)
	if err != nil {
		RespondError(w, err)
		return
	}

	req := &AnalyzeReq{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(req); err != nil {
		//fmt.Println(err)
		http.Error(w, encodeJsonResp(false,
			fmt.Sprintf("failed to decode request %v", err)),
			http.StatusBadRequest)
		return
	}

	var result string
	switch req.How {
	case ByAuthor:
		analyzeAuthorHandler(w, r, user, req)
		return
	case ByPostId:
		result, err = analyzePostHandler(w, r, req)
	default:
		err = fmt.Errorf("req.How(%d) illegal", req.How)
		http.Error(w, encodeJsonResp(false, err.Error()),
			http.StatusBadRequest)
		return
	}

	if err == sql.ErrNoRows {
		http.Error(w, encodeJsonResp(false, err.Error()),
			http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, encodeJsonResp(false, err.Error()),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, encodeJsonResp(true, result))
	return
}

func analyzeAuthorHandler(w http.ResponseWriter, r *http.Request, user string, a *AnalyzeReq) {

	userinfo, err := getUserInfo(user)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, fmt.Sprintf("loadRank user %s: no such user\n", user), http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, fmt.Sprintf("failed to scan rows of loadRank %s: %v\n", user, err), http.StatusBadRequest)
		return
	default:
	}

	authorinfo, err := getUserInfo(a.Author)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, fmt.Sprintf("loadRank user %s: no such author\n", a.Author), http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, fmt.Sprintf("failed to scan rows of loadRank %s: %v\n", a.Author, err), http.StatusBadRequest)
		return
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

func analyzePostHandler(w http.ResponseWriter, r *http.Request, req *AnalyzeReq) (string, error) {

	data, err := loadPost(req.PostId)
	if err != nil {
		return "", err
	}
	return AnalyzePost(data.Body)
}

func AnalyzePost(text string) (string, error) {

	conn, err := grpc.Dial(dataAnalysisAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return "", fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDataAnalysisClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.AnalyzePost(ctx, &pb.Text{Text: text})
	if err != nil {
		return "", fmt.Errorf("could not analyze: %v", err)
	}

	//log.Printf("Analysis result: %s\n", r.GetResult())
	Debug(fmt.Sprintf("Analysis result: %s\n", r.GetResult()))

	return r.GetResult(), nil
}
