package blog

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

type Post struct {
	Id       int64     `json:"id"`
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Date     time.Time `json:"date"`
	Modified time.Time `json:"modified"`
	Body     string    `json:"body"`
}

const (
	regexId        = `^[0-9]+$`
	regexTitle_Neg = `^[ ]*$`
	regexUsername  = `^[0-9a-zA-Z]{3,10}$`
)

type PostStatistics struct {
	PostId int64 `json:"postid"`
	Star1  int64 `json:"star1"`
	Star2  int64 `json:"star2"`
	Star3  int64 `json:"star3"`
	Star4  int64 `json:"star4"`
	Star5  int64 `json:"star5"`
}

type PostInfo struct {
	Post
	PostStatistics
}

func loadPost(id int64) (*Post, error) {

	var p Post
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `select * from post where id = ?`
	row := db.QueryRowContext(ctx, q, id)
	err := row.Scan(&p.Id, &p.Title, &p.Author, &p.Date, &p.Modified, &p.Body)

	return &p, err
}

func loadPostStatistics(id int64) (*PostStatistics, error) {

	var s PostStatistics
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `select * from poststatistics where postid = ?`
	row := db.QueryRowContext(ctx, q, id)
	err := row.Scan(&s.PostId, &s.Star1, &s.Star2, &s.Star3, &s.Star4, &s.Star5)

	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("loadPostStatistics id %d: no such id", id)
	case err != nil:
		return nil, fmt.Errorf("failed to scan rows of loadPost %d: %v", id, err)
	default:
		return &s, nil
	}
}

func (s *PostStatistics) getVote() (int, []float64) {

	percent := []float64{0, 0, 0, 0, 0}

	count := s.Star1 + s.Star2 + s.Star3 + s.Star4 + s.Star5
	sum := s.Star1 + 2*s.Star2 + 3*s.Star3 + 4*s.Star4 + 5*s.Star5
	if count == 0 {
		return 0, percent
	}

	v := float64(sum) / float64(count)
	vote := int(math.Round(v))

	percent[0] = float64(s.Star1) / float64(count)
	percent[1] = float64(s.Star2) / float64(count)
	percent[2] = float64(s.Star3) / float64(count)
	percent[3] = float64(s.Star4) / float64(count)
	percent[4] = float64(s.Star5) / float64(count)

	return vote, percent
}

func (p *Post) Validate() error {

	fail := func(what string, err error) error {
		return fmt.Errorf("fail to validate %s, %v", what, err)
	}

	if p == nil {
		return fail("reciever Post is nil", nil)
	}

	var match bool
	var err error

	id := strconv.FormatInt(p.Id, 10)
	match, err = regexp.MatchString(regexId, id)
	if !match || err != nil {
		return fail("Post.Id:"+id, err)
	}

	match, err = regexp.MatchString(regexTitle_Neg, p.Title)
	if match || err != nil {
		return fail("Post.Title:"+p.Title, err)
	}

	match, err = regexp.MatchString(regexUsername, p.Author)
	if !match || err != nil {
		return fail("Post.Author:"+p.Author, err)
	}

	return nil
}

// shall check if user exists in the db. otherwise, there maybe issues
// for example, the login user is removed from db, but login session
// is valid in redis, then the user sends a "create page" command,
// this invalid user creates the page successfully!
func (p *Post) save() error {

	fail := func(err error) error {
		return fmt.Errorf("save page failed: %v", err)
	}

	if err := p.Validate(); err != nil {
		return fail(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var now = time.Now()

	if p.Id == 0 {
		q := "INSERT INTO post (title, author, ctime, mtime, body) VALUES (?, ?, ?, ?, ?)"
		result, err := db.ExecContext(ctx, q, p.Title, p.Author, now, now, p.Body)
		if err != nil {
			return fail(err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fail(err)
		}
		p.Id = id

	} else {
		q := "UPDATE post set title = ?, body = ?, mtime = ? where id = ?"
		_, err := db.ExecContext(ctx, q, p.Title, p.Body, now, p.Id)
		if err != nil {
			return fail(err)
		}
	}

	return nil
}

func DeletePost(id int64) error {

	q := `DELETE FROM post WHERE id = ?`
	_, err := db.Exec(q, id)
	return err
}

func getPostInfo(id int64) (PostInfo, error) {

	var p PostInfo
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	q := `SELECT post.*, ` +
		`IFNULL(poststatistics.postid,0), ` +
		`IFNULL(poststatistics.star1,0), ` +
		`IFNULL(poststatistics.star2,0), ` +
		`IFNULL(poststatistics.star3,0), ` +
		`IFNULL(poststatistics.star4,0), ` +
		`IFNULL(poststatistics.star5,0)  ` +
		`FROM post ` +
		`LEFT JOIN poststatistics ` +
		`ON post.id = poststatistics.postid ` +
		`WHERE post.id = ?`

	row := db.QueryRowContext(ctx, q, id)
	err := row.Scan(&p.Id, &p.Title, &p.Author, &p.Date, &p.Modified, &p.Body,
		&p.PostId, &p.Star1, &p.Star2, &p.Star3, &p.Star4, &p.Star5)

	return p, err
}

func getPostsInfo() ([]PostInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var ps []PostInfo
	q := `SELECT post.*, ` +
		`IFNULL(poststatistics.postid,0), ` +
		`IFNULL(poststatistics.star1,0), ` +
		`IFNULL(poststatistics.star2,0), ` +
		`IFNULL(poststatistics.star3,0), ` +
		`IFNULL(poststatistics.star4,0), ` +
		`IFNULL(poststatistics.star5,0)  ` +
		`FROM post ` +
		`LEFT JOIN poststatistics ` +
		`ON post.id = poststatistics.postid`

	rows, err := db.QueryContext(ctx, q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var p PostInfo
		if err := rows.Scan(&p.Id, &p.Title, &p.Author, &p.Date, &p.Modified, &p.Body,
			&p.PostId, &p.Star1, &p.Star2, &p.Star3, &p.Star4, &p.Star5); err != nil {
			return nil, err
		}

		ps = append(ps, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}

func getAuthorsInfo() ([]string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var info []string
	q := `SELECT DISTINCT author FROM post ORDER BY author`

	rows, err := db.QueryContext(ctx, q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}

		info = append(info, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return info, nil
}
