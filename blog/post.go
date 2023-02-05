package blog

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
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

var dbParallelN int64

const dbParallelNLimit = 150

type PostStatistics struct {
	PostId int64    `json:"postid"`
	Star   [5]int64 `json:"star"`
}

type PostInfo struct {
	Post
	Star [5]int64 `json:"star"`
}

func loadPost(id int64) (*Post, error) {

	var p Post
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	key := Key_SQL_loadPost + fmt.Sprintf("%d", id)
	if err := DBGetCache(key, &p); err == nil {
		return &p, nil
	}

	count := atomic.AddInt64(&dbParallelN, 1)
	defer func() {
		atomic.AddInt64(&dbParallelN, -1)
	}()

	if count > dbParallelNLimit {
		s := fmt.Sprintf("parallel dbaccess reach limit %d,"+
			" drop request", dbParallelNLimit)
		return nil, &limitErr{nil, s}
	}

	q := Key_SQL_loadPost + `?`
	row := db.QueryRowContext(ctx, q, id)
	err := row.Scan(&p.Id, &p.Title, &p.Author, &p.Date, &p.Modified, &p.Body)

	if err != nil {
		Info("logPost:" + err.Error())
		return nil, err
	}

	DBUpdateCache(key, &p)

	return &p, err
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
		id := fmt.Sprintf("%d", p.Id)
		DBRemoveCache(Key_SQL_GetPostInfo + id)
		DBRemoveCache(Key_SQL_loadPost + id)
	}

	return nil
}

func DeletePost(id int64) error {

	q := `DELETE FROM post WHERE id = ?`
	_, err := db.Exec(q, id)
	s := fmt.Sprintf("%d", id)
	DBRemoveCache(Key_SQL_GetPostInfo + s)
	DBRemoveCache(Key_SQL_loadPost + s)
	return err
}

func getPostInfo(id int64) (PostInfo, error) {

	var p PostInfo
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	key := Key_SQL_GetPostInfo + fmt.Sprintf("%d", id)
	if err := DBGetCache(key, &p); err == nil {
		return p, nil
	}

	count := atomic.AddInt64(&dbParallelN, 1)
	defer func() {
		atomic.AddInt64(&dbParallelN, -1)
	}()

	if count > dbParallelNLimit {
		s := fmt.Sprintf("parallel dbaccess reach limit %d,"+
			" drop request", dbParallelNLimit)
		return p, &limitErr{nil, s}
	}

	row := db.QueryRowContext(ctx, Key_SQL_GetPostInfo+"?", id)
	err := row.Scan(&p.Id, &p.Title, &p.Author, &p.Date, &p.Modified, &p.Body,
		&p.Star[0], &p.Star[1], &p.Star[2], &p.Star[3], &p.Star[4])

	if err != nil {
		Info("getPostInfo:" + err.Error())
		return p, err
	}

	DBUpdateCache(key, &p)

	return p, err
}

func getPostsInfo() ([]PostInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	var ps []PostInfo
	q := `SELECT post.*, ` +
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
			&p.Star[0], &p.Star[1], &p.Star[2], &p.Star[3], &p.Star[4]); err != nil {
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
